# Component

---

## ⚠️  Deprecation of this repository

This repository is in process of being moved to
[`pipeline-backend`](https://github.com/instill-ai/pipeline-backend/tree/main/pkg/component).
Please, update and import the code in / from that repository instead of this
one.

A few remaining tasks remain before archiving the repository:
- [ ] Port the Pull Requests to `pipeline-backend`.
- [ ] Move [`compogen`](./tools/compogen) to `pipeline-backend`.
- [ ] Merge the contribution guidelines in the new repository.

---

A **component** is the basic building block of the [**Instill Core**](https://github.com/instill-ai/instill-core) pipeline. The pipeline consists of multiple components.
We have two types of components: connectors and operators.
This Golang package defines the common interface and functions for all components.

## Contributing

Please refer to the [Contributing Guidelines](./.github/CONTRIBUTING.md) for more details.

## Community support

Please refer to the [community](https://github.com/instill-ai/community) repository.

## License

See the [LICENSE](./LICENSE) file for licensing information.
