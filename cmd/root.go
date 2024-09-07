package cmd

import (
	"davinci/common/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"
)

var noLog bool
var silent bool

var (
	rootCmd = &cobra.Command{
		Use:   "davinci",
		Short: "multi-component client",
		Long: "multi-component client,include  database, middleware, queue, etc.\n" +
			"used for red team simulation scenarios",
	}
)

func Execute() error {
	defer log.Close()
	// 还原terminal
	oldInState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.Warn(err)
	} else {
		defer term.Restore(int(os.Stdin.Fd()), oldInState)
	}
	oldOutState, err := term.GetState(int(os.Stdout.Fd()))
	if err != nil {
		log.Warn(err)
	} else {
		defer term.Restore(int(os.Stdout.Fd()), oldOutState)
	}
	oldErrState, err := term.GetState(int(os.Stderr.Fd()))
	if err != nil {
		log.Warn(err)
	} else {
		defer term.Restore(int(os.Stderr.Fd()), oldErrState)
	}

	return rootCmd.Execute()

}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&noLog, "no-log", "", false, "not log to file")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "", false, "close info level output")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if noLog {
			log.AddLogWriter(os.Stdout)
		}
		if silent {
			log.SetLogLevel(log.WarnLevel)
		}
	}
}
