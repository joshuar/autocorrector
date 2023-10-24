# Autocorrector

![MIT](https://img.shields.io/github/license/joshuar/autocorrector)
![GitHub last commit](https://img.shields.io/github/last-commit/joshuar/autocorrector)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuar/autocorrector?style=flat-square)](https://goreportcard.com/report/github.com/joshuar/autocorrector)
[![Go Reference](https://pkg.go.dev/badge/github.com/joshuar/autocorrector.svg)](https://pkg.go.dev/github.com/joshuar/autocorrector)
[![Release](https://img.shields.io/github/release/joshuar/autocorrector.svg?style=flat-square)](https://github.com/joshuar/autocorrector/releases/latest)

## ❓ What is it?

Autocorrector is a tool similar to Autokey or AutoHotKey, but targeted mainly at
word replacements.  I wrote it because all I wanted to do was fix my fat-finger
typos automatically.  That is what autocorrector aims to do.

Autocorrector reads a TOML configuration file of key-values; the key being the
typo and the value being the replacement.  When it detects you have entered a
typo, it helpfully corrects it.

## ⬇️ Installation

> **Note**
> **This program will only run on Linux**

1. Download either the `.rpm` or `.deb` file and install using your package
   manager.
2. Run `autocorrector` or use the **autocorrector** menu entry in your desktop
   environment.

Package signatures can be verified with
[cosign](https://github.com/sigstore/cosign). To verify a package, you'll need
the [cosign.pub](cosign.pub) public key and the `.sig` file (downloaded from
[releases](https://github.com/joshuar/autocorrector/releases)) that matches the
package you want to verify. To verify a package, a command similar to the
following for the `rpm` package can be used:

```shell
cosign verify-blob --key cosign.pub --signature autocorrector-*.rpm.sig autocorrector-*.rpm
```

## 📝 Configuration and additional details

See [USAGE](USAGE.md)

## 🧑‍🤝‍🧑 Contributing

I would welcome your contribution! If you find any improvement or issue you want
to fix, feel free to send a pull request!

## 🙌 Acknowledgements

The following Go libraries and tools made autocorrector infinitely easier:

- [gokbd](https://github.com/joshuar/gokbd): library using libevdev to talk to a
  keyboard on Linux. It allows snooping the keys pressed as well as typing out
  keys.
- [fyne](https://fyne.io/): UI toolkit and system tray library.
- [zerolog](https://github.com/rs/zerolog): logging library.
- [Cobra](https://github.com/spf13/cobra): command-line interface.

Check out more awesome Go things at the [Awesome Go
List](https://github.com/avelino/awesome-go):

Icon taken from
[here](https://pixabay.com/vectors/spellcheck-correct-typo-errors-1292780/).

## License

[MIT](LICENSE)
