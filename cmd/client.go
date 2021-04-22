package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
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
					log.Println(http.ListenAndServe("localhost:6061", nil))
				}()
				log.Debug("Profiling is enabled and available at localhost:6061")
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
	clientCmd.Flags().StringVar(&correctionsFlag, "corrections", "", "list of corrections (default is $HOME/.config/autocorrector/corrections.toml)")
}

func onReady() {
	notify := newNotificationsHandler()
	go notify.handleNotifications()

	stats := wordstats.OpenWordStats()

	corrections := newCorrections()

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
				default:
					log.Debugf("Unhandled message recieved: %v", msg)
				}
			case *control.WordMsg:
				stats.AddChecked(t.Word)
				t.Correction = corrections.findCorrection(t.Word)
				if t.Correction != "" {
					manager.SendWord(t.Word, t.Correction, t.Punct)
					stats.AddCorrected(t.Word, t.Correction)
					if notify.showCorrections {
						notify.show("Correction!", fmt.Sprintf("Corrected %s with %s", t.Word, t.Correction))
					}
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
				manager.SendState(&control.StateMsg{Pause: true})
				notify.show("Autocorrector disabled", "Temporarily disabling autocorrector")
			} else {
				mEnabled.Check()
				manager.SendState(&control.StateMsg{Resume: true})
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
			manager.SendState(&control.StateMsg{Pause: true})
			systray.Quit()
		case <-mStats.ClickedCh:
			notify.show("Current Stats", fmt.Sprintf("%v words checked.\n%v words corrected.\n%.2f %% accuracy.",
				stats.GetCheckedTotal(),
				stats.GetCorrectedTotal(),
				stats.CalcAccuracy()))
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
	configMap := make(map[string]string)
	viper.Unmarshal(&configMap)
	for _, v := range configMap {
		found := viper.GetString(v)
		if found != "" {
			log.Fatalf("A replacement in the config is also listed as a typo (%v)  This won't work.", v)
		}
	}
	log.Debug("Config looks okay.")
}

func newCorrections() *corrections {
	log.Debugf("Using corrections config: %s", viper.ConfigFileUsed())
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Could not find config file: ", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Errorf("fatal error config file: %s", err))
		}
	}
	corrections := &corrections{
		correctionList:    make(map[string]string),
		updateCorrections: make(chan bool),
	}
	corrections.checkConfig()
	viper.Unmarshal(&corrections.correctionList)
	go func() {
		for {
			switch {
			case <-corrections.updateCorrections:
				corrections.checkConfig()
				viper.Unmarshal(&corrections.correctionList)
				log.Debug("Updated corrections from config file.")
			}
		}

	}()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debug("Config file has changed.")
		corrections.updateCorrections <- true
	})
	viper.WatchConfig()
	return corrections
}
