# Setup

mkdir -p pkg/dummy/config
mv definitions.json pkg/dummy/config/definitions.json

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
  "title": "Dummy",
  "description": "Performs an action",
  "type": "CONNECTOR_TYPE_DATA",
  "version": "0.1.0-alpha",
  "source_url": "github.com/instill-ai/connector/blob/main/pkg/dummy/v0"
}
]
-- want-readme.mdx --
---
title: "Dummy"
lang: "en-US"
draft: false
description: "Learn about how to set up a VDP Dummy connector https://github.com/instill-ai/vdp"
---

The Dummy component is a data connector that performs an action.
It can carry out the following tasks:

- [Dummy](#dummy)

## Release Stage

`Alpha`

## Supported Tasks

### Dummy

| Input | Type | Description |
| :--- | --- | --- |
| task | string | `TASK_DUMMY` |
