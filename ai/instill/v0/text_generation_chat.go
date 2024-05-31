package instill

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeTextGenerationChat(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		llmInput := e.convertLLMInput(input)
		taskInput := &modelPB.TaskInput_TextGenerationChat{
			TextGenerationChat: &modelPB.TextGenerationChatInput{
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
		taskOutputs, err := trigger(grpcClient, e.SystemVariables, modelName, []*modelPB.TaskInput{{Input: taskInput}})
		if err != nil {
			return nil, err
		}
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
		}

		textGenChatOutput := taskOutputs[0].GetTextGenerationChat()
		if textGenChatOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textGenChatOutput, modelName)
		}
		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(textGenChatOutput)
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
