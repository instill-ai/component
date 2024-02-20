# compogen

`compogen` is a generation tool for Instill AI component schemas. It uses the
information in a component schema to automatically generate the component
documentation.

## Installation

```shell
git clone https://github.com/instill-ai/component
cd component/tools/compogen
go install .
```

## Generate the documentation of a component

`compogen` can generate the README of a component by reading its schemas. The
format of the documentation is MDX, so the generated files can directly be used
in the Instill AI website.

```shell
compogen readme path/to/component/config path/to/component/README.mdx
```

### Validation & guidelines

In order to successfully build the README of a component, the following fields
must be present and comply with the following guidelines:

- `title`
- `description` - It should contain a single sentence describing the component.
  The template will use it next to the component title (`{{ .Title }}{{
  .Description }}.`) so it must be written in third person, present tense.
- `version` - Must be valid SemVer 2.0.0
- `type` - Connector definitions must contain this field and its value must
  match one of the (string) values defined in [protobufs](https://github.com/instill-ai/protobufs/blob/main/vdp/pipeline/v1beta/connector_definition.proto)
