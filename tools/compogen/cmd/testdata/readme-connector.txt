# Setup

mkdir -p pkg/dummy/config
cp definition.json pkg/dummy/config/definition.json
cp tasks.json pkg/dummy/config/tasks.json

# OK

compogen readme ./pkg/dummy/config ./pkg/dummy/README.mdx --connector
cmp pkg/dummy/README.mdx want-readme.mdx

-- definition.json --
{
  "available_tasks": [
    "TASK_DUMMY"
  ],
  "public": true,
  "id": "dummy",
  "title": "Dummy",
  "description": "Perform an action",
  "prerequisites": "An account at [dummy.io](https://dummy.io) is required.",
  "type": "CONNECTOR_TYPE_DATA",
  "spec": {
    "connection_specification": {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "additionalProperties": true,
      "properties": {
        "organization": {
          "description": "Specify which organization is used for the requests",
          "instillUIOrder": 1,
          "title": "Organization ID",
          "type": "string"
        },
        "api_key": {
          "description": "Fill your Dummy API key",
          "instillUIOrder": 0,
          "title": "API Key",
          "type": "string"
        }
      },
      "required": [
        "api_key"
      ],
      "title": "OpenAI Connector Resource",
      "type": "object"
    }
  },
  "release_stage": "RELEASE_STAGE_COMING_SOON",
  "source_url": "https://github.com/instill-ai/component/blob/main/pkg/connector/dummy/v0"
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
        }
      }
    }
  }
}
-- want-readme.mdx --
---
title: "Dummy"
lang: "en-US"
draft: false
description: "Learn about how to set up a VDP Dummy connector https://github.com/instill-ai/instill-core"
---

The Dummy component is a data connector that allows users to perform an action.
It can carry out the following tasks:

- [Dummy](#dummy)

## Release Stage

`Coming Soon`

## Configuration

The component configuration is defined and maintained [here](https://github.com/instill-ai/component/blob/main/pkg/connector/dummy/v0/config/definition.json).

## Connection

<InfoBlock type="info" title="Prerequisites">An account at [dummy.io](https://dummy.io) is required.</InfoBlock>

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
| API Key (required) | `api_key` | string | Fill your Dummy API key |
| Organization ID | `organization` | string | Specify which organization is used for the requests |

## Supported Tasks

### Dummy

| Input | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Task ID (required) | `task` | string | `TASK_DUMMY` |
| Durna (required) | `durna` | string | Lorem ipsum dolor sit amet, consectetur adipiscing elit |

| Output | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Orci (optional) | `orci` | string | Orci sagittis eu volutpat odio facilisis mauris sit |
