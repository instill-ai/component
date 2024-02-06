package base

import (
	_ "embed"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
)

var (
	//go:embed testdata/operatorDef.json
	operatorDefJSON []byte
	//go:embed testdata/operatorTasks.json
	operatorTasksJSON []byte
	//go:embed testdata/wantOperatorDefinition.json
	wantOperatorDefinitionJSON []byte
)

func TestOperator_ListOperatorDefinitions(t *testing.T) {
	c := qt.New(t)
	logger := zap.NewNop()

	conn := Operator{
		Component: Component{Logger: logger},
	}

	err := conn.LoadOperatorDefinitions(operatorDefJSON, operatorTasksJSON, nil)
	c.Assert(err, qt.IsNil)

	defs := conn.ListOperatorDefinitions()
	c.Assert(defs, qt.HasLen, 1)

	got := defs[0]
	c.Check(wantOperatorDefinitionJSON, qt.JSONEquals, got)
}
