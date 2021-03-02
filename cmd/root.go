package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/joshuar/autocorrector/internal/icon"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/joshuar/autocorrector/internal/wordstats"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	keyTracker  *keytracker.KeyTracker
	wordStats   *wordstats.WordStats
	cfgFile     string
	debugFlag   bool
	cpuProfile  string
	memProfile  string
	profileFlag bool
	rootCmd     = &cobra.Command{
		Use:   "autocorrector",
		Short: "Autocorrect typos and spelling mistakes.",
		Long:  `Autocorrector is a tool similar to the word replacement functionality in Autokey or AutoHotKey.`,
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// init defines flags and configuration settings
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/autocorrector/autocorrector.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	rootCmd.PersistentFlags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
}

// initConfig reads in config file
func initConfig() {
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal finding home directory: %s", err))
		os.Exit(1)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		log.Debug("Using config file specified on command-line: ", cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		var cfgFileDefault = strings.Join([]string{home, "/.config/autocorrector/autocorrector.toml"}, "")
		viper.SetConfigFile(cfgFileDefault)
		log.Debug("Using default config file: ", cfgFileDefault)
	}
}

func onReady() {
	keyTracker = keytracker.NewKeyTracker()
	wordStats = wordstats.OpenWordStats()

	systray.SetIcon(icon.Data)
	systray.SetTitle("Autocorrector")
	systray.SetTooltip("Autocorrector corrects your typos")
	mCorrections := systray.AddMenuItemCheckbox("Show Corrections", "Show corrections as they happen", false)
	mEnabled := systray.AddMenuItemCheckbox("Enabled", "Enable Autocorrector", true)
	mStats := systray.AddMenuItem("Stats", "Show current stats")
	mQuit := systray.AddMenuItem("Quit", "Quit Autocorrector")

	go keyTracker.SlurpWord(wordStats)
	go keyTracker.SnoopKeys()
	keyTracker.StartSnooping <- true

	for {
		select {
		case <-mEnabled.ClickedCh:
			if mEnabled.Checked() {
				mEnabled.Uncheck()
				keyTracker.StopSnooping <- true
				log.Info("Disabling Autocorrector")
				beeep.Notify("Autocorrector disabled", "Temporarily disabling autocorrector", "")
			} else {
				mEnabled.Check()
				keyTracker.StartSnooping <- true
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
