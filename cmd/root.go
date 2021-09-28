package cmd

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
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
			var socket *control.Socket

			kbdTracker := keytracker.NewKeyTracker()
			defer kbdTracker.CloseKeyTracker()

			kbdTracker.StartEvents()

			for {
				socket = control.NewSocket(userFlag)
				go socket.RecvData()

				for {
					select {
					case msg := <-socket.Data:
						switch t := msg.(type) {
						case *control.StateMsg:
							switch t.State {
							case control.Start:
								kbdTracker.Start()
							case control.Pause:
								kbdTracker.Pause()
							case control.Resume:
								kbdTracker.Resume()
							default:
								log.Debugf("Unhandled state: %v", msg)
							}
						case *control.WordMsg:
							w := keytracker.NewWord(t.Word, t.Correction, t.Punct)
							kbdTracker.WordCorrection <- *w
						default:
							log.Debugf("Unhandled message recieved: %v", msg)
						}
					case <-socket.Done:
						log.Debug("Received done, restarting socket...")
						socket = control.NewSocket(userFlag)
						go socket.RecvData()
					case w := <-kbdTracker.TypedWord:
						socket.SendWord(w.Word, "", w.Punct)
					}
				}
			}
		},
	}
	enableCmd = &cobra.Command{
		Use:   "enable [username]",
		Short: "Enable the autocorrector service for the specified user",
		Long:  "Copies and enables an autocorrector systemd service for the specified user",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			systemdReload := exec.Command("systemctl", "daemon-reload")
			err := systemdReload.Run()
			if err != nil {
				log.Warn(err)
				log.Warnf("Try manually running the following command and fix any errors it returns: %s", systemdReload.String())
			}
			systemdEnable := exec.Command("systemctl", "enable", "autocorrector@"+args[0])
			err = systemdEnable.Run()
			if err != nil {
				log.Warn(err)
				log.Warnf("Try manually running the following command and fix any errors it returns: %s", systemdEnable.String())
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
	rootCmd.Flags().StringVar(&userFlag, "user", "", "user to allow access to control socket")
	rootCmd.MarkFlagRequired("user")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output")
	rootCmd.Flags().BoolVarP(&profileFlag, "profile", "", false, "enable profiling")
	rootCmd.AddCommand(enableCmd)
}
