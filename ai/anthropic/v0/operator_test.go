package anthropic

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestOperator_Execute(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	bc := base.Component{Logger: zap.NewNop()}
	component := Init(bc)

	c.Run("ok - greet", func(c *qt.C) {
		exec, err := component.CreateExecution(nil, nil, taskGreet)
		c.Assert(err, qt.IsNil)

		pbIn, err := structpb.NewStruct(map[string]any{"target": "bolero-wombat"})
		c.Assert(err, qt.IsNil)

		got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Check(err, qt.IsNil)
		c.Assert(got, qt.HasLen, 1)

		// Check JSON in the output string.
		greeting := got[0].Fields["greeting"].GetStringValue()
		c.Check(greeting, qt.Equals, "Hello, bolero-wombat!")
	})
}

func TestOperator_CreateExecution(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	operator := Init(bc)

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		_, err := operator.CreateExecution(nil, nil, task)
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})
}
