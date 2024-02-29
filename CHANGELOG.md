# Changelog

## [0.12.0-beta](https://github.com/instill-ai/component/compare/v0.11.0-beta...v0.12.0-beta) (2024-02-27)


### Features

* create README.mdx generation command ([#59](https://github.com/instill-ai/component/issues/59)) ([c814c05](https://github.com/instill-ai/component/commit/c814c05eaa68b9a62b7f9a9ab8fc1253586fb197))
* extract task title generation ([#58](https://github.com/instill-ai/component/issues/58)) ([52804d4](https://github.com/instill-ai/component/commit/52804d408e44f4e92cc6d7b734fc577962a900f0))

## [0.11.0-beta](https://github.com/instill-ai/component/compare/v0.10.0-beta...v0.11.0-beta) (2024-02-14)


### Features

* add tasks to component definition ([#54](https://github.com/instill-ai/component/issues/54)) ([b067f9c](https://github.com/instill-ai/component/commit/b067f9cbaa984349f482e9cc6b6011a0b14b240b))
* introduce `instillFormat: semi-structured/json` ([#55](https://github.com/instill-ai/component/issues/55)) ([3dbaa03](https://github.com/instill-ai/component/commit/3dbaa03ff7085a851b6eb424c67733173667b447))


### Bug Fixes

* fix `instillFormat` validation when using `semi-structured` and `structured data` ([#56](https://github.com/instill-ai/component/issues/56)) ([66fea88](https://github.com/instill-ai/component/commit/66fea88f30d8f95f178235170ca1df1b1c9ab083))
* fix bug when `instillAcceptFormats` has multiple values ([#52](https://github.com/instill-ai/component/issues/52)) ([5cbfb44](https://github.com/instill-ai/component/commit/5cbfb4402c3ccc8aa379d6281161059a45183c9f))

## [0.10.0-beta](https://github.com/instill-ai/component/compare/v0.9.0-beta...v0.10.0-beta) (2024-01-28)


### Features

* add task title and description in component json schema ([#49](https://github.com/instill-ai/component/issues/49)) ([0878f99](https://github.com/instill-ai/component/commit/0878f99479f26d1a082d9477dc823e18be3fbae7))

## [0.9.0-beta](https://github.com/instill-ai/component/compare/v0.8.0-beta...v0.9.0-beta) (2024-01-12)


### Features

* add `instillUpstreamTypes: template` in component condition field ([#46](https://github.com/instill-ai/component/issues/46)) ([60b2117](https://github.com/instill-ai/component/commit/60b21171a6abaffb9381b38a4c73fd63fa8e2489))
* **schema:** add new instillFormat for chat history ([#43](https://github.com/instill-ai/component/issues/43)) ([abed794](https://github.com/instill-ai/component/commit/abed794dc3a122025ab4978fd7b4646aa8c6ae74))
* update GetOperatorDefinition functions to support dynamic definition ([#47](https://github.com/instill-ai/component/issues/47)) ([792559e](https://github.com/instill-ai/component/commit/792559e0b538742c5e53b9f5269ff22e98345d44))


### Bug Fixes

* **connector:** fix credentialFields bug inside `oneOf` schema ([#45](https://github.com/instill-ai/component/issues/45)) ([eb76043](https://github.com/instill-ai/component/commit/eb76043417a30d2dfb3cfc03087bc3417183c88d))

## [0.8.0-beta](https://github.com/instill-ai/component/compare/v0.7.1-alpha...v0.8.0-beta) (2023-12-15)


### Miscellaneous Chores

* release v0.8.0-beta ([3c1c85c](https://github.com/instill-ai/component/commit/3c1c85c3d9a57ef8ad7b21b39c7d37bb3f736cf9))

## [0.7.1-alpha](https://github.com/instill-ai/component/compare/v0.7.0-alpha...v0.7.1-alpha) (2023-11-28)


### Miscellaneous Chores

* release v0.7.1-alpha ([0015dff](https://github.com/instill-ai/component/commit/0015dfffca247b0b44a1bac6beb6d0ef81c61127))

## [0.7.0-alpha](https://github.com/instill-ai/component/compare/v0.6.1-alpha...v0.7.0-alpha) (2023-11-09)


### Features

* **component:** support json reference ([#20](https://github.com/instill-ai/component/issues/20)) ([bafafe9](https://github.com/instill-ai/component/commit/bafafe960082eb4b6b85137cc85cf71ff6dec987))


### Bug Fixes

* **component:** fix `instillShortDescription` parser bug ([#28](https://github.com/instill-ai/component/issues/28)) ([7f528ae](https://github.com/instill-ai/component/commit/7f528aef1869ce1db2f541a26d6791cabcecd59a))
* **component:** fix jsonreference pointer bug ([#29](https://github.com/instill-ai/component/issues/29)) ([dc5371e](https://github.com/instill-ai/component/commit/dc5371eae05c071a7d4ee72193a4a49a81614e7d))
* **schema:** fix schema inconsistent naming ([#23](https://github.com/instill-ai/component/issues/23)) ([dd7aa52](https://github.com/instill-ai/component/commit/dd7aa52dc2cd5185733543ec0d9171b161cee149))

## [0.6.1-alpha](https://github.com/instill-ai/component/compare/v0.6.0-alpha...v0.6.1-alpha) (2023-10-27)


### Miscellaneous Chores

* **release:** release v0.6.1-alpha ([da0483c](https://github.com/instill-ai/component/commit/da0483cefac5d39c585fc802799b65f83d70e554))

## [0.6.0-alpha](https://github.com/instill-ai/component/compare/v0.5.0-alpha...v0.6.0-alpha) (2023-10-13)


### Bug Fixes

* **execution:** fix empty task bug ([6e15dc3](https://github.com/instill-ai/component/commit/6e15dc306543495f1bc75b8b8c11d5099a843471))


### Miscellaneous Chores

* **release:** release v0.6.0-alpha ([120a147](https://github.com/instill-ai/component/commit/120a147ecaeb06d52613e738b81d6158bd74211c))

## [0.5.0-alpha](https://github.com/instill-ai/component/compare/v0.4.0-alpha...v0.5.0-alpha) (2023-09-30)


### Miscellaneous Chores

* **release:** release v0.5.0-alpha ([dd16044](https://github.com/instill-ai/component/commit/dd16044d675986f75f260935dbebeb9714d6f802))

## [0.4.0-alpha](https://github.com/instill-ai/connector/compare/v0.3.0-alpha...v0.4.0-alpha) (2023-09-13)


### Miscellaneous Chores

* **release:** release v0.4.0-alpha ([725b63f](https://github.com/instill-ai/connector/commit/725b63f948366db1670b2b0d34a0309c5ebb06c6))

## [0.3.0-alpha](https://github.com/instill-ai/connector/compare/v0.2.0-alpha...v0.3.0-alpha) (2023-08-03)


### Miscellaneous Chores

* **release:** release v0.3.0-alpha ([dfe81c0](https://github.com/instill-ai/connector/commit/dfe81c05fea87a887f94628d3908251961c0058e))

## [0.2.0-alpha](https://github.com/instill-ai/connector/compare/v0.1.0-alpha...v0.2.0-alpha) (2023-07-20)


### Miscellaneous Chores

* **release:** release v0.2.0-alpha ([fa946bd](https://github.com/instill-ai/connector/commit/fa946bd6ad4984ecba92e5fd9d0c477bd7fe21ef))

## [0.1.0-alpha](https://github.com/instill-ai/connector/compare/v0.1.0-alpha...v0.1.0-alpha) (2023-07-09)


### Features

* Added object mapper implementation and basic tests ([#7](https://github.com/instill-ai/connector/issues/7)) ([a91364b](https://github.com/instill-ai/connector/commit/a91364b7e08866259296810743803341a2097612))


### Miscellaneous Chores

* **release:** release v0.1.0-alpha ([6984052](https://github.com/instill-ai/connector/commit/6984052f8e5a80201b90e82580f10f8872c86d7e))
