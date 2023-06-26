// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	_ "net/http/pprof"
	"os"

	"github.com/joshuar/autocorrector/internal/app"
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
			setDebugging()
			setProfiling()
		},
		Run: func(cmd *cobra.Command, args []string) {
			app := app.New()
			app.Run()
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
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	rootCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
}
