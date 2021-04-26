package cmd

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/user"

	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/autocorrector/internal/keytracker"
	log "github.com/sirupsen/logrus"
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
			currentUser, err := user.Current()
			if err != nil {
				log.Fatalf("Could fetch user that ran us: %s", err)
			}
			if currentUser.Username != "root" {
				log.Fatal("autocorrector server must be run as root")
			}
			if debugFlag {
				log.SetLevel(log.DebugLevel)
			}
			if profileFlag {
				go func() {
					log.Debug(http.ListenAndServe("localhost:6060", nil))
				}()
				log.Debug("Profiling is enabled and available at localhost:6060")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			manager := control.NewConnManager(userFlag)
			go manager.Start()

			keyTracker := keytracker.NewKeyTracker()
			keyTracker.EventWatcher(manager)
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
	rootCmd.Flags().StringVar(&userFlag, "user", "", "user to allow access to control socket")
	rootCmd.MarkFlagRequired("user")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	rootCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
}
