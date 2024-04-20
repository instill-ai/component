package base

import (
	_ "embed"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
)

var (
	//go:embed testdata/connectorDef.json
	connectorDefJSON []byte
	//go:embed testdata/connectorTasks.json
	connectorTasksJSON []byte
	//go:embed testdata/connectorAdditional.json
	connectorAdditionalJSON []byte
	//go:embed testdata/wantConnectorDefinition.json
	wantConnectorDefinitionJSON []byte
)

func TestConnector_ListConnectorDefinitions(t *testing.T) {
	c := qt.New(t)
	logger := zap.NewNop()

	conn := BaseConnector{
		Logger: logger,
	}

	err := conn.LoadConnectorDefinition(
		connectorDefJSON,
		connectorTasksJSON,
		map[string][]byte{"additional.json": connectorAdditionalJSON})
	c.Assert(err, qt.IsNil)

	got, err := conn.GetConnectorDefinition(nil)
	c.Assert(err, qt.IsNil)
	c.Check(wantConnectorDefinitionJSON, qt.JSONEquals, got)
}
