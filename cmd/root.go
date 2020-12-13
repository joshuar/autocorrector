package cmd

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	cfgFile   string
	debugFlag bool
	rootCmd   = &cobra.Command{
		Use:   "autocorrector run",
		Short: "Autocorrect typos and spelling mistakes.",
		Long:  `Autocorrector is a tool similar to the word replacement functionality in Autokey or AutoHotKey.`,
		Run:   func(cmd *cobra.Command, args []string) {},
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
}

// initConfig reads in config file
func initConfig() {
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal finding home directory: %s", err))
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
