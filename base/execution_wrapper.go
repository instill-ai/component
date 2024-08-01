package base

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"
)

var _ IExecution = &ExecutionWrapper{}

// ExecutionWrapper performs validation and usage collection around the
// execution of a component.
type ExecutionWrapper struct {
	IExecution
}

// Execute wraps the execution method with validation and usage collection.
func (e *ExecutionWrapper) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if err := Validate(inputs, e.GetTaskInputSchema(), "inputs"); err != nil {
		return nil, err
	}

	newUH := e.GetComponent().UsageHandlerCreator()
	h, err := newUH(e)
	if err != nil {
		return nil, err
	}

	if err := h.Check(ctx, inputs); err != nil {
		return nil, err
	}

	outputs, err := e.IExecution.Execute(ctx, inputs)
	if err != nil {
		return nil, err
	}

	if err := Validate(outputs, e.GetTaskOutputSchema(), "outputs"); err != nil {
		return nil, err
	}

	if err := h.Collect(ctx, inputs, outputs); err != nil {
		return nil, err
	}

	return outputs, err
}
