package cmd

import (
	"os"

	"github.com/joshuar/autocorrector/internal/wordstats"
	"github.com/spf13/cobra"
)

var (
	logStatsFlag bool
)

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolVarP(&logStatsFlag, "log", "l", false, "Show log of corrections")
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print statistics from the database",
	Long:  `Show stats such as number of checked/corrected words and accuracy.`,
	Run: func(cmd *cobra.Command, args []string) {
		wordStats := wordstats.OpenWordStats()
		wordStats.ShowStats()
		if logStatsFlag {
			wordStats.ShowLog()
		}
		wordStats.CloseWordStats()
		os.Exit(0)
	},
}
