package vertexai

import (
	"context"
	_ "embed"
	"fmt"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"github.com/instill-ai/component/base"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type speechRecognitionInput struct {
	Audio      string `json:"audio"`
	ModelName  string `json:"model-name"`
	SampleRate int    `json:"sample-rate"`
	Language   string `json:"language-code"`
}

type speechRecognitionOutput struct {
	Text string `json:"text"`
}

func (e *execution) speechRecognition(in *structpb.Struct) (*structpb.Struct, error) {
	setupStruct := vertexAISetup{}
	err := base.ConvertFromStructpb(e.GetSetup(), &setupStruct)
	if err != nil {
		return nil, err
	}
	inputStruct := speechRecognitionInput{}
	err = base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	credJSON := []byte(setupStruct.Cred)
	client, err := speech.NewClient(ctx, option.WithCredentialsJSON(credJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: int32(inputStruct.SampleRate),
			LanguageCode:    inputStruct.Language,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: []byte(inputStruct.Audio),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error recognizing speech: %w", err)
	}

	outputStruct := speechRecognitionOutput{}

	for _, result := range resp.Results {
		outputStruct.Text += result.Alternatives[0].Transcript
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}
