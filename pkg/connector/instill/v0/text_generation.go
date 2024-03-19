package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *Execution) executeTextGeneration(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		llmInput := e.convertLLMInput(input)
		taskInput := &modelPB.TaskInput_TextGeneration{
			TextGeneration: &modelPB.TextGenerationInput{
				Prompt:        llmInput.Prompt,
				PromptImages:  llmInput.PromptImages,
				ChatHistory:   llmInput.ChatHistory,
				SystemMessage: llmInput.SystemMessage,
				MaxNewTokens:  llmInput.MaxNewTokens,
				Temperature:   llmInput.Temperature,
				TopK:          llmInput.TopK,
				Seed:          llmInput.Seed,
				ExtraParams:   llmInput.ExtraParams,
			},
		}

		// only support batch 1
		req := modelPB.TriggerUserModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}
		ctx := metadata.NewOutgoingContext(context.Background(), getRequestMetadata(e.Config))
		res, err := grpcClient.TriggerUserModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		taskOutputs := res.GetTaskOutputs()
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
		}

		textGenOutput := taskOutputs[0].GetTextGeneration()
		if textGenOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textGenOutput, modelName)
		}
		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(textGenOutput)
		if err != nil {
			return nil, err
		}
		output := &structpb.Struct{}
		err = protojson.Unmarshal(outputJSON, output)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)

	}
	return outputs, nil
}
