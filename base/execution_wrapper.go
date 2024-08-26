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

type inputReaderInterceptor struct {
	InputReader
	schema       string
	usageHandler UsageHandler
	inputs       []*structpb.Struct
}

func (ir *inputReaderInterceptor) Read(ctx context.Context) (inputs []*structpb.Struct, err error) {
	inputs, err = ir.InputReader.Read(ctx)
	if err != nil {
		return nil, err
	}
	if err = Validate(inputs, ir.schema, "inputs"); err != nil {
		return nil, err
	}

	if err = ir.usageHandler.Check(ctx, inputs); err != nil {
		return nil, err
	}
	ir.inputs = inputs
	return inputs, nil
}

type outputWriterInterceptor struct {
	OutputWriter
	schema                 string
	inputReaderInterceptor *inputReaderInterceptor
	usageHandler           UsageHandler
}

func (ow *outputWriterInterceptor) Write(ctx context.Context, outputs []*structpb.Struct) (err error) {

	if err := Validate(outputs, ow.schema, "outputs"); err != nil {
		return err
	}

	if err := ow.usageHandler.Collect(ctx, ow.inputReaderInterceptor.inputs, outputs); err != nil {
		return err
	}

	return ow.OutputWriter.Write(ctx, outputs)

}

// Execute wraps the execution method with validation and usage collection.
func (e *ExecutionWrapper) Execute(ctx context.Context, ir InputReader, ow OutputWriter) error {

	newUH := e.GetComponent().UsageHandlerCreator()
	h, err := newUH(e)
	if err != nil {
		return err
	}

	iri := &inputReaderInterceptor{
		InputReader:  ir,
		schema:       e.GetTaskInputSchema(),
		usageHandler: h,
	}

	owi := &outputWriterInterceptor{
		OutputWriter:           ow,
		schema:                 e.GetTaskOutputSchema(),
		inputReaderInterceptor: iri,
		usageHandler:           h,
	}

	if err := e.IExecution.Execute(ctx, iri, owi); err != nil {
		return err
	}

	return nil

}
