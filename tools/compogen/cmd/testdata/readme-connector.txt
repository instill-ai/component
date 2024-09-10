# Setup

mkdir -p pkg/dummy/config
cp definition.json pkg/dummy/config/definition.json
cp setup.json pkg/dummy/config/setup.json
cp tasks.json pkg/dummy/config/tasks.json

mkdir -p pkg/dummy/.compogen
cp extra-setup.mdx pkg/dummy/.compogen/extra-setup.mdx

# OK

compogen readme ./pkg/dummy/config ./pkg/dummy/README.mdx --extraContents setup=./pkg/dummy/.compogen/extra-setup.mdx
cmp pkg/dummy/README.mdx want-readme.mdx

-- definition.json --
{
  "availableTasks": [
    "TASK_DUMMY"
  ],
  "public": true,
  "id": "dummy",
  "title": "Dummy",
  "vendor": "Dummy Inc.",
  "description": "Perform an action",
  "prerequisites": "An account at [dummy.io](https://dummy.io) is required.",
  "type": "COMPONENT_TYPE_DATA",
  "releaseStage": "RELEASE_STAGE_COMING_SOON",
  "sourceUrl": "https://github.com/instill-ai/component/blob/main/data/dummy/v0"
}

-- setup.json --
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "additionalProperties": true,
  "properties": {
    "organization": {
      "description": "Specify which organization is used for the requests",
      "instillUIOrder": 1,
      "title": "Organization ID",
      "type": "string"
    },
    "api-key": {
      "description": "Fill in your Dummy API key",
      "instillUIOrder": 0,
      "title": "API Key",
      "type": "string"
    },
    "authentication": {
      "description": "Authentication method to use for the Dummy",
      "instillUIOrder": 0,
      "oneOf": [
        {
          "properties": {
            "auth-type": {
              "const": "NO_AUTH",
              "description": "No Authentication",
              "instillUIOrder": 0,
              "order": 0,
              "type": "string"
            }
          },
          "required": [
            "auth-type"
          ],
          "title": "No Auth"
        },
        {
          "properties": {
            "auth-type": {
              "const": "AUTH_1",
              "description": "Auth 1",
              "instillUIOrder": 0,
              "order": 0,
              "type": "string"
            },
            "auth-way": {
              "description": "ways for Auth 1",
              "instillUpstreamTypes": [
                "value"
              ],
              "instillAcceptFormats": [
                "string"
              ],
              "enum": [
                "header",
                "query"
              ],
              "instillUIOrder": 1,
              "order": 1,
              "title": "Auth Way",
              "type": "string"
            }
          },
          "required": [
            "auth-type",
            "auth-way"
          ],
          "title": "Auth 1"
        }
      ],
      "order": 1,
      "title": "Authentication",
      "type": "object"
    }
  },
  "required": [
    "api-key"
  ],
  "title": "OpenAI Connection",
  "type": "object"
}

