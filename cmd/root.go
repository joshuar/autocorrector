package cmd

import (
	_ "net/http/pprof"
	"os"
	"os/exec"

	"github.com/joshuar/autocorrector/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	userFlag    string
	debugFlag   bool
	profileFlag bool
	rootCmd     = &cobra.Command{
		Use:   "autocorrector",
		Short: "Autocorrect typos and spelling mistakes.",
		Long:  `Autocorrector is a tool similar to the word replacement functionality in Autokey or AutoHotKey.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setLogging()
			ensureEUID()
			setDebugging()
			setProfiling()
		},
		Run: func(cmd *cobra.Command, args []string) {
			server.Run(userFlag)
		},
	}
	enableCmd = &cobra.Command{
		Use:   "enable [username]",
		Short: "Enable the autocorrector service for the specified user",
		Long:  "Copies and enables an autocorrector systemd service for the specified user",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ensureEUID()
			systemdReload := exec.Command("systemctl", "daemon-reload")
			err := systemdReload.Run()
			if err != nil {
				log.Warn().Err(err).Msgf("Try manually running the following command and fix any errors it returns: %s", systemdReload.String())
			}
			systemdEnable := exec.Command("systemctl", "enable", "autocorrector@"+args[0])
			err = systemdEnable.Run()
			if err != nil {
				log.Warn().Err(err).Msgf("Try manually running the following command and fix any errors it returns: %s", systemdEnable.String())
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("Could not start.")
		os.Exit(1)
	}
}

// init defines flags and configuration settings
func init() {
	rootCmd.Flags().StringVar(&userFlag, "user", "", "user to allow access to control socket")
	rootCmd.MarkFlagRequired("user")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	rootCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
	rootCmd.AddCommand(enableCmd)
}
