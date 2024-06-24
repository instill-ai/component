package base

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	//go:embed testdata/connectorDef.json
	connectorDefJSON []byte
	//go:embed testdata/connectorTasks.json
	connectorTasksJSON []byte
	//go:embed testdata/connectorConfig.json
	connectorConfigJSON []byte
	//go:embed testdata/connectorAdditional.json
	connectorAdditionalJSON []byte
	//go:embed testdata/wantConnectorDefinition.json
	wantConnectorDefinitionJSON []byte
)

func TestComponent_ListConnectorDefinitions(t *testing.T) {
	c := qt.New(t)
	logger := zap.NewNop()

	conn := Component{
		Logger: logger,
	}

	err := conn.LoadDefinition(
		connectorDefJSON,
		connectorConfigJSON,
		connectorTasksJSON,
		map[string][]byte{"additional.json": connectorAdditionalJSON})
	c.Assert(err, qt.IsNil)

	got, err := conn.GetDefinition(nil, nil)
	c.Assert(err, qt.IsNil)
	gotJSON, err := protojson.Marshal(got)
	c.Assert(err, qt.IsNil)

	wantConnectorDefinitionStruct := map[string]any{}
	err = json.Unmarshal(wantConnectorDefinitionJSON, &wantConnectorDefinitionStruct)
	c.Assert(err, qt.IsNil)
	c.Check(gotJSON, qt.JSONEquals, wantConnectorDefinitionStruct)
}

func TestUtil_GetFileExtension(t *testing.T) {
	c := qt.New(t)

	file, err := os.Open("./testdata/test_image.png")
	c.Assert(err, qt.IsNil)
	defer file.Close()
	wantFileExtension := "png"

	reader := bufio.NewReader(file)
	content, err := io.ReadAll(reader)
	c.Assert(err, qt.IsNil)

	fileBase64 := base64.StdEncoding.EncodeToString(content)
	fileBase64 = "data:image/png;base64," + fileBase64
	fmt.Println(fileBase64)
	gotFileExtension := GetBase64FileExtensionSlow(fileBase64)
	c.Check(gotFileExtension, qt.Equals, wantFileExtension)
	gotFileExtension = GetBase64FileExtensionFast(fileBase64)
	c.Check(gotFileExtension, qt.Equals, wantFileExtension)
}
