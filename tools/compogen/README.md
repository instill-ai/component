# `compogen`

`compogen` is a generation tool for Instill AI component schemas. It uses the
information in a component schema to automatically generate the component
documentation.

## Installation

```shell
go install github.com/instill-ai/component/tools/compogen@latest
```

## Generate the documentation of a component

`compogen` can generate the README of a component by reading its schemas. The
format of the documentation is MDX, so the generated files can directly be used
in the Instill AI website.

```shell
compogen readme path/to/component/config path/to/component/README.mdx
```

### Validation & guidelines

In order to successfully build the README of a component, the `definition.json`
and `tasks.json` files must be present in the component configuration directory.

The `definition.json` file must contain an array with one object in which the
following fields must be present and comply with the following guidelines:

- `id`.
- `title`.
- `description` - It should contain a single sentence describing the component.
  The template will use it next to the component title (`{{ .Title }}{{
  .Description }}.`) so it must be written in imperative tense.
- `release_stage` - Must be the string representation of one of the nonzero
  values of `ComponentDefinition.ReleaseStage`,defined in
  [protobufs](https://github.com/instill-ai/protobufs/blob/main/vdp/pipeline/v1beta/connector_definition.proto).
- `type` - Connector definitions must contain this field and its value must
  match one of the (string) values defined in [protobufs](https://github.com/instill-ai/protobufs/blob/main/vdp/pipeline/v1beta/connector_definition.proto).
- `available_tasks` - This array must have at least one value, which should be
  one of the root-level keys in the `tasks.json` file.
- `source_url` - Must be a valid URL. It must not end with a slash, as the
  definitions path will be appended.

Certain optional fields modify the document behaviour:

- `public`, when `true`, will set the `draft` property to `false`.
- For connector components, the content of `prerequisites` will be displayed in
  an info block next to the resource configuration details.
  - Note that this section only applies when a connector is being documented,
    i.e. when the `--connector` flag is passed.`
- A table will be built for the `spec.connection_specification` properties. They
  must contain an `instillUIOrder` field so the row order is deterministic.

## TODO

- Support `oneOf` schemas for resource properties, present in, e.g., the [REST API](https://github.com/instill-ai/component/blob/main/application/restapi/v0/config/definition.json#L26) connectors.
  - We might leverage some Go implementation of JSON schema. Some candidates:
    - [santhosh-tekuri/jsonschema](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v5#Schema)
    - [omissis/go-jsonschema](https://github.com/omissis/go-jsonschema/blob/934012d/pkg/schemas/model.go#L107)
    - [invopop/jsonschema](https://github.com/invopop/jsonschema/blob/a446707/schema.go#L14)
    - [swaggest/jsonschema-go](https://pkg.go.dev/github.com/swaggest/jsonschema-go#Schema)
  - The schema loading carried out by the `component/base` package in
    `LoadConnectorDefinition` or `LoadOperatorDefinition` might also be
    useful, although it is oriented to transforming the data to a `structpb.Struct`
    rather than to define the object structure.
- In the "supported tasks" tables, provide better documentation for nested
  arrays and objects (currently the type doesn't support nesting).
- If task definitions contain examples for the (required) input and output
  fields, generate param samples as in https://github.com/instill-ai/instill.tech/blob/dedaaa3/docs/v0.12.0-beta/vdp/ai-connectors/openai.en.mdx#L148
- Implement a way to inject extra sections if a component needs further
  documentation (e.g. by adding a `doc.json` file with a structured array that
  describes the position and content of the new section.

## Next steps

- `compogen validate` might be used validate any component configuration.
- `compogen new [--operator]` might be used to generate the skeleton of a component.
- In the future we might want to generate documentation in different languages.
This will require some thought.