-- tasks.json --
{
  "TASK_DUMMY": {
    "input": {
      "properties": {
        "durna": {
          "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit",
          "instillUIOrder": 0,
          "title": "Durna",
          "type": "string"
        },
        "strategy": {
          "description": "Chunking strategy",
          "instillUIOrder": 1,
          "properties": {
            "setting": {
              "description": "Chunk Setting",
              "additionalProperties": true,
              "type": "object",
              "title": "Chunk Setting",
              "instillUIOrder": 0,
              "required": [
                "chunk-method"
              ],
              "oneOf": [
                {
                  "properties": {
                    "chunk-method": {
                      "const": "Token",
                      "type": "string"
                    },
                    "model-name": {
                      "description": "The name of the model used for tokenization.",
                      "enum": [
                        "gpt-4",
                        "gpt-3.5-turbo"
                      ],
                      "instillUIOrder": 1,
                      "title": "Model",
                      "type": "string"
                    }
                  },
                  "required": ["chunk-method"],
                  "type": "object"
                },
                {
                  "properties": {
                    "chunk-method": {
                      "const": "Markdown",
                      "type": "string"
                    },
                    "model-name": {
                      "description": "The name of the model used for tokenization.",
                      "enum": [
                        "gpt-4",
                        "gpt-3.5-turbo"
                      ],
                      "instillUIOrder": 1,
                      "title": "Model",
                      "type": "string"
                    }
                  },
                  "required": ["chunk-method"],
                  "type": "object"
                }
              ]
            }
          },
          "title": "Strategy",
          "required": [
            "setting"
          ],
          "type": "object"
        }
      },
      "required": [
        "durna"
      ]
    },
    "output": {
      "properties": {
        "orci": {
          "description": "Orci sagittis eu volutpat odio facilisis mauris sit",
          "instillFormat": "string",
          "instillUIOrder": 0,
          "title": "Orci",
          "type": "string"
        },
        "conversations": {
          "description": "An array of conversations with thread messages",
          "instillUIOrder": 0,
          "title": "Conversations",
          "type": "array",
          "items": {
            "title": "conversation details",
            "type": "object",
            "properties": {
              "message": {
                "description": "message to start a conversation",
                "instillUIOrder": 0,
                "title": "Start Conversation Message",
                "type": "string"
              },
              "start-date": {
                "description": "when a conversation starts",
                "instillUIOrder": 1,
                "title": "Start Date",
                "type": "string"
              },
              "last-date": {
                "description": "Date of the last message",
                "instillUIOrder": 2,
                "title": "Last Date",
                "type": "string"
              },
              "thread-reply-messages": {
                "description": "replies in a conversation",
                "instillUIOrder": 0,
                "title": "Replied messages",
                "type": "array",
                "items": {
                  "title": "relied details",
                  "type": "object",
                  "properties": {
                    "message": {
                      "description": "message to reply a conversation",
                      "instillFormat": "string",
                      "instillUIOrder": 3,
                      "title": "Replied Message",
                      "type": "string"
                    }
                  },
                  "required": [
                    "message"
                  ]
                }
              }
            },
            "required": [
              "message",
              "start-date"
            ]
          }
        }
      }
    }
  }
}
-- extra-setup.mdx --
This is some crucial information about setup: do it before execution.
-- want-readme.mdx --
---
title: "Dummy"
lang: "en-US"
draft: false
description: "Learn about how to set up a VDP Dummy component https://github.com/instill-ai/instill-core"
---

The Dummy component is a data component that allows users to perform an action.
It can carry out the following tasks:

- [Dummy](#dummy)



## Release Stage

`Coming Soon`



## Configuration

The component configuration is defined and maintained [here](https://github.com/instill-ai/component/blob/main/data/dummy/v0/config/definition.json).




## Setup

<InfoBlock type="info" title="Prerequisites">An account at [dummy.io](https://dummy.io) is required.</InfoBlock>




In order to communicate with Dummy Inc., the following connection details need to be
provided. You may specify them directly in a pipeline recipe as key-value pairs
withing the component's `setup` block, or you can create a **Connection** from
the [**Integration Settings**](https://www.instill.tech/docs/vdp/integration)
page and reference the whole `setup` as `setup:
${connection.<my-connection-id>}`.

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
| API Key (required) | `api-key` | string | Fill in your Dummy API key  |
| Authentication | `authentication` | object | Authentication method to use for the Dummy  |
| Organization ID | `organization` | string | Specify which organization is used for the requests  |

This is some crucial information about setup: do it before execution.







#### authentication option: 0

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
|  auth-type   | auth-type | string |  please intput  "NO_AUTH"   |

#### authentication option: 1

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
|  auth-type   | auth-type | string |  please intput  "AUTH_1"   |
| Auth Way   | auth-way | string |  ways for Auth 1    There are options: <br/>- header<br/>- query   |


## Supported Tasks

### Dummy


| Input | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Task ID (required) | `task` | string | `TASK_DUMMY` |
| Durna (required) | `durna` | string | Lorem ipsum dolor sit amet, consectetur adipiscing elit |
| Strategy | `strategy` | object | Chunking strategy |






#### Strategy

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
| Chunk Setting | setting | object | Chunk Setting  |








#### setting option: 0

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
|  chunk-method   | chunk-method | string |  please intput  "Token"   |
| Model   | model-name | string |  The name of the model used for tokenization.    There are options: <br/>- gpt-4<br/>- gpt-3.5-turbo   |


#### setting option: 1

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
|  chunk-method   | chunk-method | string |  please intput  "Markdown"   |
| Model   | model-name | string |  The name of the model used for tokenization.    There are options: <br/>- gpt-4<br/>- gpt-3.5-turbo   |





| Output | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Conversations (optional) | `conversations` | array[object] | An array of conversations with thread messages |
| Orci (optional) | `orci` | string | Orci sagittis eu volutpat odio facilisis mauris sit |







#### Conversations

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
| Last Date | last-date | string | Date of the last message |
| Start Conversation Message | message | string | message to start a conversation |
| Start Date | start-date | string | when a conversation starts |
| Replied messages | thread-reply-messages | array | replies in a conversation |



#### Replied messages

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
| Replied Message | message | string | message to reply a conversation |







