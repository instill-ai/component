# Setup

mkdir -p pkg/dummy/config
cp definitions.json pkg/dummy/config/definitions.json
cp tasks.json pkg/dummy/config/tasks.json

# OK

compogen readme ./pkg/dummy/config ./pkg/dummy/README.mdx --connector
cmp pkg/dummy/README.mdx want-readme.mdx

-- definitions.json --
[
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
      "resource_specification": {
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
    "source_url": "https://github.com/instill-ai/connector/blob/main/pkg/dummy/v0"
  }
]
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
description: "Learn about how to set up a VDP Dummy connector https://github.com/instill-ai/vdp"
---

The Dummy component is a data connector that allows users to perform an action.
It can carry out the following tasks:

- [Dummy](#dummy)

## Release Stage

`Coming Soon`

## Configuration

The component configuration is defined and maintained [here](https://github.com/instill-ai/connector/blob/main/pkg/dummy/v0/config/definitions.json).

<InfoBlock type="info" title="Prerequisites">An account at [dummy.io](https://dummy.io) is required.</InfoBlock>

| Field | Field ID | Type | Note |
| :--- | :--- | :--- | :--- |
| API Key (required) | `api_key` | string | Fill your Dummy API key |
| Organization ID | `organization` | string | Specify which organization is used for the requests |

Dummy connector resources can be created in two ways:

### No-code Setup

1. Go to the **Connectors** page and click **+ Create Connector**.
1. Select **Dummy**.
1. Fill in a unique ID for the resource. Optionally, give a short description in the **Description** field.
1. Fill in the required fields described in [Resource Configuration](#resource-configuration).

### Low-code Setup

<CH.Code>

```shellscript cURL(Instill-Cloud)
curl -X POST https://api.instill.tech/vdp/v1beta/users/<user-id>/connectors \
--header 'Authorization: Bearer <Instill-Cloud-API-Token>' \
--data '{
  "id": "my-dummy",
  "connector_definition_name": "connector-definitions/dummy",
  "description": "Optional description",
  "configuration": {
    "api_key": <api-key>,
    "organization": <organization>
  }
}'
```

```shellscript cURL(Instill-Core)
curl -X POST http://localhost:8080/vdp/v1beta/users/<user-id>/connectors \
--header 'Authorization: Bearer <Instill-Core-API-Token>' \
--data '{
  "id": "my-dummy",
  "connector_definition_name": "connector-definitions/dummy",
  "description": "Optional description",
  "configuration": {
    "api_key": <api-key>,
    "organization": <organization>
  }
}'
```

</CH.Code>

For other component operations, please refer to the [API reference](https://openapi.instill.tech/reference/pipelinepublicservice).

## Supported Tasks

### Dummy

| Input | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Task ID (required) | `task` | string | `TASK_DUMMY` |
| Durna (required) | `durna` | string | Lorem ipsum dolor sit amet, consectetur adipiscing elit |

| Output | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Orci (optional) | `orci` | string | Orci sagittis eu volutpat odio facilisis mauris sit |
