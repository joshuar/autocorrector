# How it works

- Autocorrector works by detecting when a word has been entered, checking this word against a list of corrections and applying a correction if found.
- Word detection is based on looking for a series of alphanumeric characters following by a punctuation character. Once the punctuation character is detected, Autocorrector looks for and if found, replaces typos.
- This means Autocorrector has some caveats:
  - Autocorrector is not a generic pattern replacement tool; it can only replace one word with another word. If you need generic pattern replacement, for example, replacing a sequence of characters with a different sequence of characters, you will need another tool.
  - Autocorrector can potentially work as a text expander, as long as the pattern to trigger expansion ends with a punctuation character. However, this is not its primary goal.

# Detailed Installation

1. Download either the `.rpm` or `.deb` file and install using your package manager.
2. Set up the client by running `autocorrector client setup` as your user.
  - This will create `$HOME/.config/autocorrector/corrections.toml` (using the default corrections file).
  - An auto-start entry will be created in `~/.config/autostart`.
3. Set up the server by running `sudo autocorrector enable $USERNAME` (substitute `$USERNAME` for your user name).
  - This will enable a service, `autocorrector@USERNAME` for the specified user.
4. Start the service with `systemctl restart autocorrector@USERNAME`.
5. Run `autocorrector client` or use the **autocorrector** menu entry in your desktop environment.

- Steps **2** and **3** only need to be run once on first install. On upgrades just restart the server and re-run the client (steps **4** and **5**)

# Managing corrections

- Autocorrector looks for a list of corrections in a `corrections.toml` file located in `$HOME/.config/autocorrector/corrections.toml` by default.
- This file is [TOML formatted](https://toml.io/en/).
- A default list of common English typos is provided, sourced from [here](https://www.autohotkey.com/download/AutoCorrect.ahk). 
- You can add/remove corrections from the file and Autocorrector, if running, will pick-up the changes automatically.
- The tray icon menu provides a convenient way to edit the file, select the *Edit* option and the corrections file will open in your preferred text editor for quick editing.

# Other features

## Show corrections as they are made
- You can get a notification when Autocorrector makes a correction by toggling the *Show Corrections* option in the tray icon menu.

## Show statistics
- Simple statistics on Autocorrector usage can be displayed in the tray icon menu with the *Show Stats* option.



