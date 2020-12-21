package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/joshuar/autocorrector/internal/wordstats"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	keyTracker *keytracker.KeyTracker
	wordStats  *wordstats.WordStats
	cfgFile    string
	debugFlag  bool
	cpuProfile string
	memProfile string
	rootCmd    = &cobra.Command{
		Use:   "autocorrector run",
		Short: "Autocorrect typos and spelling mistakes.",
		Long:  `Autocorrector is a tool similar to the word replacement functionality in Autokey or AutoHotKey.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cpuProfile != "" {
				log.Infof("Profling CPU to file %s", cpuProfile)
				f, err := os.Create(cpuProfile)
				if err != nil {
					log.Fatal("could not create CPU profile: ", err)
				}
				defer f.Close() // error handling omitted for example
				if err := pprof.StartCPUProfile(f); err != nil {
					log.Fatal("could not start CPU profile: ", err)
				}
				defer pprof.StopCPUProfile()
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			systray.Run(onReady, onExit)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if memProfile != "" {
				log.Infof("Profling Mem to file %s", memProfile)
				f, err := os.Create(memProfile)
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				defer f.Close() // error handling omitted for example
				runtime.GC()    // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
			}
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
	rootCmd.PersistentFlags().StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	rootCmd.PersistentFlags().StringVar(&memProfile, "memprofile", "", "write mem profile to `file`")
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

	// Read in the config file.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Could not find config file: ", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Errorf("Fatal error config file: %s", err))
		}
	}
	// Run checkConfig to ensure config used is safe.
	checkConfig()
	log.Debug("Config checks passed")
	viper.WatchConfig()
}

// checkConfig runs checks on the provided config to ensure it is safe to use
func checkConfig() {
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
