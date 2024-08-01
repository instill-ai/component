package instill

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeVisualQuestionAnswering(grpcClient modelPB.ModelPublicServiceClient, nsID string, modelID string, version string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s/%s/%s", inputs, nsID, modelID, version)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		llmInput := e.convertLLMInput(input)
		taskInput := &modelPB.TaskInput_VisualQuestionAnswering{
			VisualQuestionAnswering: &modelPB.VisualQuestionAnsweringInput{
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
		taskOutputs, err := trigger(grpcClient, e.SystemVariables, nsID, modelID, version, []*modelPB.TaskInput{{Input: taskInput}})
		if err != nil {
			return nil, err
		}
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", taskOutputs, nsID, modelID, version)
		}

		visualQuestionAnsweringOutput := taskOutputs[0].GetVisualQuestionAnswering()
		if visualQuestionAnsweringOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", visualQuestionAnsweringOutput, nsID, modelID, version)
		}
		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(visualQuestionAnsweringOutput)
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
