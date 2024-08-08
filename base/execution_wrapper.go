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

type outputWriterInterceptor struct {
	ow      OutputWriter
	outputs []*structpb.Struct
}

func (ow *outputWriterInterceptor) Write(ctx context.Context, outputs []*structpb.Struct) (err error) {
	ow.outputs = outputs
	return ow.ow.Write(ctx, outputs)
}

func (ow *outputWriterInterceptor) GetOutputs(ctx context.Context) (outputs []*structpb.Struct, err error) {
	return ow.outputs, nil
}

// Execute wraps the execution method with validation and usage collection.
func (e *ExecutionWrapper) Execute(ctx context.Context, ir InputReader, ow OutputWriter) error {

	inputs, err := ir.Read(ctx)
	if err != nil {
		return err
	}
	if err := Validate(inputs, e.GetTaskInputSchema(), "inputs"); err != nil {
		return err
	}

	newUH := e.GetComponent().UsageHandlerCreator()
	h, err := newUH(e)
	if err != nil {
		return err
	}

	if err := h.Check(ctx, inputs); err != nil {
		return err
	}

	owi := &outputWriterInterceptor{ow: ow}
	err = e.IExecution.Execute(ctx, ir, owi)
	if err != nil {
		return err
	}

	outputs, err := owi.GetOutputs(ctx)
	if err != nil {
		return err
	}

	if err := Validate(outputs, e.GetTaskOutputSchema(), "outputs"); err != nil {
		return err
	}

	if err := h.Collect(ctx, inputs, outputs); err != nil {
		return err
	}

	return nil

}
