package artifact

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/instill-ai/component/base"
	artifactPB "github.com/instill-ai/protogen-go/artifact/artifact/v1alpha"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

type RetrieveChatHistoryInput struct {
	Namespace       string `json:"namespace"`
	CatalogID       string `json:"catalog-id"`
	ConversationID  string `json:"conversation-id"`
	Role            string `json:"role"`
	MessageType     string `json:"message-type"`
	Duration        string `json:"duration"`
	MaxMessageCount int    `json:"max-message-count"`
}

type RetrieveChatHistoryOutput struct {
	Messages      []Message `json:"messages"`
	NextPageToken string    `json:"next-page-token"`
}

type Message struct {
	MessageUID     string `json:"message-uid"`
	CatalogID      string `json:"catalog-id"`
	ConversationID string `json:"conversation-id"`
	Role           string `json:"role"`
	MessageType    string `json:"message-type"`
	Content        string `json:"content"`
	CreateTime     string `json:"create-time"`
	UpdateTime     string `json:"update-time"`
}

func (in *RetrieveChatHistoryInput) Validate() error {
	if in.Role != "" && in.Role != "user" && in.Role != "assistant" {
		return fmt.Errorf("role must be either 'user' or 'assistant'")
	}

	if in.MessageType != "" && in.MessageType != "MESSAGE_TYPE_TEXT" {
		return fmt.Errorf("message-type must be 'MESSAGE_TYPE_TEXT'")
	}

	if in.Duration != "" {
		_, err := time.ParseDuration(in.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
	}
	return nil
}

func (out *RetrieveChatHistoryOutput) Filter(inputStruct RetrieveChatHistoryInput, messages []*artifactPB.Message) {
	for _, message := range messages {

		if inputStruct.Role != "" && inputStruct.Role != message.Role {
			continue
		}

		if inputStruct.MessageType != "" && inputStruct.MessageType != message.Type.String() {
			continue
		}

		if inputStruct.Duration != "" {
			duration, _ := time.ParseDuration(inputStruct.Duration)
			if time.Since(message.CreateTime.AsTime()) > duration {
				continue
			}
		}
		if inputStruct.MaxMessageCount > 0 && len(out.Messages) >= inputStruct.MaxMessageCount {
			break
		}
		out.Messages = append(out.Messages, Message{
			MessageUID:     message.Uid,
			CatalogID:      message.CatalogUid,
			ConversationID: message.ConversationUid,
			Role:           message.Role,
			MessageType:    message.Type.String(),
			Content:        message.Content,
			CreateTime:     message.CreateTime.AsTime().Format(time.RFC3339),
			UpdateTime:     message.UpdateTime.AsTime().Format(time.RFC3339),
		})
	}
}

func (e *execution) retrieveChatHistory(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := RetrieveChatHistoryInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input struct: %w", err)
	}

	err = inputStruct.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	artifactClient, connection := e.client, e.connection
	defer connection.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(e.SystemVariables))

	res, err := artifactClient.ListMessages(ctx, &artifactPB.ListMessagesRequest{
		NamespaceId:           inputStruct.Namespace,
		CatalogId:             inputStruct.CatalogID,
		ConversationId:        inputStruct.ConversationID,
		IncludeSystemMessages: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	output := RetrieveChatHistoryOutput{
		Messages: make([]Message, 0),
	}

	output.Filter(inputStruct, res.Messages)

	for res.NextPageToken != "" || (len(output.Messages) < inputStruct.MaxMessageCount && inputStruct.MaxMessageCount > 0) {
		res, err = artifactClient.ListMessages(ctx, &artifactPB.ListMessagesRequest{
			NamespaceId:           inputStruct.Namespace,
			CatalogId:             inputStruct.CatalogID,
			ConversationId:        inputStruct.ConversationID,
			IncludeSystemMessages: true,
			PageToken:             res.NextPageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to list messages: %w", err)
		}

		output.Filter(inputStruct, res.Messages)
		output.NextPageToken = res.NextPageToken
	}

	return base.ConvertToStructpb(output)

}

type WriteChatMessageInput struct {
	Namespace      string `json:"namespace"`
	CatalogID      string `json:"catalog-id"`
	ConversationID string `json:"conversation-id"`
	Content        string `json:"content"`
	Role           string `json:"role"`
	MessageType    string `json:"message-type"`
}

type WriteChatMessageOutput struct {
	MessageUID     string `json:"message-uid"`
	CatalogID      string `json:"catalog-id"`
	ConversationID string `json:"conversation-id"`
	Role           string `json:"role"`
	MessageType    string `json:"message-type"`
	Content        string `json:"content"`
	CreateTime     string `json:"create-time"`
	UpdateTime     string `json:"update-time"`
}

func (in *WriteChatMessageInput) Validate() error {
	if in.Role != "" && in.Role != "user" && in.Role != "assistant" {
		return fmt.Errorf("role must be either 'user' or 'assistant'")
	}

	if in.MessageType != "" && in.MessageType != "MESSAGE_TYPE_TEXT" {
		return fmt.Errorf("message-type must be 'MESSAGE_TYPE_TEXT'")
	}

	return nil
}

func (e *execution) writeChatMessage(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := WriteChatMessageInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input struct: %w", err)
	}

	err = inputStruct.Validate()

	if err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	artifactClient, connection := e.client, e.connection
	defer connection.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(e.SystemVariables))

	_, err = artifactClient.CreateConversation(ctx, &artifactPB.CreateConversationRequest{
		NamespaceId:    inputStruct.Namespace,
		CatalogId:      inputStruct.CatalogID,
		ConversationId: inputStruct.ConversationID,
	})

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Println("Conversation already exists")
		} else {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
	}

	res, err := artifactClient.CreateMessage(ctx, &artifactPB.CreateMessageRequest{
		NamespaceId:    inputStruct.Namespace,
		CatalogId:      inputStruct.CatalogID,
		ConversationId: inputStruct.ConversationID,
		Role:           inputStruct.Role,
		Type:           artifactPB.Message_MessageType(artifactPB.Message_MessageType_value[inputStruct.MessageType]),
		Content:        inputStruct.Content,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	messageOutput := res.Message

	output := WriteChatMessageOutput{
		MessageUID:     messageOutput.Uid,
		CatalogID:      inputStruct.CatalogID,
		ConversationID: inputStruct.ConversationID,
		Role:           messageOutput.Role,
		MessageType:    messageOutput.Type.String(),
		Content:        messageOutput.Content,
		CreateTime:     messageOutput.CreateTime.AsTime().Format(time.RFC3339),
		UpdateTime:     messageOutput.UpdateTime.AsTime().Format(time.RFC3339),
	}

	return base.ConvertToStructpb(output)

}
