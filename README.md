# Alpha Software!  Use at own risk!

# Autocorrector

## What is it?

Autocorrector is a tool similar to Autokey or AutoHotKey, but targeted mainly at word replacements.  I wrote it because all I really wanted to do is fix my fat-finger typos automagically.  That is what autocorrector aims to do.  

Autocorrector reads a TOML configuration file of key-values; the key being the typo and the value being the replacement.  When it detects you have entered a typo, it helpfully corrects it.

The libraries to do the magic replacement are X11 only on Linux. Untested on other OSes.

## Made possible with

The following Go libraries and tools made autocorrector infinitely easier:

- [Logrus](https://github.com/sirupsen/logrus):  structured logger for Go (golang), completely API compatible with the standard library logger.
- [Viper](https://github.com/spf13/viper): Go configuration with fangs!
- [robotgo](https://github.com/go-vgo/robotgo): Go Native cross-platform GUI system automation. Control the mouse, keyboard and other.

Check out more awesome Go things at the [Awesome Go List](https://github.com/avelino/awesome-go):

The default list of replacements is sourced from the following AutoHotKey script containing common English typos and mispellings:

- https://www.autohotkey.com/download/AutoCorrect.ahk

## Requirements
- Golang

## Installation

1. Download and install in your GOPATH:

```
go get github.com/joshuar/autocorrector
cd $GOPATH/src/github.com/joshuar/autocorrector
go install
```

2. Run the `autocorrector` command from `$GOPATH/bin/autocorrector` or use the provided systemd service file to start it with your user systemd instance.


