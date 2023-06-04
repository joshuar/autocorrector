package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/joshuar/autocorrector/internal/app"
	"github.com/joshuar/autocorrector/internal/wordstats"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	default_corrections_path = "/usr/share/autocorrector/corrections.toml"
	default_desktop_file     = "/usr/share/applications/autocorrector.desktop"
)

var (
	correctionsFlag string
	logStatsFlag    bool
	clientCmd       = &cobra.Command{
		Use:   "client",
		Short: "Client creates tray icon for control and notifications",
		Long:  `With the client running, you can pause correction and see notifications.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setLogging()
			setDebugging()
			setProfiling()
			if correctionsFlag != "" {
				log.Info().Msgf("Using config file specified on command-line: ", correctionsFlag)
				viper.SetConfigFile(correctionsFlag)
			} else {
				viper.SetConfigName("corrections")
				viper.SetConfigType("toml")
				viper.AddConfigPath(xdg.ConfigHome + "/autocorrector")
				viper.AddConfigPath("/usr/share/autocorrector")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			app := app.New()
			app.Run()
		},
	}
	clientSetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Set-up an autocorrector client",
		Long:  "The client set-up command will make a copy of the default corrections file and create an autostart entry for the current user",
		Run: func(cmd *cobra.Command, args []string) {
			configDirectory := xdg.ConfigHome + "/autocorrector"
			log.Info().Msgf("Creating configuration directory %s.", configDirectory)
			err := os.Mkdir(configDirectory, 0755)
			if e, ok := err.(*os.PathError); ok {
				if e.Err == syscall.EEXIST {
					log.Warn().Msg("Configuration directory already exists")
				} else {
					log.Panic().Err(err).Msgf("Unable to create configuration directory %s.", configDirectory)
				}
			}
			defaultCorrections, err := os.ReadFile(default_corrections_path)
			if err != nil {
				log.Panic().Err(err).Msg("Unable to read default corrections file.")
			}
			defaultCorrectionsFile := xdg.ConfigHome + "/autocorrector/corrections.toml"
			log.Info().Msgf("Copying default configuration file %s.", defaultCorrectionsFile)
			err = os.WriteFile(defaultCorrectionsFile, defaultCorrections, 0755)
			if err != nil {
				log.Panic().Err(err).Msgf("Unable to write corrections file %s.", defaultCorrectionsFile)
			}
			log.Info().Msg("Creating autostart entry")
			err = os.Symlink(default_desktop_file, xdg.ConfigHome+"/autostart/autocorrector.desktop")
			if e, ok := err.(*os.LinkError); ok && e.Err == syscall.EEXIST {
				log.Warn().Msg("Autostart entry already exists")
			} else {
				log.Panic().Err(err).Msg("Unable to create autostart entry.")
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
