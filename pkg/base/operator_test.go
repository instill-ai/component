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

	op := BaseOperator{
		Logger: logger,
	}

	err := op.LoadOperatorDefinition(operatorDefJSON, operatorTasksJSON, nil)
	c.Assert(err, qt.IsNil)

	got, err := op.GetOperatorDefinition(nil)
	c.Assert(err, qt.IsNil)
	c.Check(wantOperatorDefinitionJSON, qt.JSONEquals, got)
}
