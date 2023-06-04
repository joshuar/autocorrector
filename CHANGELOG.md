# Changelog

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
