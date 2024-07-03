package base

import (
	"context"
	"testing"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	qt "github.com/frankban/quicktest"
)

func TestComponentExecution_GetComponent(t *testing.T) {
	c := qt.New(t)

	cmp := &testComp{Component{}}
	err := cmp.LoadDefinition(
		connectorDefJSON,
		connectorConfigJSON,
		connectorTasksJSON,
		map[string][]byte{"additional.json": connectorAdditionalJSON})
	c.Assert(err, qt.IsNil)

	x, _ := cmp.CreateExecution(nil, nil, "TASK_TEXT_EMBEDDINGS")
	got := x.Execution.GetComponent()
	c.Check(got.GetID(), qt.Equals, "openai")
}

type testExec struct{ ComponentExecution }

func (e *testExec) Execute(_ context.Context, _ []*structpb.Struct) ([]*structpb.Struct, error) {
	return nil, nil
}

type testComp struct{ Component }

func (c *testComp) CreateExecution(_ map[string]any, _ *structpb.Struct, task string) (*ExecutionWrapper, error) {
	x := ComponentExecution{Component: c, Task: task}
	return &ExecutionWrapper{Execution: &testExec{x}}, nil
}
