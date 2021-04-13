package cmd

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
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
	// statsCmd.Flags().BoolVarP(&logStatsFlag, "log", "l", false, "Show log of corrections")
}

func onReady() {

	// socket := control.newClientSocket()
	// go socket.AcceptConnections()

	letters := []string{"a", "e", "i", "o", "u"}

	for i := range letters {
		SendMessage([]byte(letters[i]))
		time.Sleep(1 * time.Second)
	}

	os.Exit(0)

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
				log.Info("Disabling Autocorrector")
				// keyTracker.Pause = true
				beeep.Notify("Autocorrector disabled", "Temporarily disabling autocorrector", "")
			} else {
				mEnabled.Check()
				log.Info("Enabling Autocorrector")
				// keyTracker.Pause = false
				beeep.Notify("Autocorrector enabled", "Re-enabling autocorrector", "")

			}
		case <-mCorrections.ClickedCh:
			if mCorrections.Checked() {
				mCorrections.Uncheck()
				// keyTracker.ShowCorrections = false
				beeep.Notify("Hiding Corrections", "Hiding notifications for corrections", "")
			} else {
				mCorrections.Check()
				// keyTracker.ShowCorrections = true
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

func SendMessage(message []byte) {
	user, err := user.Lookup("joshua")
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.Dial("unix", "/tmp/autocorrector"+user.Username+".sock")
	if err != nil {
		log.Errorf("Failed to dial: %s", err)
	}
	defer c.Close()
	count, err := c.Write(message)
	if err != nil {
		log.Errorf("Write error: %s", err)
	}
	log.Infof("Wrote %d bytes", count)
}
