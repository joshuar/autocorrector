# Usage

## How it works

- Autocorrector works by detecting when a word has been entered, checking this
  word against a list of corrections and applying a correction if found.
- Word detection is based on looking for a series of alphanumeric characters
  following by a punctuation character. Once the punctuation character is
  detected, Autocorrector looks for and if found, replaces typos.
- This means Autocorrector has some caveats:
  - Autocorrector is not a generic pattern replacement tool; it can only replace
    one word with another word. If you need generic pattern replacement, for
    example, replacing a sequence of characters with a different sequence of
    characters, you will need another tool.
  - Autocorrector can potentially work as a text expander, as long as the
    pattern to trigger expansion ends with a punctuation character. However,
    this is not its primary goal.

## Managing corrections

- Autocorrector looks for a list of corrections in a `corrections.toml` file
  located in one of the following places:
  - `$HOME/.config/autocorrector/corrections.toml` (does not exist by default)
  - `/usr/share/autocorrector/corrections.toml`
- This file is [TOML formatted](https://toml.io/en/).
- The default list (`/usr/share/autocorrector/corrections.toml`) is
  machine-generated from [Wikipedia's list of common
  mispellings](https://en.wikipedia.org/wiki/Wikipedia:Lists_of_common_misspellings).
  As it is machine-generated, there may be some unwanted or unexpected
  corrections. Some cleaning of the list is done. The code for generating the
  default list can be found in the `tools/scraper` directory of the source
  code repository.
- You can add/remove corrections by copying the default list to
  `$HOME/.config/autocorrector/corrections.toml` and editing the file. Autocorrector, if running,
  will pick up the changes automatically.

## Other features

### Temporarily disable autocorrector

- You can temporarily disable autocorrector through the *Toggle Corrections*
  option in the tray icon menu.

### Show corrections as they are made

- You can get a notification when Autocorrector makes a correction by toggling
  the *Toggle Notifications* option in the tray icon menu.

### Show statistics

- Simple statistics on Autocorrector usage can be displayed in the tray icon
  menu with the *Show Stats* option.
