
![Apache 2.0](https://img.shields.io/github/license/joshuar/autocorrector) 
![GitHub last commit](https://img.shields.io/github/last-commit/joshuar/autocorrector)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuar/autocorrector?style=flat-square)](https://goreportcard.com/report/github.com/joshuar/autocorrector) 
[![Go Reference](https://pkg.go.dev/badge/github.com/joshuar/autocorrector.svg)](https://pkg.go.dev/github.com/joshuar/autocorrector)
[![Release](https://img.shields.io/github/release/joshuar/autocorrector.svg?style=flat-square)](https://github.com/joshuar/autocorrector/releases/latest)

# Autocorrector

## What is it?

Autocorrector is a tool similar to Autokey or AutoHotKey, but targeted mainly at word replacements.  I wrote it because all I wanted to do was fix my fat-finger typos automatically.  That is what autocorrector aims to do.  

Autocorrector reads a TOML configuration file of key-values; the key being the typo and the value being the replacement.  When it detects you have entered a typo, it helpfully corrects it.

**This program will only run on Linux**
## Made possible with

The following Go libraries and tools made autocorrector infinitely easier:

- [Logrus](https://github.com/sirupsen/logrus):  structured logger for Go (golang), completely API compatible with the standard library logger.
- [Viper](https://github.com/spf13/viper): Go configuration with fangs!
- [Cobra](https://github.com/spf13/cobra): A Commander for modern Go CLI interactions.
- [systray](https://github.com/getlantern/systray): A cross-platform Go library to place an icon and menu in the notification area.
- [beeep](https://github.com/gen2brain/beeep): a cross-platform library for sending desktop notifications and beeps.
- [bbolt](https://github.com/etcd-io/bbolt): a pure Go key/value store.

Check out more awesome Go things at the [Awesome Go List](https://github.com/avelino/awesome-go):

The default list of replacements is sourced from the following AutoHotKey script containing common English typos and misspellings:

- https://www.autohotkey.com/download/AutoCorrect.ahk

Icon taken from [here](https://pixabay.com/vectors/spellcheck-correct-typo-errors-1292780/).

## Requirements
- Golang

## Installation (for rpm/deb)

1. Download either the `.rpm` or `.deb` file and install using your package manager.
2. Setup the client by running `autocorrector client setup` as your user.
  - This will create `$HOME/.config/autocorrector/corrections.toml` (using the default corrections file).
  - An auto-start entry will be created in `~/.config/autostart`.
3. Setup the server by running `sudo autocorrector enable $USERNAME` (substitute `$USERNAME` for your user name).
  - This will enable a service, `autocorrector@USERNAME` for the specified user.
4. Start the service with `systemctl start autocorrector@USERNAME`.
5. Run `autocorrector client` or use the **autocorrector** menu entry in your desktop environment.
