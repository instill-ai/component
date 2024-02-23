# Setup

mkdir -p pkg/dummy/config
cp definitions.json pkg/dummy/config/definitions.json
cp tasks.json pkg/dummy/config/tasks.json

# NOK - Wrong files

! compogen readme pkg/dummy/wrong pkg/dummy/README.mdx --operator
cmp stderr want-no-defs

mkdir -p pkg/dummy/wrong
cp definitions.json pkg/dummy/wrong/definitions.json
! compogen readme pkg/dummy/wrong pkg/dummy/README.mdx --operator
cmp stderr want-no-tasks

! compogen readme pkg/dummy/config pkg/wrong/README.mdx --operator
cmp stderr want-wrong-target

# OK

compogen readme ./pkg/dummy/config ./pkg/dummy/README.mdx --operator
cmp pkg/dummy/README.mdx want-readme.mdx

-- definitions.json --
[
  {
    "available_tasks": [
      "TASK_DUMMY",
      "TASK_DUMMIER_THAN_DUMMY"
    ],
    "public": true,
    "spec": {},
    "id": "dummy",
    "title": "Dummy",
    "description": "Performs an action",
    "version": "0.1.0-alpha",
    "source_url": "https://github.com/instill-ai/operator/blob/main/pkg/dummy/v0"
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
  },
  "TASK_DUMMIER_THAN_DUMMY": {
    "title": "Dummier",
    "instillShortDescription": "This task is dummier than `TASK_DUMMY`.",
    "input": {
      "properties": {
        "cursus": {
          "description": "Cursus mattis molestie a iaculis at erat pellentesque adipiscing commodo",
          "instillUIOrder": 0,
          "title": "Cursus",
          "type": "string"
        }
      },
      "required": [
        "cursus"
      ]
    },
    "output": {
      "properties": {
        "elementum": {
          "description": "Tellus elementum sagittis vitae et",
          "instillUIOrder": 0,
          "title": "Elementum",
          "type": "string"
        },
        "errors": {
          "description": "Error messages, if any, during the dummy process",
          "instillUIOrder": 3,
          "title": "Errors",
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "meta": {
          "description": "Donec ac meta tempor orci dapibus ultrices in",
          "instillUIOrder": 1,
          "title": "Meta",
          "type": "object",
          "properties": {
            "tortor": {
              "description": "Tincidunt tortor aliquam nulla",
              "instillUIOrder": 0,
              "title": "Tincidunt tortor",
              "type": "string"
            },
            "arcu": {
              "description": "Bibendum arcu vitae elementum curabitur vitae nunc sed velit",
              "instillUIOrder": 1,
              "title": "Arcu",
              "type": "string"
            }
          },
          "required": []
        },
        "nullam_non": {
          "description": "Id faucibus nisl tincidunt eget nullam non",
          "instillUIOrder": 2,
          "title": "Nullam non",
          "type": "number"
        }
      },
      "required": [
        "elementum",
        "meta",
        "nullam_non",
        "error"
      ]
    }
  }
}
-- want-no-defs --
Error: open pkg/dummy/wrong/definitions.json: no such file or directory
-- want-no-tasks --
Error: open pkg/dummy/wrong/tasks.json: no such file or directory
-- want-wrong-target --
Error: open pkg/wrong/README.mdx: no such file or directory
-- want-invalid-def --
Error: invalid definitions file:
Definitions field has an invalid length
-- want-readme.mdx --
---
title: "Dummy"
lang: "en-US"
draft: false
description: "Learn about how to set up a VDP Dummy operator https://github.com/instill-ai/vdp"
---

The Dummy component is an operator that performs an action.
It can carry out the following tasks:

- [Dummy](#dummy)
- [Dummier](#dummier)

## Release Stage

`Alpha`

## Configuration

The component configuration is defined and maintained [here](https://github.com/instill-ai/operator/blob/main/pkg/dummy/v0/config/definitions.json).

## Supported Tasks

### Dummy

| Input | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Task ID (required) | `task` | string | `TASK_DUMMY` |
| Durna (required) | `durna` | string | Lorem ipsum dolor sit amet, consectetur adipiscing elit |

| Output | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Orci (optional) | `orci` | string | Orci sagittis eu volutpat odio facilisis mauris sit |

### Dummier

This task is dummier than `TASK_DUMMY`.

| Input | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Task ID (required) | `task` | string | `TASK_DUMMIER_THAN_DUMMY` |
| Cursus (required) | `cursus` | string | Cursus mattis molestie a iaculis at erat pellentesque adipiscing commodo |

| Output | ID | Type | Description |
| :--- | :--- | :--- | :--- |
| Elementum | `elementum` | string | Tellus elementum sagittis vitae et |
| Meta | `meta` | object | Donec ac meta tempor orci dapibus ultrices in |
| Nullam non | `nullam_non` | number | Id faucibus nisl tincidunt eget nullam non |
| Errors (optional) | `errors` | array[string] | Error messages, if any, during the dummy process |
