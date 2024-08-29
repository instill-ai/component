# Changelog

## [0.26.0-beta](https://github.com/instill-ai/component/compare/v0.25.0-beta...v0.26.0-beta) (2024-08-29)


### Features

* add milvus component ([#299](https://github.com/instill-ai/component/issues/299)) ([a48c211](https://github.com/instill-ai/component/commit/a48c211f8b0b8c1cf663fad01937542a90855dc2))
* add zilliz component ([#297](https://github.com/instill-ai/component/issues/297)) ([92726ef](https://github.com/instill-ai/component/commit/92726ef6a77c79a4d516a24b385996abfc895eb9))
* **artifact:** improve artifact component ([#289](https://github.com/instill-ai/component/issues/289)) ([44ea196](https://github.com/instill-ai/component/commit/44ea1963f065dd64ab47b301e376492f2601c6ac))
* **document:** improve document operator ([#287](https://github.com/instill-ai/component/issues/287)) ([c0d8d31](https://github.com/instill-ai/component/commit/c0d8d31bd58c0f64e8e2790cd3706c0e173e11d9))
* **hubspot:** add 4 tasks and modify Retrieve Association task and Get Thread task ([#265](https://github.com/instill-ai/component/issues/265)) ([62903ec](https://github.com/instill-ai/component/commit/62903ecba99df296398db335bd9000380194c8f1))
* introduce interfaces InputReader and OutputWriter ([#294](https://github.com/instill-ai/component/issues/294)) ([e26ecef](https://github.com/instill-ai/component/commit/e26ecef9df5295742e2c509ea19821a5affc558e))
* **jira:** add action tasks ([#241](https://github.com/instill-ai/component/issues/241)) ([e756e31](https://github.com/instill-ai/component/commit/e756e3187d0463c472c016a8eaeb0b16d799a760))
* make the API key be optional for Instill-Credit-supported component ([#305](https://github.com/instill-ai/component/issues/305)) ([0f9a7b2](https://github.com/instill-ai/component/commit/0f9a7b2d29c31188b910446b1c21ce6c7a3e6261))
* **openai:** revert go-openai and add support for streaming ([#301](https://github.com/instill-ai/component/issues/301)) ([aa605fa](https://github.com/instill-ai/component/commit/aa605fabd78869d7f07c446314642c3856591b86))
* **openai:** use go-openai client ([#295](https://github.com/instill-ai/component/issues/295)) ([aa20a16](https://github.com/instill-ai/component/commit/aa20a16365e981299634a4a346a0e81fc1a7ca21))
* **sql:** add ssl/tls input as base64 encoded and move engine to setup ([#282](https://github.com/instill-ai/component/issues/282)) ([390e2b8](https://github.com/instill-ai/component/commit/390e2b838301b21b2529c9ef5b4091a50deaeb5b))
* use error type for component definition not found error ([#302](https://github.com/instill-ai/component/issues/302)) ([cfcee78](https://github.com/instill-ai/component/commit/cfcee7893783f1002193b8fbc5a53bcb4d856ed3))
* **web:** improve web operator ([#292](https://github.com/instill-ai/component/issues/292)) ([1da84af](https://github.com/instill-ai/component/commit/1da84af9ce6ff2d072dbfac3d2e58fe08b3a2c1b))


### Bug Fixes

* **document:** catch the error if there is no data in sheet ([#296](https://github.com/instill-ai/component/issues/296)) ([21bebbd](https://github.com/instill-ai/component/commit/21bebbd6d5a861bb8df39337fa890597c72c1fe1))
* **hubspot:** fix test code ([#298](https://github.com/instill-ai/component/issues/298)) ([98c4261](https://github.com/instill-ai/component/commit/98c4261323a06091d067df3679d0520dab10a288))
* **text:** fix chunk position bugs ([#307](https://github.com/instill-ai/component/issues/307)) ([cfc9076](https://github.com/instill-ai/component/commit/cfc907630fb309294b2a5e15d145a0a3a0be180d))
* **text:** fix the bug if there are 2 exact same chunks ([#308](https://github.com/instill-ai/component/issues/308)) ([b58909f](https://github.com/instill-ai/component/commit/b58909f36591e82a5e1f0162e085082117b2cc9b))

## [0.25.0-beta](https://github.com/instill-ai/component/compare/v0.24.0-beta...v0.25.0-beta) (2024-08-13)


### Features

* add a hook to avoid we miss make document ([#244](https://github.com/instill-ai/component/issues/244)) ([4c4531d](https://github.com/instill-ai/component/commit/4c4531d4e2ef08e796c16ba10717486265fb2ae9))
* add elasticsearch component ([#211](https://github.com/instill-ai/component/issues/211)) ([eb492ca](https://github.com/instill-ai/component/commit/eb492ca6d72f87ed35db8df17ab84b54f9a230be))
* add Fireworks AI component ([#237](https://github.com/instill-ai/component/issues/237)) ([0c40652](https://github.com/instill-ai/component/commit/0c406524f6904044610dc8055ab8204cf7b32318))
* add Groq component ([#269](https://github.com/instill-ai/component/issues/269)) ([1401220](https://github.com/instill-ai/component/commit/140122058bcc9d65588267f2692248f6625957db))
* add mongodb component ([#198](https://github.com/instill-ai/component/issues/198)) ([2cb550f](https://github.com/instill-ai/component/commit/2cb550f7483a20d1635040171ae227fa3542fe17))
* add qdrant component ([#271](https://github.com/instill-ai/component/issues/271)) ([bd2b9e6](https://github.com/instill-ai/component/commit/bd2b9e6f0e614b58f52707c6c3bacbe63d49f05e))
* add weaviate component ([#246](https://github.com/instill-ai/component/issues/246)) ([cb3e667](https://github.com/instill-ai/component/commit/cb3e667bc1277471fcba4998f691abeb4fa8383e))
* add WhatsApp component ([#226](https://github.com/instill-ai/component/issues/226)) ([28d0de8](https://github.com/instill-ai/component/commit/28d0de87523a4b8054f610dd0f7fd35441776750))
* **artifact:** add artifact component ([#268](https://github.com/instill-ai/component/issues/268)) ([dabf472](https://github.com/instill-ai/component/commit/dabf472ce21d9ac6430e76e0e8a7552c2fb9b034))
* **artifact:** add artifact component ([#275](https://github.com/instill-ai/component/issues/275)) ([15fc0d2](https://github.com/instill-ai/component/commit/15fc0d2eb054bc1e6c24a973d3ae1bf417064c16))
* **document:** integrate pdf2md in document operator ([#277](https://github.com/instill-ai/component/issues/277)) ([07360d1](https://github.com/instill-ai/component/commit/07360d13b70f8dc4e406dc045a499a79a905d04b))
* **groq, fireworksai:** take out the unsupported models from instill credit ([#283](https://github.com/instill-ai/component/issues/283)) ([8978acd](https://github.com/instill-ai/component/commit/8978acdefd22337502d7b932428f572565635fb4))
* make component ID accessible on IExecution ([#257](https://github.com/instill-ai/component/issues/257)) ([dd63656](https://github.com/instill-ai/component/commit/dd636560e9d82d66e9fdead03549a3cafd671288))
* **openai:** support `gpt-4o-2024-08-06` and structured output ([#280](https://github.com/instill-ai/component/issues/280)) ([8bdaef7](https://github.com/instill-ai/component/commit/8bdaef74781362763b488664ca39dd09ae6a1d76))
* **sql:** add TASK_INSERT_MANY and fix sql query validation ([#252](https://github.com/instill-ai/component/issues/252)) ([3a93cea](https://github.com/instill-ai/component/commit/3a93cea082a0acf22171ce5db4958d2e1c39efc2))
* **text:** add tokenizer for cohere & new gpt-4o ([#276](https://github.com/instill-ai/component/issues/276)) ([5d8cec3](https://github.com/instill-ai/component/commit/5d8cec362516154e24555b81c2cc1f55f13e417b))
* **text:** revert "add tokenizer for cohere & new gpt-4o ([#276](https://github.com/instill-ai/component/issues/276))" ([910a330](https://github.com/instill-ai/component/commit/910a330e40e2d275b81a670126152c96404f0d4c))


### Bug Fixes

* **artifact:** add the description to remind users to add file extension ([#281](https://github.com/instill-ai/component/issues/281)) ([5ff5d7a](https://github.com/instill-ai/component/commit/5ff5d7a0bdd4cd82f06857abafc9a078928346a2))
* ignore bold case and add all line to result ([#272](https://github.com/instill-ai/component/issues/272)) ([219c77e](https://github.com/instill-ai/component/commit/219c77e6ecc67cd3b53e041b7f228c1f6bf3bdc1))

## [0.24.0-beta](https://github.com/instill-ai/component/compare/v0.23.0-beta...v0.24.0-beta) (2024-07-31)


### Features

* add audio operator ([#236](https://github.com/instill-ai/component/issues/236)) ([fe8abff](https://github.com/instill-ai/component/commit/fe8abff3525528772005da193ee49f5b3dd7c9ed))
* add handler to auto-fill missing default values ([#210](https://github.com/instill-ai/component/issues/210)) ([dcad3f0](https://github.com/instill-ai/component/commit/dcad3f013263a5b8c2649d96431fb929b69e4d98))
* add HubSpot component ([#199](https://github.com/instill-ai/component/issues/199)) ([b3936a8](https://github.com/instill-ai/component/commit/b3936a84dac53562f6d50506918a2d98341ea7c6))
* add Jira component ([#205](https://github.com/instill-ai/component/issues/205)) ([51f3ed7](https://github.com/instill-ai/component/commit/51f3ed78470ab5a82a200953d15e9abe5338dcee))
* add Ollama component ([#224](https://github.com/instill-ai/component/issues/224)) ([810f850](https://github.com/instill-ai/component/commit/810f85080c8f7297db38c8187919430e85365765))
* add sql component ([#193](https://github.com/instill-ai/component/issues/193)) ([9a373f3](https://github.com/instill-ai/component/commit/9a373f3f84cf53bea6d4847a7afdf6349a7d63d2))
* add token count for each chunk ([#235](https://github.com/instill-ai/component/issues/235)) ([bb69104](https://github.com/instill-ai/component/commit/bb691049863fe0474f3975474e673cc51bef8d16))
* add video operator to fulfil unstructured data process ([#238](https://github.com/instill-ai/component/issues/238)) ([a1459d7](https://github.com/instill-ai/component/commit/a1459d709f3abbe2b746070647cb4667612df4b1))
* **document:** add docx doc pptx ppt html to transform to text in markdown format ([#232](https://github.com/instill-ai/component/issues/232)) ([2932db9](https://github.com/instill-ai/component/commit/2932db94abec9e3bca768cc31c02ecf8b24622c1))
* **document:** move ConvertToText task from text operator to document operator ([#248](https://github.com/instill-ai/component/issues/248)) ([699ca70](https://github.com/instill-ai/component/commit/699ca70b2474e0e985d14cac9bf6498b79dbdc86))
* introduce event handler interface ([#253](https://github.com/instill-ai/component/issues/253)) ([9599b42](https://github.com/instill-ai/component/commit/9599b4246e4253938a8a2299116ab433ae3b9e6c))
* **restapi:** recategorize the restapi component as a generic component ([#249](https://github.com/instill-ai/component/issues/249)) ([fbfc3a3](https://github.com/instill-ai/component/commit/fbfc3a312734e6b90a45cd5fe62e24b2ba2e7471))
* **website:** add scrape sitemap function ([#239](https://github.com/instill-ai/component/issues/239)) ([8648326](https://github.com/instill-ai/component/commit/86483265e6dfb16245563937a5c08deeeceebc7b))


### Bug Fixes

* bug of duplicate document ([#256](https://github.com/instill-ai/component/issues/256)) ([e028a6e](https://github.com/instill-ai/component/commit/e028a6e2eed230616dfc23619f544820320c29ac))
* bug of json without setting array for images ([#259](https://github.com/instill-ai/component/issues/259)) ([4aeae69](https://github.com/instill-ai/component/commit/4aeae6975fc82bbed8455d318d4c1b56fc2748e8))
* change md format to html tag for correct frontend link ([#240](https://github.com/instill-ai/component/issues/240)) ([7e16b2b](https://github.com/instill-ai/component/commit/7e16b2b7f494acccad06fd5f43fd2667d8eeadbb))
* revert the alias because they are same as package name ([#243](https://github.com/instill-ai/component/issues/243)) ([1d9c42d](https://github.com/instill-ai/component/commit/1d9c42d70487d96324ef095c8abc4158479b7b76))

## [0.23.0-beta](https://github.com/instill-ai/component/compare/v0.22.0-beta...v0.23.0-beta) (2024-07-19)


### Features

* add new models in open ai ([#229](https://github.com/instill-ai/component/issues/229)) ([b8e39ae](https://github.com/instill-ai/component/commit/b8e39ae535b2a795295c191ff810898791ceaa3b))


### Bug Fixes

* fix markdown chunking bugs ([#228](https://github.com/instill-ai/component/issues/228)) ([2194773](https://github.com/instill-ai/component/commit/21947732e259ea471d654c1041eef5e360f247ff))
* **github:** patch missing fields ([#227](https://github.com/instill-ai/component/issues/227)) ([c61b134](https://github.com/instill-ai/component/commit/c61b13495fe73c4c52aae863528e59588227209e))
* **restapi:** fix response body missing problem ([#222](https://github.com/instill-ai/component/issues/222)) ([47a28dd](https://github.com/instill-ai/component/commit/47a28dd52521ac3bb2480dc9eed0e6548136245d))
* **text:** fix bug and replace markdown chunking ([#221](https://github.com/instill-ai/component/issues/221)) ([298c91a](https://github.com/instill-ai/component/commit/298c91a474314301dd5982aaf5cc60cb1229c7e9))

## [0.22.0-beta](https://github.com/instill-ai/component/compare/v0.21.0-beta...v0.22.0-beta) (2024-07-16)


### Features

* add GitHub component ([#177](https://github.com/instill-ai/component/issues/177)) ([46e5a8e](https://github.com/instill-ai/component/commit/46e5a8e9122d900c3010705b9a0003c7e23a7d41))
* add JQ input field that accepts any type ([#201](https://github.com/instill-ai/component/issues/201)) ([cba4aac](https://github.com/instill-ai/component/commit/cba4aacdda1bca8fe676f299248c04547996c828))
* **cohere:** add Cohere component ([#187](https://github.com/instill-ai/component/issues/187)) ([63fd578](https://github.com/instill-ai/component/commit/63fd57891a10a61f4454816bd39e53cb282ee291))
* **cohere:** add cohere to be able to use instill credit ([#213](https://github.com/instill-ai/component/issues/213)) ([80415b1](https://github.com/instill-ai/component/commit/80415b1467a61348b6d8d32c7199f73de2b6256e))
* GitHub component pagination ([#212](https://github.com/instill-ai/component/issues/212)) ([4b8bbc7](https://github.com/instill-ai/component/commit/4b8bbc7ad39600d115f5f422c69e04181ae497a3))
* **instill:** send requester UID, if present, on model trigger ([#202](https://github.com/instill-ai/component/issues/202)) ([31422cd](https://github.com/instill-ai/component/commit/31422cda00c507e6a53a3f288de16dba2ca9e6cf))
* **mistral:** add Mistral AI component ([#204](https://github.com/instill-ai/component/issues/204)) ([12aaf4f](https://github.com/instill-ai/component/commit/12aaf4f3954c19bf9eab48b8c70459881bdca340))
* **openai:** add dimensions in openai component ([#200](https://github.com/instill-ai/component/issues/200)) ([0d08912](https://github.com/instill-ai/component/commit/0d089121d280e663c36529426e4518411b58f6c2))
* **text:** add input and output and fix bugs ([#209](https://github.com/instill-ai/component/issues/209)) ([56ab3eb](https://github.com/instill-ai/component/commit/56ab3eba4b0286f51caa53d1b3e27aae9113c73b))
* unify pipeline and component usage handlers ([#197](https://github.com/instill-ai/component/issues/197)) ([e27e46c](https://github.com/instill-ai/component/commit/e27e46c0876b68217084ed9d56ea2a77ee081fe2))


### Bug Fixes

* fix instillUpstreamTypes not correctly render the JSON schema ([#216](https://github.com/instill-ai/component/issues/216)) ([bb603bd](https://github.com/instill-ai/component/commit/bb603bd7c74ab6a57ec07158477a99d968cd1c80))
* **mistralai:** svg naming is wrong ([#218](https://github.com/instill-ai/component/issues/218)) ([108817a](https://github.com/instill-ai/component/commit/108817a2a85befbc55b6ed8e39b9a24881804154))
* **text:** hotfix the bug from langchaingo without importing the function o… ([#217](https://github.com/instill-ai/component/issues/217)) ([4cfc263](https://github.com/instill-ai/component/commit/4cfc263154730e0dc04b7fc51c9a27585b947299))
* typo ([#195](https://github.com/instill-ai/component/issues/195)) ([d6b2a42](https://github.com/instill-ai/component/commit/d6b2a42e7f3d0d3cede4ded8c4c973585077f4bd))

## [0.21.0-beta](https://github.com/instill-ai/component/compare/v0.20.2-beta...v0.21.0-beta) (2024-07-02)


### Features

* add mail component ([#178](https://github.com/instill-ai/component/issues/178)) ([04b19d0](https://github.com/instill-ai/component/commit/04b19d0537e8870b1207de6910eb362c517a2eed))
* add read task for gcs ([#155](https://github.com/instill-ai/component/issues/155)) ([77fe2fc](https://github.com/instill-ai/component/commit/77fe2fc60f22bc5121d5835947954af3ba4f7400))
* add read task in bigquery component ([#156](https://github.com/instill-ai/component/issues/156)) ([4d2e7ec](https://github.com/instill-ai/component/commit/4d2e7ecc10904a93ffa32dd4600856dccbba68b7))
* **anthropic:** add Anthropic component ([#176](https://github.com/instill-ai/component/issues/176)) ([030881d](https://github.com/instill-ai/component/commit/030881dac345759cf00c1f33880b9c1398b8f3a9))
* **anthropic:** add UsageHandler functions in anthropic ([#186](https://github.com/instill-ai/component/issues/186)) ([ebaa61f](https://github.com/instill-ai/component/commit/ebaa61f66e1996540fa6b0c4f425408bec70b290))
* **compogen:** add extra section with --extraContents flag' ([#171](https://github.com/instill-ai/component/issues/171)) ([391bb98](https://github.com/instill-ai/component/commit/391bb9850aa6dca207c2c5157beb0f5d6fa011cb))
* **instill:** remove extra-params field ([#188](https://github.com/instill-ai/component/issues/188)) ([b17ff73](https://github.com/instill-ai/component/commit/b17ff73fe350d3785031ec154c37e2f83a352978))
* **redis:** simplify the TLS configuration ([#194](https://github.com/instill-ai/component/issues/194)) ([0a8baf7](https://github.com/instill-ai/component/commit/0a8baf73440e0fef5d2601701f0bcfccd8e5e363))


### Bug Fixes

* **all:** fix typos ([#174](https://github.com/instill-ai/component/issues/174)) ([cb3c2fb](https://github.com/instill-ai/component/commit/cb3c2fbbb7362885c3f763fc5f690e0526b19bc5))
* **compogen:** wrong bracket direction in substitution ([#184](https://github.com/instill-ai/component/issues/184)) ([dfe8306](https://github.com/instill-ai/component/commit/dfe83060f2b5024c72e28340246498d1a553497a))
* expose input and output for anthropic for instill credit ([#190](https://github.com/instill-ai/component/issues/190)) ([a36e876](https://github.com/instill-ai/component/commit/a36e876869b3c2812736b66f56fa4ebe0a9f4985))
* update doc ([#185](https://github.com/instill-ai/component/issues/185)) ([6e6639a](https://github.com/instill-ai/component/commit/6e6639aa59ab0bc8c78c24a596e1393be9cd7db5))

## [0.20.2-beta](https://github.com/instill-ai/component/compare/v0.20.1-beta...v0.20.2-beta) (2024-06-21)


### Bug Fixes

* **openai:** fix OpenAI image_url field can't work ([#170](https://github.com/instill-ai/component/issues/170)) ([f077309](https://github.com/instill-ai/component/commit/f07730960e853163f067ab471aa8326683420c8c))
* **website:** fix typo ([481bc83](https://github.com/instill-ai/component/commit/481bc836e8a3c1b0bfeade93a6cb76b63d95c6bf))

## [0.20.1-beta](https://github.com/instill-ai/component/compare/v0.20.0-beta...v0.20.1-beta) (2024-06-19)


### Bug Fixes

* **redis:** fix typo ([#167](https://github.com/instill-ai/component/issues/167)) ([9dd783a](https://github.com/instill-ai/component/commit/9dd783a1852d273360f55245f258e8d74687c846))

## [0.20.0-beta](https://github.com/instill-ai/component/compare/v0.19.1-beta...v0.20.0-beta) (2024-06-18)


### Features

* add toggle image tag function for document pdf to markdown task ([#162](https://github.com/instill-ai/component/issues/162)) ([f12ecd2](https://github.com/instill-ai/component/commit/f12ecd286699241361e8e81790ad97f6f2707eaf))
* use camelCase in JSON files ([#158](https://github.com/instill-ai/component/issues/158)) ([ecf4dd9](https://github.com/instill-ai/component/commit/ecf4dd906133ba3f89792bd91e7bcc19ff48a40b))
* use kebab-case for all component input and output properties ([#164](https://github.com/instill-ai/component/issues/164)) ([4a82be7](https://github.com/instill-ai/component/commit/4a82be7a0e4ba011db9997af9f2209735f5f0b61))

## [0.19.1-beta](https://github.com/instill-ai/component/compare/v0.19.0-beta...v0.19.1-beta) (2024-06-13)


### Bug Fixes

* fix component document links ([#159](https://github.com/instill-ai/component/issues/159)) ([fd38a8b](https://github.com/instill-ai/component/commit/fd38a8bfa2954c7818d79e94b7658c5ad7f2c4b5))

## [0.19.0-beta](https://github.com/instill-ai/component/compare/v0.18.0-beta...v0.19.0-beta) (2024-06-05)


### Features

* add pdf component ([#138](https://github.com/instill-ai/component/issues/138)) ([517afcf](https://github.com/instill-ai/component/commit/517afcffaff085864b34568de43425ae7c5e7fc0))
* add task chunk text ([#139](https://github.com/instill-ai/component/issues/139)) ([7b36553](https://github.com/instill-ai/component/commit/7b365537a95494b5d8ec6889fa84e1638738a0a5))
* **instill:** adopt latest Model endpoints ([#146](https://github.com/instill-ai/component/issues/146)) ([7f2537b](https://github.com/instill-ai/component/commit/7f2537baa8258811db8d25b7ce2e270a4392c2d3))
* optimise ux for slack component ([#143](https://github.com/instill-ai/component/issues/143)) ([ed60235](https://github.com/instill-ai/component/commit/ed60235d4f734fc8f0f1713fdba97e1b96c645d5))
* refactor package structure ([#140](https://github.com/instill-ai/component/issues/140)) ([4853d4c](https://github.com/instill-ai/component/commit/4853d4cdb12ffe238a97c6d62e653b91cd3e1311))
* support markdown to text function in text operator ([#149](https://github.com/instill-ai/component/issues/149)) ([dcbae37](https://github.com/instill-ai/component/commit/dcbae3711d65a70301cd634e21bd9786fe508aa7))
* unify component interface ([#144](https://github.com/instill-ai/component/issues/144)) ([ad35e10](https://github.com/instill-ai/component/commit/ad35e10384ba6eb1e7fe4997de0bbe5a6e2111a4))


### Bug Fixes

* bug of failure of document component ([#152](https://github.com/instill-ai/component/issues/152)) ([aed51f8](https://github.com/instill-ai/component/commit/aed51f883c7b5fc244159a34c57cf3a608106b50))

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
