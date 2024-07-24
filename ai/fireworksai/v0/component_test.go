package fireworksai

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
)

func TestComponent_Execute(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	c.Run("ok - supported task", func(c *qt.C) {
		task := TaskTextGenerationChat

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.IsNil)
	})
	c.Run("ok - supported task", func(c *qt.C) {
		task := TaskTextEmbeddings

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.IsNil)
	})

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})
}
