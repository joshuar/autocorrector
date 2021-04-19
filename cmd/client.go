package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/icon"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	correctionsFlag string
	clientCmd       = &cobra.Command{
		Use:   "client",
		Short: "Client creates tray icon for control and notifications",
		Long:  `With the client running, you can pause correction and see notifications.`,
		Run: func(cmd *cobra.Command, args []string) {
			systray.Run(onReady, onExit)

		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(clientCmd)
	rootCmd.Flags().StringVar(&correctionsFlag, "corrections", "", "list of corrections (default is $HOME/.config/autocorrector/corrections.toml)")
}

func onReady() {

	socket := control.NewSocket("")
	manager := control.NewConnManager()
	go manager.Start()
	go socket.AcceptConnections(manager)
	socket.SendMessage(control.ResumeServer, nil)
	go func() {
		for msg := range socket.Data {
			switch {
			case msg.Type == control.Acknowledge:
				log.Debugf("Got acknowledgement from server: %s", msg.Data.(string))
			case msg.Type == control.ServerStarted:
				log.Debug("Server has started")
				socket.SendMessage(control.ResumeServer, nil)
			case msg.Type == control.ServerStopped:
				log.Debug("Server has stopped")
			case msg.Type == control.Notification:
				notificationData := msg.Data.(control.NotificationData)
				beeep.Notify(notificationData.Title, notificationData.Message, "")
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
				socket.SendMessage(control.PauseServer, nil)
				beeep.Notify("Autocorrector disabled", "Temporarily disabling autocorrector", "")
			} else {
				mEnabled.Check()
				socket.SendMessage(control.ResumeServer, nil)
				beeep.Notify("Autocorrector enabled", "Re-enabling autocorrector", "")

			}
		case <-mCorrections.ClickedCh:
			if mCorrections.Checked() {
				mCorrections.Uncheck()
				socket.SendMessage(control.HideNotifications, nil)
				beeep.Notify("Hiding Corrections", "Hiding notifications for corrections", "")
			} else {
				mCorrections.Check()
				socket.SendMessage(control.ShowNotifications, nil)
				beeep.Notify("Showing Corrections", "Notifications for corrections will be shown as they are made", "")

			}
		case <-mQuit.ClickedCh:
			log.Info("Requesting quit")
			systray.Quit()
		case <-mStats.ClickedCh:
			socket.SendMessage(control.GetStats, nil)
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
