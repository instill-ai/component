package numbers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	_ "embed"
	b64 "encoding/base64"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const urlRegisterAsset = "https://api.numbersprotocol.io/api/v3/assets/"
const urlUserMe = "https://api.numbersprotocol.io/api/v3/auth/users/me"

var once sync.Once
var connector base.IConnector

//go:embed config/definitions.json
var definitionsJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

type CommitCustomLicense struct {
	Name     *string `json:"name,omitempty"`
	Document *string `json:"document,omitempty"`
}
type CommitCustom struct {
	AssetCreator      *string              `json:"assetCreator,omitempty"`
	DigitalSourceType *string              `json:"digitalSourceType,omitempty"`
	MiningPreference  *string              `json:"miningPreference,omitempty"`
	GeneratedThrough  string               `json:"generatedThrough"`
	GeneratedBy       *string              `json:"generatedBy,omitempty"`
	CreatorWallet     *string              `json:"creatorWallet,omitempty"`
	License           *CommitCustomLicense `json:"license,omitempty"`
	Metadata          *struct {
		Pipeline *struct {
			UID    *string     `json:"uid,omitempty"`
			Recipe interface{} `json:"recipe,omitempty"`
		} `json:"pipeline,omitempty"`
		Owner *struct {
			UID *string `json:"uid,omitempty"`
		} `json:"owner,omitempty"`
	} `json:"instillMetadata,omitempty"`
}

type Meta struct {
	Proof struct {
		Hash      string `json:"hash"`
		MIMEType  string `json:"mimeType"`
		Timestamp string `json:"timestamp"`
	} `json:"proof"`
}

type Register struct {
	Caption         *string       `json:"caption,omitempty"`
	Headline        *string       `json:"headline,omitempty"`
	NITCommitCustom *CommitCustom `json:"nit_commit_custom,omitempty"`
	Meta
}

type Input struct {
	Images            []string `json:"images"`
	AssetCreator      *string  `json:"asset_creator,omitempty"`
	Caption           *string  `json:"caption,omitempty"`
	Headline          *string  `json:"headline,omitempty"`
	DigitalSourceType *string  `json:"digital_source_type,omitempty"`
	MiningPreference  *string  `json:"mining_preference,omitempty"`
	GeneratedBy       *string  `json:"generated_by,omitempty"`
	License           *struct {
		Name     *string `json:"name,omitempty"`
		Document *string `json:"document,omitempty"`
	} `json:"license,omitempty"`
	Metadata *struct {
		Pipeline *struct {
			UID    *string     `json:"uid,omitempty"`
			Recipe interface{} `json:"recipe,omitempty"`
		} `json:"pipeline,omitempty"`
		Owner *struct {
			UID *string `json:"uid,omitempty"`
		} `json:"owner,omitempty"`
	} `json:"metadata,omitempty"`
}

type Output struct {
	AssetUrls []string `json:"asset_urls"`
}

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {

		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}

	})
	return connector
}

func getToken(config *structpb.Struct) string {
	return fmt.Sprintf("token %s", config.GetFields()["capture_token"].GetStringValue())
}

func (e *Execution) registerAsset(data []byte, reg Register) (string, error) {

	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	var fw io.Writer
	var err error

	fileName, _ := uuid.NewV4()
	if fw, err = w.CreateFormFile("asset_file", fileName.String()+mimetype.Detect(data).Extension()); err != nil {
		return "", err
	}

	if _, err := io.Copy(fw, bytes.NewReader(data)); err != nil {
		return "", err
	}

	if reg.Caption != nil {
		_ = w.WriteField("caption", *reg.Caption)
	}

	if reg.Headline != nil {
		_ = w.WriteField("headline", *reg.Headline)
	}

	if reg.NITCommitCustom != nil {
		commitBytes, _ := json.Marshal(*reg.NITCommitCustom)
		_ = w.WriteField("nit_commit_custom", string(commitBytes))
	}
	metaBytes, _ := json.Marshal(Meta{})
	_ = w.WriteField("meta", string(metaBytes))

	w.Close()

	req, err := http.NewRequest("POST", urlRegisterAsset, &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", getToken(e.Config))

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return "", err
	}

	if res.StatusCode == http.StatusCreated {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		var jsonRes map[string]interface{}
		_ = json.Unmarshal(bodyBytes, &jsonRes)
		if cid, ok := jsonRes["cid"]; ok {
			return cid.(string), nil
		} else {
			return "", fmt.Errorf("register file failed")
		}

	} else {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf(string(bodyBytes))
	}
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	var outputs []*structpb.Struct

	for _, input := range inputs {

		assetUrls := []string{}

		inputStruct := Input{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			return nil, err
		}

		for _, image := range inputStruct.Images {
			imageBytes, err := b64.StdEncoding.DecodeString(base.TrimBase64Mime(image))
			if err != nil {
				return nil, err
			}

			var commitLicense *CommitCustomLicense
			if inputStruct.License != nil {
				commitLicense = &CommitCustomLicense{
					Name:     inputStruct.License.Name,
					Document: inputStruct.License.Document,
				}
			}

			commitCustom := CommitCustom{
				AssetCreator:      inputStruct.AssetCreator,
				DigitalSourceType: inputStruct.DigitalSourceType,
				MiningPreference:  inputStruct.MiningPreference,
				GeneratedThrough:  "https://instill.tech", //TODO: support Core Host
				GeneratedBy:       inputStruct.GeneratedBy,
				License:           commitLicense,
				Metadata:          inputStruct.Metadata,
			}

			reg := Register{
				Caption:         inputStruct.Caption,
				Headline:        inputStruct.Headline,
				NITCommitCustom: &commitCustom,
			}
			assetCid, err := e.registerAsset(imageBytes, reg)
			if err != nil {
				return nil, err
			}

			assetUrls = append(assetUrls, fmt.Sprintf("https://verify.numbersprotocol.io/asset-profile?nid=%s", assetCid))
		}

		outputStruct := Output{
			AssetUrls: assetUrls,
		}

		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)

	}

	return outputs, nil

}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {

	req, err := http.NewRequest("GET", urlUserMe, nil)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, nil
	}
	req.Header.Set("Authorization", getToken(config))

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, nil
	}
	if res.StatusCode == http.StatusOK {
		return pipelinePB.Connector_STATE_CONNECTED, nil
	}
	return pipelinePB.Connector_STATE_ERROR, nil
}
