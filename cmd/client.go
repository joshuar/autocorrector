package cmd

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/adrg/xdg"
	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/icon"
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
					log.Debug(http.ListenAndServe("localhost:6061", nil))
				}()
				log.Debug("Profiling is enabled and available at localhost:6061")
			}
			if correctionsFlag != "" {
				log.Debug("Using config file specified on command-line: ", correctionsFlag)
				viper.SetConfigFile(correctionsFlag)
			} else {
				cfgFileDefault, err := xdg.ConfigFile("autocorrector/corrections.toml")
				if err != nil {
					log.Fatal(err)
				}
				viper.SetConfigFile(cfgFileDefault)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			systray.Run(onReady, onExit)
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	clientCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
	clientCmd.Flags().StringVar(&correctionsFlag, "corrections", "", fmt.Sprintf("list of corrections (default is %s/autocorrector/corrections.toml)", xdg.ConfigHome))
}

func onReady() {

	socket := control.ConnectSocket()
	go socket.RecvData()

	notify := newNotificationsHandler()
	go notify.handleNotifications()

	stats := wordstats.OpenWordStats()

	corrections := newCorrections()

	log.Debug("Client has started, asking server to resume tracking keys")
	socket.SendState(control.Resume)

	go func() {
		systray.SetIcon(icon.Data)
		systray.SetTitle("Autocorrector")
		systray.SetTooltip("Autocorrector corrects your typos")
		mCorrections := systray.AddMenuItemCheckbox("Show Corrections", "Show corrections as they happen", false)
		mEnabled := systray.AddMenuItemCheckbox("Enabled", "Enable autocorrector", true)
		mStats := systray.AddMenuItem("Stats", "Show current stats")
		mEdit := systray.AddMenuItem("Edit", "Edit the list of corrections")
		mQuit := systray.AddMenuItem("Quit", "Quit autocorrector")

		for {
			select {
			case <-mEnabled.ClickedCh:
				if mEnabled.Checked() {
					mEnabled.Uncheck()
					socket.SendState(control.Pause)
					notify.show("Autocorrector disabled", "Temporarily disabling autocorrector")
				} else {
					mEnabled.Check()
					socket.SendState(control.Resume)
					notify.show("Autocorrector enabled", "Re-enabling autocorrector")

				}
			case <-mCorrections.ClickedCh:
				if mCorrections.Checked() {
					mCorrections.Uncheck()
					notify.showCorrections = false
					notify.show("Hiding Corrections", "Hiding notifications for corrections")
				} else {
					mCorrections.Check()
					notify.showCorrections = true
					notify.show("Showing Corrections", "Notifications for corrections will be shown as they are made")

				}
			case <-mQuit.ClickedCh:
				log.Info("Requesting quit")
				socket.SendState(control.Pause)
				systray.Quit()
			case <-mStats.ClickedCh:
				notify.show("Current Stats", stats.GetStats())
			case <-mEdit.ClickedCh:
				cmd := exec.Command("xdg-open", viper.ConfigFileUsed())
				if err := cmd.Run(); err != nil {
					log.Error(err)
				}
			}
		}
	}()

	for msg := range socket.Data {
		switch t := msg.(type) {
		case *control.StateMsg:
			switch t.State {
			case control.Start:
				log.Debug("Server has started, asking it to resume tracking keys")
				socket.SendState(control.Resume)
			case control.Stop:
				log.Debug("Server has stopped")
			default:
				log.Debugf("Unhandled state: %v", msg)
			}
		case *control.WordMsg:
			stats.AddChecked(t.Word)
			t.Correction = corrections.findCorrection(t.Word)
			if t.Correction != "" {
				socket.SendWord(t.Word, t.Correction, t.Punct)
				stats.AddCorrected(t.Word, t.Correction)
				if notify.showCorrections {
					notify.show("Correction!", fmt.Sprintf("Corrected %s with %s", t.Word, t.Correction))
				}
			}
		default:
			log.Debugf("Unhandled message recieved: %v", msg)
		}
	}
}

func onExit() {
}

type notificationMsg struct {
	title, message string
}
type notificationsHandler struct {
	showCorrections bool
	notification    chan *notificationMsg
}

func (nh *notificationsHandler) handleNotifications() {
	for n := range nh.notification {
		beeep.Notify(n.title, n.message, "")
	}
}

func (nh *notificationsHandler) show(t string, m string) {
	n := &notificationMsg{
		title:   t,
		message: m,
	}
	nh.notification <- n
}

func newNotificationsHandler() *notificationsHandler {
	return &notificationsHandler{
		showCorrections: false,
		notification:    make(chan *notificationMsg),
	}
}

type corrections struct {
	correctionList    map[string]string
	updateCorrections chan bool
}

func (c *corrections) findCorrection(mispelling string) string {
	return c.correctionList[mispelling]
}

func (c *corrections) checkConfig() {
	// check if any value is also a key
	// in this case, we'd end up with replacing the typo then replacing the replacement
	c.correctionList = make(map[string]string)
	viper.Unmarshal(&c.correctionList)
	for _, v := range c.correctionList {
		found := viper.GetString(v)
		if found != "" {
			log.Warnf("A replacement (%s) in the config is also listed as a typo. Deleting it to avoid recursive error.", v)
			delete(c.correctionList, found)
		}
	}
	log.Debug("Config looks okay.")
}

func newCorrections() *corrections {
	log.Debugf("Using corrections config at %s", viper.ConfigFileUsed())
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Could not find config file: ", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Errorf("fatal error config file: %s", err))
		}
	}
	corrections := &corrections{
		updateCorrections: make(chan bool),
	}
	corrections.checkConfig()
	go func() {
		for {
			switch {
			case <-corrections.updateCorrections:
				corrections.checkConfig()
			}
		}

	}()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debugf("Config file %s has changed, getting updates.", viper.ConfigFileUsed())
		corrections.updateCorrections <- true
	})
	viper.WatchConfig()
	return corrections
}
