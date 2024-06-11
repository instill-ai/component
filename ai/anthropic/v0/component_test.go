package anthropic

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
)

func TestOperator_Execute(t *testing.T) {

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
