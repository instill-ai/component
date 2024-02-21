# Setup

mkdir -p pkg/dummy/config
mv definitions.json pkg/dummy/config/definitions.json

mkdir -p pkg/wrongdef/config
mv wrongdef.json pkg/wrongdef/config/definitions.json

# NOK - Wrong files

! compogen readme pkg/dummy/wrong pkg/dummy/README.mdx --operator
cmp stderr want-wrong-config

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
-- wrongdef.json --
[
]
-- want-wrong-config --
Error: open pkg/dummy/wrong/definitions.json: no such file or directory
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
- [Dummier Than Dummy](#dummier-than-dummy)

## Release Stage

`Alpha`

## Configuration

The component configuration is defined and maintained [here](https://github.com/instill-ai/operator/blob/main/pkg/dummy/v0/config/definitions.json).

## Supported Tasks

### Dummy

| Input | Type | Description |
| :--- | --- | --- |
| task | string | `TASK_DUMMY` |

### Dummier Than Dummy

| Input | Type | Description |
| :--- | --- | --- |
| task | string | `TASK_DUMMIER_THAN_DUMMY` |

