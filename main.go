package main

import (
	"fmt"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/joshuar/autocorrector/cmd"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/joshuar/autocorrector/internal/wordstats"

	log "github.com/sirupsen/logrus"
)

var (
	keyTracker *keytracker.KeyTracker
	wordStats  *wordstats.WordStats
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	cmd.Execute()
	keyTracker = keytracker.NewKeyTracker()
	wordStats = wordstats.OpenWordStats()

	systray.SetIcon(icon.Data)
	systray.SetTitle("Autocorrector")
	systray.SetTooltip("Autocorrector corrects your typos")
	mCorrections := systray.AddMenuItemCheckbox("Show Corrections", "Show corrections as they happen", false)
	mEnabled := systray.AddMenuItemCheckbox("Enabled", "Enable Autocorrector", true)
	mStats := systray.AddMenuItem("Stats", "Show current stats")
	mQuit := systray.AddMenuItem("Quit", "Quit Autocorrector")

	go keyTracker.SlurpWords(wordStats)
	go keyTracker.SnoopKeys()

	for {
		select {
		case <-mEnabled.ClickedCh:
			if mEnabled.Checked() {
				mEnabled.Uncheck()
				keyTracker.Disabled = true
				log.Info("Disabling Autocorrector")
				beeep.Notify("Autocorrector disabled", "Temporarily disabling autocorrector", "")
			} else {
				mEnabled.Check()
				keyTracker.Disabled = false
				log.Info("Enabling Autocorrector")
				beeep.Notify("Autocorrector enabled", "Re-enabling autocorrector", "")

			}
		case <-mCorrections.ClickedCh:
			if mCorrections.Checked() {
				mCorrections.Uncheck()
				keyTracker.ShowCorrections = false
				beeep.Notify("Hiding Corrections", "Hiding notifications for corrections", "")
			} else {
				mCorrections.Check()
				keyTracker.ShowCorrections = true
				beeep.Notify("Showing Corrections", "Notifications for corrections will be shown as they are made", "")

			}
		case <-mQuit.ClickedCh:
			log.Info("Requesting quit")
			systray.Quit()
		case <-mStats.ClickedCh:
			beeep.Notify("Current Stats",
				fmt.Sprintf("%v words checked.\n%v words corrected.\n%.2f %% accuracy.",
					wordStats.GetCheckedTotal(),
					wordStats.GetCorrectedTotal(),
					wordStats.CalcAccuracy()),
				"")
		}
	}
}

func onExit() {
	wordStats.CloseWordStats()
	keyTracker.CloseKeyTracker()
}
