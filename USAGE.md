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
  located in `$HOME/.config/autocorrector/corrections.toml` by default.
- This file is [TOML formatted](https://toml.io/en/).
- A default list of common English typos is provided, sourced from
  [here](https://www.autohotkey.com/download/AutoCorrect.ahk).
- You can add/remove corrections from the file and Autocorrector, if running,
  will pick up the changes automatically.
- The tray icon menu provides a convenient way to edit the file, select the
  *Edit* option and the corrections file will open in your preferred text editor
  for quick editing.

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
