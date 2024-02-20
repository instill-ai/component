# Setup

mkdir -p pkg/dummy/config
mv definitions.json pkg/dummy/config/definitions.json

# NOK - Wrong files

! compogen readme pkg/dummy/wrong pkg/dummy/README.mdx
cmp stderr want-wrong-config

! compogen readme pkg/dummy/config pkg/wrong/README.mdx
cmp stderr want-wrong-target

# OK

compogen readme ./pkg/dummy/config ./pkg/dummy/README.mdx
cmp pkg/dummy/README.mdx want-readme.mdx

-- want-wrong-config --
Error: open pkg/dummy/wrong/definitions.json: no such file or directory
-- want-wrong-target --
Error: open pkg/wrong/README.mdx: no such file or directory
-- definitions.json --
[
  {
    "available_tasks": [
    ],
    "public": true,
    "title": "Dummy",
    "version": "0.1.0-alpha",
    "source_url": "github.com/instill-ai/operator/blob/main/pkg/base64/v0"
  }
]
-- want-readme.mdx --
---
title: "Dummy"
lang: "en-US"
draft: false
description: "Learn about how to set up a VDP Dummy operator https://github.com/instill-ai/vdp"
---

The Dummy component is an operator that performs an action.
It can carry out the following tasks:

- [Text Generation](#text-generation)
- [Text Embeddings](#text-embeddings)
- [Speech Recognition](#speech-recognition)
- [Text to Image](#text-to-image)

## Release Stage

`Alpha`
