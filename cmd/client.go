package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/icon"
	"github.com/joshuar/autocorrector/internal/wordstats"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Client creates tray icon for control and notifications",
		Long:  `With the client running, you can pause correction and see notifications.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if profileFlag {
				go func() {
					log.Println(http.ListenAndServe("localhost:6060", nil))
				}()
				log.Debug("Profiling is enabled and available at localhost:6060")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			systray.Run(onReady, onExit)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	clientCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")

}

func onReady() {
	stats := wordstats.OpenWordStats()

	manager := control.NewConnManager("")
	go manager.Start()
	log.Debug("Client has started, asking server to resume tracking keys")
	manager.SendState(&control.StateMsg{Resume: true})
	go func() {
		for msg := range manager.Data {
			switch t := msg.(type) {
			case *control.StateMsg:
				switch {
				case t.Start:
					log.Debug("Server has started, asking it to resume tracking keys")
					manager.SendState(&control.StateMsg{Resume: true})
				case t.Stop:
					log.Debug("Server has stopped")
				// case msg.Type == control.Notification:
				// 	notificationData := msg.Data.(control.NotificationData)
				// 	beeep.Notify(notificationData.Title, notificationData.Message, "")
				default:
					log.Debugf("Unhandled message recieved: %v", msg)
				}
			case *control.StatsMsg:
				log.Debug("got stats message")
				if t.Correction != "" {
					stats.AddCorrected(t.Word, t.Correction)
				} else {
					stats.AddChecked(t.Word)
				}
			default:
				log.Debugf("Unhandled message recieved: %v", msg)
			}
		}
	}()

	systray.SetIcon(icon.Data)
	systray.SetTitle("Autocorrector")
	systray.SetTooltip("Autocorrector corrects your typos")
	mCorrections := systray.AddMenuItemCheckbox("Show Corrections", "Show corrections as they happen", false)
	mEnabled := systray.AddMenuItemCheckbox("Enabled", "Enable Autocorrector", true)
	mStats := systray.AddMenuItem("Stats", "Show current stats")
	mQuit := systray.AddMenuItem("Quit", "Quit Autocorrector")

	for {
		select {
		case <-mEnabled.ClickedCh:
			if mEnabled.Checked() {
				mEnabled.Uncheck()
				// manager.SendMessage(control.PauseServer, nil)
				beeep.Notify("Autocorrector disabled", "Temporarily disabling autocorrector", "")
			} else {
				mEnabled.Check()
				// manager.SendMessage(control.ResumeServer, nil)
				beeep.Notify("Autocorrector enabled", "Re-enabling autocorrector", "")

			}
		case <-mCorrections.ClickedCh:
			if mCorrections.Checked() {
				mCorrections.Uncheck()
				// manager.SendMessage(control.HideNotifications, nil)
				beeep.Notify("Hiding Corrections", "Hiding notifications for corrections", "")
			} else {
				mCorrections.Check()
				// manager.SendMessage(control.ShowNotifications, nil)
				beeep.Notify("Showing Corrections", "Notifications for corrections will be shown as they are made", "")

			}
		case <-mQuit.ClickedCh:
			log.Info("Requesting quit")
			// manager.SendMessage(control.PauseServer, nil)
			systray.Quit()
		case <-mStats.ClickedCh:
			beeep.Notify("Current Stats", fmt.Sprintf("%v words checked.\n%v words corrected.\n%.2f %% accuracy.",
				stats.GetCheckedTotal(),
				stats.GetCorrectedTotal(),
				stats.CalcAccuracy()), "")
		}
	}
}

func onExit() {
}

// initConfig reads in config file
func initConfig() {
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(fmt.Errorf("fatal finding home directory: %s", err))
		os.Exit(1)
	}

	if correctionsFlag != "" {
		// Use config file from the flag.
		log.Debug("Using config file specified on command-line: ", correctionsFlag)
		viper.SetConfigFile(correctionsFlag)
	} else {
		var cfgFileDefault = strings.Join([]string{home, "/.config/autocorrector/autocorrector.toml"}, "")
		viper.SetConfigFile(cfgFileDefault)
		log.Debug("Using default config file: ", cfgFileDefault)
	}
}
