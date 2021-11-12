package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/getlantern/systray"
	"github.com/joshuar/autocorrector/assets/icon"
	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/corrections"
	"github.com/joshuar/autocorrector/internal/notifications"
	"github.com/joshuar/autocorrector/internal/wordstats"
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debugFlag {
				log.SetLevel(log.DebugLevel)
			}
			if profileFlag {
				go func() {
					log.Info(http.ListenAndServe("localhost:6061", nil))
				}()
				log.Info("Profiling is enabled and available at localhost:6061")
			}
			if correctionsFlag != "" {
				log.Debug("Using config file specified on command-line: ", correctionsFlag)
				viper.SetConfigFile(correctionsFlag)
			} else {
				viper.SetConfigName("corrections")
				viper.SetConfigType("toml")
				viper.AddConfigPath(xdg.ConfigHome + "/autocorrector")
				viper.AddConfigPath("/usr/local/share/autocorrector")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			systray.Run(onReady, onExit)
		},
	}
	clientSetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Set-up an autocorrector client",
		Long:  "The client set-up command will make a copy of the default corrections file and create an autostart entry for the current user",
		Run: func(cmd *cobra.Command, args []string) {
			err := os.Mkdir(xdg.ConfigHome+"/autocorrector", 0755)
			if e, ok := err.(*os.PathError); ok && e.Err == syscall.EEXIST {
				log.Warn(err)
			} else {
				log.Fatal(err)
			}
			defaultCorrections, err := ioutil.ReadFile("/usr/local/share/autocorrector/corrections.toml")
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile(xdg.ConfigHome+"/autocorrector/corrections.toml", defaultCorrections, 0755)
			if err != nil {
				log.Fatal(err)
			}
			err = os.Symlink("/usr/local/share/applications/autocorrector.desktop", xdg.ConfigHome+"/autostart/autocorrector.desktop")
			if e, ok := err.(*os.LinkError); ok && e.Err == syscall.EEXIST {
				log.Warn(err)
			} else {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	clientCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
	clientCmd.Flags().StringVar(&correctionsFlag, "corrections", "", fmt.Sprintf("list of corrections (default is %s/autocorrector/corrections.toml)", xdg.ConfigHome))
	clientCmd.AddCommand(clientSetupCmd)
}

func onReady() {
	socket := control.ConnectSocket()
	go socket.RecvData()

	notify := notifications.NewNotificationsHandler()

	stats := wordstats.OpenWordStats()
	corrections := corrections.NewCorrections()

	log.Debug("Client has started, asking server to resume tracking keys")
	socket.ResumeServer()

	go func() {
		for msg := range socket.Data {
			// case: recieved data on the socket
			switch t := msg.(type) {
			case *control.WordMsg:
				stats.AddChecked(t.Word)
				if t.Correction = corrections.FindCorrection(t.Word); t.Correction != "" {
					socket.SendWord(t.Word, t.Correction, t.Punct)
					stats.AddCorrected(t.Word, t.Correction)
					notify.Send("Correction!", fmt.Sprintf("Corrected %s with %s", t.Word, t.Correction))
				}
			default:
				log.Debugf("Unknown message received: %v", msg)
			}
		}
	}()

	go func() {
		systray.SetIcon(icon.Default)
		systray.SetTooltip("Autocorrector corrects your typos")
		mCorrections := systray.AddMenuItemCheckbox("Show Corrections", "Show corrections as they happen", false)
		mEnabled := systray.AddMenuItemCheckbox("Enabled", "Enable autocorrector", true)
		mStats := systray.AddMenuItem("Show Stats", "Updates current stats")
		mStatsDisplay := mStats.AddSubMenuItem("Stats", "Latest stats grab")
		mEdit := systray.AddMenuItem("Edit", "Edit the list of corrections")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit autocorrector")

		for {
			select {
			case <-mEnabled.ClickedCh:
				if mEnabled.Checked() {
					mEnabled.Uncheck()
					socket.PauseServer()
					systray.SetIcon(icon.Disabled)
				} else {
					mEnabled.Check()
					socket.ResumeServer()
					systray.SetIcon(icon.Default)
				}
			case <-mCorrections.ClickedCh:
				if mCorrections.Checked() {
					mCorrections.Uncheck()
					systray.SetIcon(icon.Default)
					notify.ShowNotifications = false
				} else {
					mCorrections.Check()
					notify.ShowNotifications = true
					notify.Send("Showing Corrections", "Notifications for corrections will be shown as they are made")
					systray.SetIcon(icon.Notifying)
				}
			case <-mQuit.ClickedCh:
				log.Info("Requesting quit")
				socket.PauseServer()
				systray.Quit()
			case <-mStats.ClickedCh:
				mStatsDisplay.SetTitle(stats.GetStats())
			case <-mEdit.ClickedCh:
				cmd := exec.Command("xdg-open", viper.ConfigFileUsed())
				if err := cmd.Run(); err != nil {
					log.Error(err)
				}
			case <-socket.Done:
				log.Debug("Received done, restarting socket...")
				socket = control.ConnectSocket()
				go socket.RecvData()
				socket.ResumeServer()
			}
		}
	}()
}

func onExit() {
}
