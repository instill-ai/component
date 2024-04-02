package airbyte

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	dockerclient "github.com/docker/docker/client"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.Connector
	dockerClient *dockerclient.Client
	cache        *bigcache.BigCache
	options      ConnectorOptions
}

type ConnectorOptions struct {
	MountSourceVDP        string
	MountTargetVDP        string
	MountSourceAirbyte    string
	MountTargetAirbyte    string
	ExcludeLocalConnector bool
}

type Execution struct {
	base.Execution
	connector *Connector
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {

		dockerClient, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			logger.Error(err.Error())
		}
		// defer dockerClient.Close()
		cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(60 * time.Minute))
		if err != nil {
			logger.Error(err.Error())
		}

		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
			dockerClient: dockerClient,
			cache:        cache,
			options:      options,
		}

		err = connector.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())

		}

		if options.ExcludeLocalConnector {
			def, _ := connector.GetConnectorDefinitionByID("airbyte-destination-local-json", nil, nil)
			(*def).Tombstone = true
			def, _ = connector.GetConnectorDefinitionByID("airbyte-destination-csv", nil, nil)
			(*def).Tombstone = true
			def, _ = connector.GetConnectorDefinitionByID("airbyte-destination-sqlite", nil, nil)
			(*def).Tombstone = true
			def, _ = connector.GetConnectorDefinitionByID("airbyte-destination-duckdb", nil, nil)
			(*def).Tombstone = true
		}

		InitAirbyteCatalog(logger)

	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	e.connector = c
	return e, nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	// Create ConfiguredAirbyteCatalog
	cfgAbCatalog := ConfiguredAirbyteCatalog{
		Streams: []ConfiguredAirbyteStream{
			{
				Stream:              &TaskOutputAirbyteCatalog.Streams[0],
				SyncMode:            "full_refresh", // TODO: config
				DestinationSyncMode: "append",       // TODO: config
			},
		},
	}

	byteCfgAbCatalog, err := json.Marshal(&cfgAbCatalog)
	if err != nil {
		return nil, fmt.Errorf("marshal AirbyteMessage error: %w", err)
	}

	// Create AirbyteMessage RECORD type, i.e., AirbyteRecordMessage in JSON Line format
	var byteAbMsgs []byte

	for idx, input := range inputs {

		b, err := protojson.MarshalOptions{
			UseProtoNames: true,
		}.Marshal(input.Fields["data"].GetStructValue())
		if err != nil {
			return nil, fmt.Errorf("data [%d] error: %w", idx, err)
		}
		abMsg := AirbyteMessage{}
		abMsg.Type = "RECORD"
		abMsg.Record = &AirbyteRecordMessage{
			Stream:    TaskOutputAirbyteCatalog.Streams[0].Name,
			Data:      b,
			EmittedAt: time.Now().UnixMilli(),
		}
		b, err = json.Marshal(&abMsg)
		if err != nil {
			return nil, fmt.Errorf("marshal AirbyteMessage error: %w", err)
		}
		b = []byte(string(b) + "\n")
		byteAbMsgs = append(byteAbMsgs, b...)
	}

	// Remove the last "\n"
	byteAbMsgs = byteAbMsgs[:len(byteAbMsgs)-1]

	connDef, err := e.connector.GetConnectorDefinitionByUID(e.UID, nil, nil)
	if err != nil {
		return nil, err
	}
	imageName := connDef.VendorAttributes.GetFields()[e.Config.GetFields()["destination"].GetStringValue()].GetStringValue()
	containerName := fmt.Sprintf("%s.%d.write", e.UID, time.Now().UnixNano())
	configFileName := fmt.Sprintf("%s.%d.write", e.UID, time.Now().UnixNano())
	catalogFileName := fmt.Sprintf("%s.%d.write", e.UID, time.Now().UnixNano())

	// If there is already a container run dispatched in the previous attempt, return exitCodeOK directly
	if _, err := e.connector.cache.Get(containerName); err == nil {
		return nil, nil
	}

	// Write config into a container local file (always overwrite)
	configFilePath := fmt.Sprintf("%s/connector-data/config/%s.json", e.connector.options.MountTargetVDP, configFileName)
	if err := os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to create folders for filepath %s", configFilePath), "WriteContainerLocalFileError", err)
	}

	configuration := func() []byte {
		if e.Config != nil {
			b, err := e.Config.MarshalJSON()
			if err != nil {
				e.Logger.Error(err.Error())
			}
			return b
		}
		return []byte{}
	}()
	if err := os.WriteFile(configFilePath, configuration, 0644); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to write connector config file %s", configFilePath), "WriteContainerLocalFileError", err)
	}

	// Write catalog into a container local file (always overwrite)
	catalogFilePath := fmt.Sprintf("%s/connector-data/catalog/%s.json", e.connector.options.MountTargetVDP, catalogFileName)
	if err := os.MkdirAll(filepath.Dir(catalogFilePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to create folders for filepath %s", catalogFilePath), "WriteContainerLocalFileError", err)
	}
	if err := os.WriteFile(catalogFilePath, byteCfgAbCatalog, 0644); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to write connector catalog file %s", catalogFilePath), "WriteContainerLocalFileError", err)
	}

	defer func() {
		// Delete config local file
		if _, err := os.Stat(configFilePath); err == nil {
			if err := os.Remove(configFilePath); err != nil {
				e.Logger.Error(fmt.Sprintln("Activity", "ImageName", imageName, "ContainerName", containerName, "Error", err))
			}
		}

		// Delete catalog local file
		if _, err := os.Stat(catalogFilePath); err == nil {
			if err := os.Remove(catalogFilePath); err != nil {
				e.Logger.Error(fmt.Sprintln("Activity", "ImageName", imageName, "ContainerName", containerName, "Error", err))
			}
		}
	}()

	out, err := e.connector.dockerClient.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err := io.Copy(os.Stdout, out); err != nil {
		return nil, err
	}

	resp, err := e.connector.dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			Image:        imageName,
			AttachStdin:  true,
			AttachStdout: true,
			OpenStdin:    true,
			StdinOnce:    true,
			Tty:          true,
			Cmd: []string{
				"write",
				"--config",
				configFilePath,
				"--catalog",
				catalogFilePath,
			},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type: func() mount.Type {
						if string(e.connector.options.MountSourceVDP[0]) == "/" {
							return mount.TypeBind
						}
						return mount.TypeVolume
					}(),
					Source: e.connector.options.MountSourceVDP,
					Target: e.connector.options.MountTargetVDP,
				},
				{
					Type: func() mount.Type {
						if string(e.connector.options.MountSourceVDP[0]) == "/" {
							return mount.TypeBind
						}
						return mount.TypeVolume
					}(),
					Source: e.connector.options.MountSourceAirbyte,
					Target: e.connector.options.MountTargetAirbyte,
				},
			},
		},
		nil, nil, containerName)
	if err != nil {
		return nil, err
	}

	hijackedResp, err := e.connector.dockerClient.ContainerAttach(context.Background(), resp.ID, container.AttachOptions{
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return nil, err
	}

	// need to append "\n" and "ctrl+D" at the end of the input message
	_, err = hijackedResp.Conn.Write(append(byteAbMsgs, 10, 4))
	if err != nil {
		return nil, err
	}

	if err := e.connector.dockerClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	var bufStdOut bytes.Buffer
	if _, err := bufStdOut.ReadFrom(hijackedResp.Reader); err != nil {
		return nil, err
	}

	if err := e.connector.dockerClient.ContainerRemove(context.Background(), resp.ID,
		container.RemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
		return nil, err
	}

	// Set cache flag (empty value is fine since we need only the entry record)
	if err := e.connector.cache.Set(containerName, []byte{}); err != nil {
		return nil, err
	}

	e.Logger.Info(fmt.Sprintln("Activity",
		"ImageName", imageName,
		"ContainerName", containerName,
		"STDOUT", bufStdOut.String()))

	// Delete the cache entry only after the write completed
	if err := e.connector.cache.Delete(containerName); err != nil {
		e.Logger.Error(err.Error())
	}

	outputs := []*structpb.Struct{}
	for range inputs {
		outputs = append(outputs, &structpb.Struct{})
	}

	return outputs, nil
}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {

	def, err := c.GetConnectorDefinitionByUID(defUID, nil, nil)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	imageName := def.VendorAttributes.GetFields()[config.GetFields()["destination"].GetStringValue()].GetStringValue()
	containerName := fmt.Sprintf("%s.%d.check", defUID, time.Now().UnixNano())
	configFilePath := fmt.Sprintf("%s/connector-data/config/%s.json", c.options.MountTargetVDP, containerName)

	// Write config into a container local file
	if err := os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm); err != nil {
		return pipelinePB.Connector_STATE_ERROR, fmt.Errorf(fmt.Sprintf("unable to create folders for filepath %s", configFilePath), "WriteContainerLocalFileError", err)
	}

	configuration := func() []byte {
		if config != nil {
			b, err := config.MarshalJSON()
			if err != nil {
				c.Logger.Error(err.Error())
			}
			return b
		}
		return []byte{}
	}()

	if err := os.WriteFile(configFilePath, configuration, 0644); err != nil {
		return pipelinePB.Connector_STATE_ERROR, fmt.Errorf(fmt.Sprintf("unable to write connector config file %s", configFilePath), "WriteContainerLocalFileError", err)
	}

	defer func() {
		// Delete config local file
		if _, err := os.Stat(configFilePath); err == nil {
			if err := os.Remove(configFilePath); err != nil {
				c.Logger.Error(fmt.Sprintf("ImageName: %s, ContainerName: %s, Error: %v", imageName, containerName, err))
			}
		}
	}()

	out, err := c.dockerClient.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	defer out.Close()

	if _, err := io.Copy(os.Stdout, out); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	resp, err := c.dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			Image: imageName,
			Tty:   false,
			Cmd: []string{
				"check",
				"--config",
				configFilePath,
			},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type: func() mount.Type {
						if string(c.options.MountSourceVDP[0]) == "/" {
							return mount.TypeBind
						}
						return mount.TypeVolume
					}(),
					Source: c.options.MountSourceVDP,
					Target: c.options.MountTargetVDP,
				},
			},
		},
		nil, nil, containerName)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	if err := c.dockerClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	statusCh, errCh := c.dockerClient.ContainerWait(context.Background(), resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return pipelinePB.Connector_STATE_ERROR, err
		}
	case <-statusCh:
	}

	if out, err = c.dockerClient.ContainerLogs(context.Background(),
		resp.ID,
		container.LogsOptions{
			ShowStdout: true,
		},
	); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	if err := c.dockerClient.ContainerRemove(context.Background(), containerName,
		container.RemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	var bufStdOut, bufStdErr bytes.Buffer
	if _, err := stdcopy.StdCopy(&bufStdOut, &bufStdErr, out); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	var teeStdOut io.Reader = strings.NewReader(bufStdOut.String())
	var teeStdErr io.Reader = strings.NewReader(bufStdErr.String())
	teeStdOut = io.TeeReader(teeStdOut, &bufStdOut)
	teeStdErr = io.TeeReader(teeStdErr, &bufStdErr)

	var byteStdOut, byteStdErr []byte
	if byteStdOut, err = io.ReadAll(teeStdOut); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	if byteStdErr, err = io.ReadAll(teeStdErr); err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	c.Logger.Info(fmt.Sprintf("ImageName, %s, ContainerName, %s, STDOUT, %v, STDERR, %v", imageName, containerName, byteStdOut, byteStdErr))

	scanner := bufio.NewScanner(&bufStdOut)
	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			return pipelinePB.Connector_STATE_ERROR, err
		}

		var jsonMsg map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &jsonMsg); err == nil {
			switch jsonMsg["type"] {
			case "CONNECTION_STATUS":
				switch jsonMsg["connectionStatus"].(map[string]interface{})["status"] {
				case "SUCCEEDED":
					return pipelinePB.Connector_STATE_CONNECTED, nil
				case "FAILED":
					return pipelinePB.Connector_STATE_ERROR, nil
				default:
					return pipelinePB.Connector_STATE_ERROR, fmt.Errorf("UNKNOWN STATUS")
				}
			}
		}
	}
	return pipelinePB.Connector_STATE_ERROR, nil
}
