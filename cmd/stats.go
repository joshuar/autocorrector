package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/joshuar/autocorrector/internal/wordstats"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print statistics from the database",
	Long:  `Show stats such as number of checked/corrected words and accuracy.`,
	Run: func(cmd *cobra.Command, args []string) {
		wordStats := wordstats.OpenWordStats()
		log.Infof("%v words checked.", wordStats.GetCheckedTotal())
		log.Infof("%v words corrected.", wordStats.GetCorrectedTotal())
		log.Infof("Accuracy is: %.2f %%.", wordStats.CalcAccuracy())
		wordStats.CloseWordStats()
		os.Exit(0)
	},
}
