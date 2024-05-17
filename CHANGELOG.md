# Changelog

## [0.18.0-beta](https://github.com/instill-ai/component/compare/v0.17.0-beta...v0.18.0-beta) (2024-05-17)


### Features

* **instill:** adopt latest Instill Model endpoints ([#133](https://github.com/instill-ai/component/issues/133)) ([a53661c](https://github.com/instill-ai/component/commit/a53661c565b492c31d319e3f8946c17e5176eae9))

## [0.17.0-beta](https://github.com/instill-ai/component/compare/v0.16.0-beta...v0.17.0-beta) (2024-05-15)


### Features

* add additional attribute in JSON schema for Instill Credit ([#118](https://github.com/instill-ai/component/issues/118)) ([a6751fa](https://github.com/instill-ai/component/commit/a6751fa8737be97e9b895924986798d6429c3986))
* add global secrets to StabilityAI connector ([#122](https://github.com/instill-ai/component/issues/122)) ([1db0c9f](https://github.com/instill-ai/component/commit/1db0c9f9d34db415e5ba157d1263f25d99944bb9))
* add sourceTag for pinecone ([#117](https://github.com/instill-ai/component/issues/117)) ([b202da1](https://github.com/instill-ai/component/commit/b202da1ac36d3a7ca1c4cef8dbb41b19a3d8e986))
* allow global API key on OpenAI connector ([#110](https://github.com/instill-ai/component/issues/110)) ([42bccdd](https://github.com/instill-ai/component/commit/42bccdddf02d4bcc79576762a4bad16fa29c0fb5))
* implement Slack component ([#120](https://github.com/instill-ai/component/issues/120)) ([1ecff8a](https://github.com/instill-ai/component/commit/1ecff8ac7612d785bf172a414d0768dd4df9c084))
* **openai:** support gpt-4o model ([#127](https://github.com/instill-ai/component/issues/127)) ([536f5af](https://github.com/instill-ai/component/commit/536f5af2acc3456249da1b35e16f59a46ee071a6))
* update Instill Credit supported model list ([#123](https://github.com/instill-ai/component/issues/123)) ([0b0cf81](https://github.com/instill-ai/component/commit/0b0cf81a2cc0b01558335d2bad3ffab4dc9911c9))


### Bug Fixes

* Fix the bug of setting channelID ([#125](https://github.com/instill-ai/component/issues/125)) ([47bc192](https://github.com/instill-ai/component/commit/47bc192129a7ddae77166d2b8ae987d5c9b2015d))
* **slack:** add `instillSecret: true` to `token` field ([#126](https://github.com/instill-ai/component/issues/126)) ([7751585](https://github.com/instill-ai/component/commit/7751585698e7388dcaf0e101c55ee6ccb2a19a25))

## [0.16.0-beta](https://github.com/instill-ai/component/compare/v0.15.0-beta...v0.16.0-beta) (2024-04-30)


### Features

* add contribution guide ([#106](https://github.com/instill-ai/component/issues/106)) ([3957579](https://github.com/instill-ai/component/commit/3957579268f4088798ecafeeb6d0553103c3987a))


### Bug Fixes

* typos in contribution guide ([#108](https://github.com/instill-ai/component/issues/108)) ([f01c049](https://github.com/instill-ai/component/commit/f01c0491e8eb171ccae0a076763405800f84383d))

## [0.15.0-beta](https://github.com/instill-ai/component/compare/v0.14.1-beta...v0.15.0-beta) (2024-04-25)


### Features

* add `UsageHandler` interface ([#87](https://github.com/instill-ai/component/issues/87)) ([b9d9645](https://github.com/instill-ai/component/commit/b9d9645f8bdbf5d63eb56b0b0a4510b016870970))
* adjust `IConnector` interface ([#83](https://github.com/instill-ai/component/issues/83)) ([46ea796](https://github.com/instill-ai/component/commit/46ea7960392deb2804c9a097af7871e5960e0523))
* adjust `Test()` interface ([#81](https://github.com/instill-ai/component/issues/81)) ([763cc6d](https://github.com/instill-ai/component/commit/763cc6d000e317bb54af9fccf115ed54994987f4))
* adopt system variables ([#92](https://github.com/instill-ai/component/issues/92)) ([e8ae4e1](https://github.com/instill-ai/component/commit/e8ae4e145fc022f3a86a1f4b93b3fe5967bc44a2))
* **airbyte:** remove Airbyte connectors ([#88](https://github.com/instill-ai/component/issues/88)) ([ec770d6](https://github.com/instill-ai/component/commit/ec770d62cf52fa099a9be6825432d76f0d211f4a))
* **airbyte:** remove local connector and refine definition ([#85](https://github.com/instill-ai/component/issues/85)) ([8203316](https://github.com/instill-ai/component/commit/8203316e021022ae801a10cd6275f8e34d3dd1ab))
* **compogen:** use jsonref when generating the README ([#99](https://github.com/instill-ai/component/issues/99)) ([ff49157](https://github.com/instill-ai/component/commit/ff491574a96b9316a4012fec5b312b6526aa13dc))
* expose `IsCredentialField` interface ([#93](https://github.com/instill-ai/component/issues/93)) ([6cd2801](https://github.com/instill-ai/component/commit/6cd2801569a34b734cf58f59e6e02cb3cd4acd08))
* **instill:** drop support for "external mode" ([#101](https://github.com/instill-ai/component/issues/101)) ([b0c091b](https://github.com/instill-ai/component/commit/b0c091b5e51090046659c82f261f57d76dd41b99))
* merge resource spec into component spec ([#86](https://github.com/instill-ai/component/issues/86)) ([a6de70e](https://github.com/instill-ai/component/commit/a6de70e1e3ab4e46548a8847fc723628a1a09260))


### Bug Fixes

* **airbyte:** add missing release_stage ([#84](https://github.com/instill-ai/component/issues/84)) ([9c0a57d](https://github.com/instill-ai/component/commit/9c0a57d6e66e185bbc58ec968a8d874bad713aca))
* **instill:** add nil check for GetConnectorDefinitionByUID ([0ca6fc3](https://github.com/instill-ai/component/commit/0ca6fc3cf9ecefdcd74e132fbafd2cdcbbfb180e))
* **numbers:** fix recipe data bug ([#103](https://github.com/instill-ai/component/issues/103)) ([36480f8](https://github.com/instill-ai/component/commit/36480f8402d7ed09842c21093781a07d725a4c46))

## [0.14.1-beta](https://github.com/instill-ai/component/compare/v0.14.0-beta...v0.14.1-beta) (2024-04-01)


### Bug Fixes

* **compogen:** better installation command ([#78](https://github.com/instill-ai/component/issues/78)) ([f1b453b](https://github.com/instill-ai/component/commit/f1b453b322ae43df1b18773affc5c65a9fcd3059))
* **instill:** fix multi-region connection problem ([#76](https://github.com/instill-ai/component/issues/76)) ([591a6a2](https://github.com/instill-ai/component/commit/591a6a20233f5d1509cb678e27b133d4dc329e20))

## [0.14.0-beta](https://github.com/instill-ai/component/compare/v0.13.0-beta...v0.14.0-beta) (2024-03-29)


### Features

* document release stages and versions ([#70](https://github.com/instill-ai/component/issues/70)) ([91457dc](https://github.com/instill-ai/component/commit/91457dc5b4cf9bd397d42a70d37f3e75c3398095))
* merge connector and operator repos into this repo ([#72](https://github.com/instill-ai/component/issues/72)) ([2fd6b1d](https://github.com/instill-ai/component/commit/2fd6b1dd65b2b50eeb2f89209b939b397164d272))
* read release stage in auto generated docs from field in definitions ([#68](https://github.com/instill-ai/component/issues/68)) ([90ea333](https://github.com/instill-ai/component/commit/90ea333c3f443e9c3f0a9306ce72317254a61210))
* remove pre-release label in version ([#75](https://github.com/instill-ai/component/issues/75)) ([f0320d3](https://github.com/instill-ai/component/commit/f0320d3a2107daa5f1d9463c0b417a838f957434))


### Bug Fixes

* document pre-release version removal ([#71](https://github.com/instill-ai/component/issues/71)) ([e527a11](https://github.com/instill-ai/component/commit/e527a11e19530b225602bb49ad87fd18ae076ff1))

## [0.13.0-beta](https://github.com/instill-ai/component/compare/v0.12.0-beta...v0.13.0-beta) (2024-03-07)


### Features

* expose description field in component entities ([#62](https://github.com/instill-ai/component/issues/62)) ([85bbc22](https://github.com/instill-ai/component/commit/85bbc223c1df208d0c619af20a6ad693761dc36f))
* simplify `openapi_specifications` to `data_specifications` ([#64](https://github.com/instill-ai/component/issues/64)) ([7c27d15](https://github.com/instill-ai/component/commit/7c27d15e4e01290b728458a2f711029bb600a0a8))


### Bug Fixes

* **vdp:** better casting errors ([#65](https://github.com/instill-ai/component/issues/65)) ([81e34c4](https://github.com/instill-ai/component/commit/81e34c476e97f388d46b9cdb496291748b897c63))

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
