//go:generate compogen readme --connector ./config ./README.mdx
package numbers

import (
	"bytes"
	"context"
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
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const urlRegisterAsset = "https://api.numbersprotocol.io/api/v3/assets/"
const urlUserMe = "https://api.numbersprotocol.io/api/v3/auth/users/me"

var once sync.Once
var con *connector

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

type connector struct {
	base.Connector
}

type execution struct {
	base.ConnectorExecution
}

type CommitCustomLicense struct {
	Name     *string `json:"name,omitempty"`
	Document *string `json:"document,omitempty"`
}
type CommitCustom struct {
	AssetCreator      *string               `json:"assetCreator,omitempty"`
	DigitalSourceType *string               `json:"digitalSourceType,omitempty"`
	MiningPreference  *string               `json:"miningPreference,omitempty"`
	GeneratedThrough  string                `json:"generatedThrough"`
	GeneratedBy       *string               `json:"generatedBy,omitempty"`
	CreatorWallet     *string               `json:"creatorWallet,omitempty"`
	License           *CommitCustomLicense  `json:"license,omitempty"`
	Metadata          *CommitCustomMetadata `json:"instillMetadata,omitempty"`
}

type CommitCustomMetadata struct {
	Pipeline struct {
		UID    string
		Recipe interface{}
	}
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
}

type Output struct {
	AssetUrls []string `json:"asset_urls"`
}

func Init(bc base.Connector) *connector {
	once.Do(func() {
		con = &connector{Connector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		ConnectorExecution: base.ConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
	}}, nil
}

func getToken(config *structpb.Struct) string {
	return fmt.Sprintf("token %s", config.GetFields()["capture_token"].GetStringValue())
}

func (e *execution) registerAsset(data []byte, reg Register) (string, error) {

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
	req.Header.Set("Authorization", getToken(e.Connection))

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

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

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

			meta := CommitCustomMetadata{
				Pipeline: struct {
					UID    string
					Recipe interface{}
				}{
					UID:    e.GetSystemVariables()["__PIPELINE_UID"].(string),
					Recipe: e.GetSystemVariables()["__PIPELINE_RECIPE"],
				},
			}
			commitCustom := CommitCustom{
				AssetCreator:      inputStruct.AssetCreator,
				DigitalSourceType: inputStruct.DigitalSourceType,
				MiningPreference:  inputStruct.MiningPreference,
				GeneratedThrough:  "https://instill.tech", //TODO: support Core Host
				GeneratedBy:       inputStruct.GeneratedBy,
				License:           commitLicense,
				Metadata:          &meta,
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

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {

	req, err := http.NewRequest("GET", urlUserMe, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", getToken(connection))

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		return fmt.Errorf("connection error")
	}
	return nil
}
