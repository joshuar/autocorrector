# Changelog

## [3.1.4](https://github.com/joshuar/autocorrector/compare/v3.1.3...v3.1.4) (2023-10-24)


### Bug Fixes

* **app,db,keytracker:** better handling of app shutdown ([719ae12](https://github.com/joshuar/autocorrector/commit/719ae127ceb3536c54fd20a6803b5ee294114f6c))
* **app:** shutdown works correctly from tray menu ([2eff060](https://github.com/joshuar/autocorrector/commit/2eff060f722814d2cd288b283dfec16c47eef2c3))

## [3.1.3](https://github.com/joshuar/autocorrector/compare/v3.1.2...v3.1.3) (2023-10-22)


### Bug Fixes

* specify cosign version ([f178ea2](https://github.com/joshuar/autocorrector/commit/f178ea275c1a194ae8c289a6ef3eac7c33802e98))

## [3.1.2](https://github.com/joshuar/autocorrector/compare/v3.1.1...v3.1.2) (2023-10-22)


### Bug Fixes

* install cosign in actions ([ac7ef25](https://github.com/joshuar/autocorrector/commit/ac7ef25bbc1c131583807056c2d5494e05be8df7))

## [3.1.1](https://github.com/joshuar/autocorrector/compare/v3.1.0...v3.1.1) (2023-10-22)


### Bug Fixes

* **ui:** commit URL txt files ([eb07d91](https://github.com/joshuar/autocorrector/commit/eb07d915ce4e79cad6d0f1d0383d8d174e2f516c))

## [3.1.0](https://github.com/joshuar/autocorrector/compare/v3.0.1...v3.1.0) (2023-10-22)


### Features

* **db:** sync database to disk every hour ([c06af4e](https://github.com/joshuar/autocorrector/commit/c06af4eb3356d33ca14cd2270a53ab9e814329e9))


### Bug Fixes

* **keytracker:** remove non-working pattern matching code ([e495f1b](https://github.com/joshuar/autocorrector/commit/e495f1b7932273f2b9f8557eee121267898d238c))

## [3.0.1](https://github.com/joshuar/autocorrector/compare/v3.0.0...v3.0.1) (2023-10-18)


### Miscellaneous Chores

* release 3.0.1 packaging fixes ([5c4ff60](https://github.com/joshuar/autocorrector/commit/5c4ff60afaaa347224940c451845f3f2aa91fc4a))

## [3.0.0](https://github.com/joshuar/autocorrector/compare/v2.1.1...v3.0.0) (2023-10-18)


### ⚠ BREAKING CHANGES

* **app,keytracker:** restructure corrections code
* **db,app,corrections:** improve corrections tracking
* **all:** improve stats tracking

### Features

* **all:** improve stats tracking ([35656f8](https://github.com/joshuar/autocorrector/commit/35656f84684e7076509baab3b0b81b4fdb4bae90))
* **app,keytracker:** restructure corrections code ([b6ae8da](https://github.com/joshuar/autocorrector/commit/b6ae8daa5b6a16de960145a489212dc11bd9aa4f))
* **app:** embed icons and urls rather than defining in code ([4f7f82d](https://github.com/joshuar/autocorrector/commit/4f7f82deaf481f14c91a53c65d7a56ca23e68edb))
* **db,app,corrections:** improve corrections tracking ([1e811de](https://github.com/joshuar/autocorrector/commit/1e811de89dcce5eb5628ccfa89c2abbd63c36f3f))
* **keytracker:** adjust interfaces for keytracker ([6aa3d34](https://github.com/joshuar/autocorrector/commit/6aa3d3416ec94eb560d43b808a26fca0dfbc49c5))
* **keytracker:** improve goroutine handling ([c6177d6](https://github.com/joshuar/autocorrector/commit/c6177d6024e6ed5a309856625a76123a478f827f))
* **tools/scraper:** add a tool to scrape common misspellings off Wikipedia ([9b412ec](https://github.com/joshuar/autocorrector/commit/9b412ecf49b009360e93012ea4ead957835b533e))
* **tools:** scraper output saves as corrections.toml ([ff32e41](https://github.com/joshuar/autocorrector/commit/ff32e41db1b4596ca9fca71ee9cac1f5b1d17817))


### Bug Fixes

* **app:** quit from tray icon runs appropriate actions ([6681999](https://github.com/joshuar/autocorrector/commit/66819995b8d737740ef7f5ab9ea0027d5f4c3df6))
* **corrections:** remove spew ([a6a96fd](https://github.com/joshuar/autocorrector/commit/a6a96fdebd529c02d4b36116afc3560278929c12))

## [2.1.1](https://github.com/joshuar/autocorrector/compare/v2.1.0...v2.1.1) (2023-08-26)


### Bug Fixes

* **app:** auto-complete mistake ([696a54a](https://github.com/joshuar/autocorrector/commit/696a54a31b788d324715f3bfc039bf96b86ce57b))

## [2.1.0](https://github.com/joshuar/autocorrector/compare/v2.0.0...v2.1.0) (2023-08-25)


### Features

* **app,wordstats:** start tracking session stats ([091d823](https://github.com/joshuar/autocorrector/commit/091d823941cd683e875f8c31aec5eee7345ccc99))
* **app:** better layout and label for lifetime stats ([84a77c9](https://github.com/joshuar/autocorrector/commit/84a77c9c97260f9e4df3065b3b852b5652814585))

## [2.0.0](https://github.com/joshuar/autocorrector/compare/v1.1.2...v2.0.0) (2023-06-26)


### ⚠ BREAKING CHANGES

* working code without sockets

### Features

* **app,keytracker:** split out channel handling and word details from keytracker ([998e156](https://github.com/joshuar/autocorrector/commit/998e156a8efbc3f4fc4134df227556a80b862293))
* **app:** add settings, report issue, request feature tray menu options ([74377bc](https://github.com/joshuar/autocorrector/commit/74377bcd34e2e6fd846784c9cae05eda41d4cb86))
* working code without sockets ([80b2028](https://github.com/joshuar/autocorrector/commit/80b2028c49b937a1e7d55157f50b992e553f9937))


### Bug Fixes

* **app,keytracker:** corrections can now be toggle on/off again ([c560ad8](https://github.com/joshuar/autocorrector/commit/c560ad8ab996d0a5e125a474f1da542becce91bc))
* **app:** "Show Stats" tray icon menu option restored ([878e12b](https://github.com/joshuar/autocorrector/commit/878e12b6505d35d6470f919369db9071d59bed44))
* **app:** notifications toggle and display working again ([03e00f5](https://github.com/joshuar/autocorrector/commit/03e00f5d2773f22bc4e9fb27aa8fbaa3a31f14c6))
* **app:** remove unused notifications code ([3c01d3b](https://github.com/joshuar/autocorrector/commit/3c01d3bd0ee410eecdde36acb85f3f981ce17dfa))
* **app:** stats tracking now working again ([8486b27](https://github.com/joshuar/autocorrector/commit/8486b27c0ae2b14dd3dce0e67cb16d49c88b01cd))
* **cmd,app,server:** merge client command into root command ([aa260aa](https://github.com/joshuar/autocorrector/commit/aa260aa1d577a0126157677931b1cef6c49166f8))
* **cmd:** remove `--user` command option and `enable` sub-command ([35f0dee](https://github.com/joshuar/autocorrector/commit/35f0dee26964a49f5a223a4832e3723dcdf94091))
* **notifications:** remove more unused notifications code ([c582389](https://github.com/joshuar/autocorrector/commit/c58238908c5524190782ed053130315ab47e61cd))
* remove no longer used client and control code ([3105d1f](https://github.com/joshuar/autocorrector/commit/3105d1f9dd85e3609e17c6fc9334e383d7c13081))

## [1.1.2](https://github.com/joshuar/autocorrector/compare/v1.1.1...v1.1.2) (2023-06-04)


### Bug Fixes

* **assets:** incorrect OnlyShowIn value removed ([0b5d6f4](https://github.com/joshuar/autocorrector/commit/0b5d6f4ac9a3e4b6dfa36d6978112c10f2bf3fe5))

## [1.1.1](https://github.com/joshuar/autocorrector/compare/v0.4.9...v1.1.1) (2023-06-04)


### ⚠ BREAKING CHANGES

* migrate to fyne for UI elements (tray icon)

### Features

* **app:** rework icons for different app states ([b171502](https://github.com/joshuar/autocorrector/commit/b171502f866d9012d7f7e94c15431abdc4dc919a))
* **client:** migrate from logrus to zerolog ([946461f](https://github.com/joshuar/autocorrector/commit/946461f1968fc17e95e4af1e8ab863fc2c0734ed))
* **client:** migrate notifications to fyne notifications ([d013afe](https://github.com/joshuar/autocorrector/commit/d013afee796f542628c6ebce5d6ac06f5ccaeb2b))
* **cmd:** create some common functions used by all commands ([f6b4efb](https://github.com/joshuar/autocorrector/commit/f6b4efbadb2b5412fd4670b63019e8e564eef950))
* **cmd:** enable flexible port selection for profiling ([406f25a](https://github.com/joshuar/autocorrector/commit/406f25ab7e0ea835bbac671e6c1e505c1e1a5cdc))
* **cmd:** migrate from logrus to zerolog ([81fa6f1](https://github.com/joshuar/autocorrector/commit/81fa6f17cdfa9bd865072e4c0a0593018bda26d4))
* **cmd:** use functions for setting debugging/profiling and checking permissions ([cc5da36](https://github.com/joshuar/autocorrector/commit/cc5da366458bfb57414f66d3e42682ade894a111))
* migrate to fyne for UI elements (tray icon) ([1ac82ef](https://github.com/joshuar/autocorrector/commit/1ac82ef34a65b94d11197016249fab6636389631))
* remove all usage of logrus in favour of zerolog ([5a6b8c4](https://github.com/joshuar/autocorrector/commit/5a6b8c4e191fcf05a32afa118739725ea9272f1a))
* **repo:** add bug report and feature request GitHub issue templates ([d698473](https://github.com/joshuar/autocorrector/commit/d698473ccff2083f0de7ed197726d0a213a9719d))
* **server:** migrate from logrus to zerolog ([c44f291](https://github.com/joshuar/autocorrector/commit/c44f2910115b522a86879d52b3231441beaa453d))


### Bug Fixes

* **assets:** .desktop file validation ([4d4252c](https://github.com/joshuar/autocorrector/commit/4d4252c27eccfefece2aae70f1fbebf102368405))
* **build:** fix dependency error for rpm ([c5ef320](https://github.com/joshuar/autocorrector/commit/c5ef320b46f2b9036024df65b2952fa4d184e7e7))
* **client:** file naming and go.mod deps ([9c5389e](https://github.com/joshuar/autocorrector/commit/9c5389ec657172fb06ae27ee8258394251b93f41))
* **client:** fix logic for client start ([835cc8e](https://github.com/joshuar/autocorrector/commit/835cc8e41eedf351d8b0bbd3ee8839bbb97b1501))
* **client:** recover when server disconnects ([b12b9c1](https://github.com/joshuar/autocorrector/commit/b12b9c1b1ea5f2732eed3f1b7fa460082f2763f1))
* **client:** remove deprecated io/ioutil usage ([9e8de7e](https://github.com/joshuar/autocorrector/commit/9e8de7e805845237257db64a71eb7f3504856f8b))
* **cmd:** change default paths for installation ([9d25c1c](https://github.com/joshuar/autocorrector/commit/9d25c1c66cd3489486c9d7a23817b12a40611341))
* **keytracker:** missing parameter for log message ([170c3c1](https://github.com/joshuar/autocorrector/commit/170c3c1ed6b1fa080f1014d6981c9345b5c8e4b5))
* **keytracker:** update logic for creating new virtual keyboard ([f5acce2](https://github.com/joshuar/autocorrector/commit/f5acce28d2929a82d4668e920585b38d7269d668))
* **server:** don't try to send corrections if no client connected ([65f02d5](https://github.com/joshuar/autocorrector/commit/65f02d55a50ea7abf1c390d74845983ba5d7da4f))


### Miscellaneous Chores

* release 1.1.1 ([0f5e8d1](https://github.com/joshuar/autocorrector/commit/0f5e8d11c609c2b373e7cf0a2058003e7727e2dc))

## [1.1.0](https://github.com/joshuar/autocorrector/compare/v1.0.1...v1.1.0) (2023-06-04)


### Features

* **app:** rework icons for different app states ([f2da747](https://github.com/joshuar/autocorrector/commit/f2da747c3b7716213bed863a38291039db71a2cb))
* **repo:** add bug report and feature request GitHub issue templates ([cb5a7c3](https://github.com/joshuar/autocorrector/commit/cb5a7c316631d1d9a0c719e4b2e11684a8509eda))


### Bug Fixes

* **assets:** .desktop file validation ([6c49dfe](https://github.com/joshuar/autocorrector/commit/6c49dfe3a7fcab80e0af76ff8d1e92202b352102))
* **cmd:** change default paths for installation ([5d8b549](https://github.com/joshuar/autocorrector/commit/5d8b549635ef79a4bf6c07a0523e264076518b22))

## [1.0.1](https://github.com/joshuar/autocorrector/compare/v1.0.0...v1.0.1) (2023-05-05)


### Bug Fixes

* **build:** fix dependency error for rpm ([aa8917a](https://github.com/joshuar/autocorrector/commit/aa8917a581127b58415c6511d9fa0e537ce4e3d2))

## [1.0.0](https://github.com/joshuar/autocorrector/compare/v0.4.9...v1.0.0) (2023-05-05)


### ⚠ BREAKING CHANGES

* migrate to fyne for UI elements (tray icon)

### Features

* **client:** migrate from logrus to zerolog ([946461f](https://github.com/joshuar/autocorrector/commit/946461f1968fc17e95e4af1e8ab863fc2c0734ed))
* **client:** migrate notifications to fyne notifications ([d013afe](https://github.com/joshuar/autocorrector/commit/d013afee796f542628c6ebce5d6ac06f5ccaeb2b))
* **cmd:** create some common functions used by all commands ([f6b4efb](https://github.com/joshuar/autocorrector/commit/f6b4efbadb2b5412fd4670b63019e8e564eef950))
* **cmd:** enable flexible port selection for profiling ([406f25a](https://github.com/joshuar/autocorrector/commit/406f25ab7e0ea835bbac671e6c1e505c1e1a5cdc))
* **cmd:** migrate from logrus to zerolog ([81fa6f1](https://github.com/joshuar/autocorrector/commit/81fa6f17cdfa9bd865072e4c0a0593018bda26d4))
* **cmd:** use functions for setting debugging/profiling and checking permissions ([cc5da36](https://github.com/joshuar/autocorrector/commit/cc5da366458bfb57414f66d3e42682ade894a111))
* migrate to fyne for UI elements (tray icon) ([1ac82ef](https://github.com/joshuar/autocorrector/commit/1ac82ef34a65b94d11197016249fab6636389631))
* remove all usage of logrus in favour of zerolog ([5a6b8c4](https://github.com/joshuar/autocorrector/commit/5a6b8c4e191fcf05a32afa118739725ea9272f1a))
* **server:** migrate from logrus to zerolog ([c44f291](https://github.com/joshuar/autocorrector/commit/c44f2910115b522a86879d52b3231441beaa453d))


### Bug Fixes

* **client:** file naming and go.mod deps ([9c5389e](https://github.com/joshuar/autocorrector/commit/9c5389ec657172fb06ae27ee8258394251b93f41))
* **client:** fix logic for client start ([835cc8e](https://github.com/joshuar/autocorrector/commit/835cc8e41eedf351d8b0bbd3ee8839bbb97b1501))
* **client:** recover when server disconnects ([b12b9c1](https://github.com/joshuar/autocorrector/commit/b12b9c1b1ea5f2732eed3f1b7fa460082f2763f1))
* **client:** remove deprecated io/ioutil usage ([9e8de7e](https://github.com/joshuar/autocorrector/commit/9e8de7e805845237257db64a71eb7f3504856f8b))
* **keytracker:** update logic for creating new virtual keyboard ([f5acce2](https://github.com/joshuar/autocorrector/commit/f5acce28d2929a82d4668e920585b38d7269d668))
* **server:** don't try to send corrections if no client connected ([65f02d5](https://github.com/joshuar/autocorrector/commit/65f02d55a50ea7abf1c390d74845983ba5d7da4f))
