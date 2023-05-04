package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/joshuar/autocorrector/internal/app"
	"github.com/joshuar/autocorrector/internal/wordstats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	correctionsFlag string
	logStatsFlag    bool
	clientCmd       = &cobra.Command{
		Use:   "client",
		Short: "Client creates tray icon for control and notifications",
		Long:  `With the client running, you can pause correction and see notifications.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setDebugging()
			setProfiling()
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
			app := app.New()
			fmt.Println("starting app...")
			app.Run()
		},
	}
	clientSetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Set-up an autocorrector client",
		Long:  "The client set-up command will make a copy of the default corrections file and create an autostart entry for the current user",
		Run: func(cmd *cobra.Command, args []string) {
			configDirectory := xdg.ConfigHome + "/autocorrector"
			log.Infof("Creating configuration directory %s", configDirectory)
			err := os.Mkdir(configDirectory, 0755)
			if e, ok := err.(*os.PathError); ok {
				if e.Err == syscall.EEXIST {
					log.Warn("Configuration directory already exists")
				} else {
					log.Fatalf("Unable to create configuration directory %s: ", configDirectory, err)
				}
			}
			defaultCorrections, err := ioutil.ReadFile("/usr/local/share/autocorrector/corrections.toml")
			if err != nil {
				log.Fatalf("Unable to read default corrections file: ", err)
			}
			defaultCorrectionsFile := xdg.ConfigHome + "/autocorrector/corrections.toml"
			log.Infof("Copying default configuration file %s", defaultCorrectionsFile)
			err = ioutil.WriteFile(defaultCorrectionsFile, defaultCorrections, 0755)
			if err != nil {
				log.Fatalf("Unable to write corrections file %s:", defaultCorrectionsFile, err)
			}
			log.Infof("Creating autostart entry")
			err = os.Symlink("/usr/local/share/applications/autocorrector.desktop", xdg.ConfigHome+"/autostart/autocorrector.desktop")
			if e, ok := err.(*os.LinkError); ok && e.Err == syscall.EEXIST {
				log.Warn("Autostart entry already exists")
			} else {
				log.Fatal("Unable to create autostart entry:", err)
			}
		},
	}
	statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Print statistics from the database",
		Long:  `Show stats such as number of checked/corrected words and accuracy.`,
		Run: func(cmd *cobra.Command, args []string) {
			wordStats := wordstats.RunStats()
			wordStats.ShowStats()
			if logStatsFlag {
				wordStats.ShowLog()
			}
			wordStats.CloseWordStats()
			os.Exit(0)
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	clientCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
	clientCmd.Flags().StringVar(&correctionsFlag, "corrections", "", fmt.Sprintf("list of corrections (default is %s/autocorrector/corrections.toml)", xdg.ConfigHome))
	clientCmd.AddCommand(clientSetupCmd)
	clientCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolVarP(&logStatsFlag, "log", "l", false, "Show log of corrections")
}
