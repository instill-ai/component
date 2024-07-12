package base

// The JSON schema template for task: TASK_TEXT_GENERATION_CHAT
/**
"$defs": {
        "multi-modal-content": {
            "instillFormat": "structured/multi-modal-content",
            "items": {
                "properties": {
                    "image-url": {
                        "properties": {
                            "url": {
                                "description": "Either a URL of the image or the base64 encoded image data.",
                                "type": "string"
                            }
                        },
                        "required": [
                            "url"
                        ],
                        "type": "object"
                    },
                    "text": {
                        "description": "The text content.",
                        "instillFormat": "string",
                        "type": "string"
                    },
                    "type": {
                        "description": "The type of the content part.",
                        "enum": [
                            "text",
                            "image_url"
                        ],
                        "instillFormat": "string",
                        "type": "string"
                    }
                },
                "required": [
                    "type"
                ],
                "type": "object"
            },
            "type": "array"
        },
        "chat-message": {
            "properties": {
                "content": {
                    "$ref": "#/$defs/multi-modal-content",
                    "description": "The message content",
                    "instillUIOrder": 1,
                    "title": "Content"
                },
                "role": {
                    "description": "The message role, i.e. 'system', 'user' or 'assistant'",
                    "instillFormat": "string",
                    "instillUIOrder": 0,
                    "title": "Role",
                    "type": "string"
                }
            },
            "required": [
                "role",
                "content"
            ],
            "title": "Chat Message",
            "type": "object"
        },
        "usage": {
            "description": "Usage tokens in {VENDOR_NAME}",
            "instillUIOrder": 1,
            "properties": {
                "input-tokens": {
                    "description": "The input tokens read by {VENDOR_NAME} model",
                    "instillFormat": "number",
                    "instillUIOrder": 2,
                    "title": "Input Tokens",
                    "type": "number"
                },
                "output-tokens": {
                    "description": "The output tokens generated by {VENDOR_NAME} model",
                    "instillFormat": "number",
                    "instillUIOrder": 3,
                    "title": "Output Tokens",
                    "type": "number"
                }
            },
            "required": [
                "input-tokens",
                "output-tokens"
            ],
            "title": "Usage",
            "type": "object"
        }
    },
    "TASK_TEXT_GENERATION_CHAT": {
        "instillShortDescription": "Provide text outputs in response to text inputs.",
        "description": "{VENDOR_NAME} text generation models (often called generative pre-trained transformers or large language models) have been trained to understand natural language, code, and images. The models provide text outputs in response to their inputs. The inputs to these models are also referred to as \"prompts\". Designing a prompt is essentially how you \u201cprogram\u201d a large language model model, usually by providing instructions or some examples of how to successfully complete a task.",
        "input": {
            "description": "Input",
            "instillEditOnNodeFields": [
                "prompt",
                "model-name"
            ],
            "instillUIOrder": 0,
            "properties": {
                "chat-history": {
                    "description": "Incorporate external chat history, specifically previous messages within the conversation. Please note that System Message will be ignored and will not have any effect when this field is populated. Each message should adhere to the format: : {\"role\": \"The message role, i.e. 'system', 'user' or 'assistant'\", \"content\": \"message content\"}.",
                    "instillAcceptFormats": [
                        "structured/chat-messages"
                    ],
                    "instillShortDescription": "Incorporate external chat history, specifically previous messages within the conversation.",
                    "instillUIOrder": 4,
                    "instillUpstreamTypes": [
                        "reference"
                    ],
                    "items": {
                        "$ref": "#/$defs/chat-message"
                    },
                    "title": "Chat history",
                    "type": "array"
                },
                "max-new-tokens": {
                    "default": 50,
                    "description": "The maximum number of tokens for model to generate",
                    "instillAcceptFormats": [
                        "integer"
                    ],
                    "instillUIOrder": 6,
                    "instillUpstreamTypes": [
                        "value",
                        "reference"
                    ],
                    "title": "Max new tokens",
                    "type": "integer"
                },
                "model-name": {
                    "enum": [
                        "{VENDOR_MODEL-1}"
                    ],
                    "example": "{VENDOR_MODEL-1}",
                    "description": "The {VENDOR_NAME} model to be used",
                    "instillAcceptFormats": [
                        "string"
                    ],
                    "instillUIOrder": 0,
                    "instillUpstreamTypes": [
                        "value",
                        "reference",
                        "template"
                    ],
                    "instillCredentialMap": {
                        "values": [
                            "{VENDOR_MODEL-1}"
                        ],
                        "targets": [
                            "setup.api-key"
                        ]
                    },
                    "title": "Model Name",
                    "type": "string"
                },
                "prompt": {
                    "description": "The prompt text",
                    "instillAcceptFormats": [
                        "string"
                    ],
                    "instillUIMultiline": true,
                    "instillUIOrder": 2,
                    "instillUpstreamTypes": [
                        "value",
                        "reference",
                        "template"
                    ],
                    "title": "Prompt",
                    "type": "string"
                },
                "prompt-images": {
                    "description": "The prompt images",
                    "instillAcceptFormats": [
                        "array:image/*"
                    ],
                    "instillUIOrder": 3,
                    "instillUpstreamTypes": [
                        "reference"
                    ],
                    "items": {
                        "type": "string"
                    },
                    "title": "Prompt Images",
                    "type": "array"
                },
                "seed": {
                    "description": "The seed",
                    "instillAcceptFormats": [
                        "integer"
                    ],
                    "instillUIOrder": 4,
                    "instillUpstreamTypes": [
                        "value",
                        "reference"
                    ],
                    "title": "Seed",
                    "type": "integer"
                },
                "system-message": {
                    "default": "You are a helpful assistant.",
                    "description": "The system message helps set the behavior of the assistant. For example, you can modify the personality of the assistant or provide specific instructions about how it should behave throughout the conversation. By default, the model\u2019s behavior is set using a generic message as \"You are a helpful assistant.\"",
                    "instillAcceptFormats": [
                        "string"
                    ],
                    "instillShortDescription": "The system message helps set the behavior of the assistant",
                    "instillUIMultiline": true,
                    "instillUIOrder": 2,
                    "instillUpstreamTypes": [
                        "value",
                        "reference",
                        "template"
                    ],
                    "title": "System message",
                    "type": "string"
                },
                "temperature": {
                    "default": 0.7,
                    "description": "The temperature for sampling",
                    "instillAcceptFormats": [
                        "number"
                    ],
                    "instillUIOrder": 5,
                    "instillUpstreamTypes": [
                        "value",
                        "reference"
                    ],
                    "title": "Temperature",
                    "type": "number"
                },
                "top-k": {
                    "default": 10,
                    "description": "Top k for sampling",
                    "instillAcceptFormats": [
                        "integer"
                    ],
                    "instillUIOrder": 5,
                    "instillUpstreamTypes": [
                        "value",
                        "reference"
                    ],
                    "title": "Top K",
                    "type": "integer"
                }
            },
            "required": [
                "prompt",
                "model-name"
            ],
            "title": "Input",
            "type": "object"
        },
        "output": {
            "description": "Output",
            "instillUIOrder": 0,
            "properties": {
                "text": {
                    "description": "Model Output",
                    "instillUIOrder": 0,
                    "instillFormat": "string",
                    "instillUIMultiline": true,
                    "title": "Text",
                    "type": "string"
                },
                "usage": {
                    "$ref": "#/$defs/usage"
                }
            },
            "required": [
                "text"
            ],
            "title": "Output",
            "type": "object"
        }
    }


*/
type ChatMessage struct {
	Role    string              `json:"role"`
	Content []MultiModalContent `json:"content"`
}
type URL struct {
	URL string `json:"url"`
}

type MultiModalContent struct {
	ImageURL URL    `json:"image-url"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

type TemplateTextGenerationInput struct {
	ChatHistory  []ChatMessage `json:"chat-history"`
	MaxNewTokens int           `json:"max-new-tokens"`
	ModelName    string        `json:"model-name"`
	Prompt       string        `json:"prompt"`
	PromptImages []string      `json:"prompt-images"`
	Seed         int           `json:"seed"`
	SystemMsg    string        `json:"system-message"`
	Temperature  float64       `json:"temperature"`
	TopK         int           `json:"top-k"`
}

type GenerativeTextModelUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}

type EmbeddingTextModelUsage struct {
	Tokens int `json:"tokens"`
}

type TemplateTextGenerationOutput struct {
	Text  string                   `json:"text"`
	Usage GenerativeTextModelUsage `json:"usage"`
}
