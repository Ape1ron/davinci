package cmd

import (
	"davinci/core"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	mongoHost   string
	mongoPort   int
	mongoUser   string
	mongoPasswd string
	mongoDbName string
	mongoExec   string

	mongoCmd = &cobra.Command{
		Use:   "mongo [exec|shell|auto_gather]",
		Short: "mongo client",
		Long: "mongo client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       mongo interactive shell\n" +
			"  - auto_gather: automatically collect database information, including users, databases,\n" +
			"                 tables, table structures, and the first 5 rows of data for each table",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.MongoDb{
				Host:   mongoHost,
				Port:   mongoPort,
				User:   mongoUser,
				Passwd: mongoPasswd,
				DbName: mongoDbName,
				Cmd:    mongoExec,
			}
			defer service.Close()
			switch args[0] {
			case "exec":
				if mongoExec == "" {
					return fmt.Errorf("davinci mongo exec must have cmd")
				}
				service.ExecuteOnce()
			case "shell":
				service.Shell()
			case "auto_gather":
				service.AutoGather()
			}
			return nil
		},
	}
)

func init() {

	mongoCmd.Flags().StringVarP(&mongoHost, "host", "H", "127.0.0.1", "mongodb ip addr")
	mongoCmd.Flags().IntVarP(&mongoPort, "port", "P", 27017, "mongodb port")
	mongoCmd.Flags().StringVarP(&mongoUser, "user", "u", "", "username")
	mongoCmd.Flags().StringVarP(&mongoPasswd, "passwd", "p", "", "pasword")
	mongoCmd.Flags().StringVarP(&mongoDbName, "dbName", "d", "", "database name")
	mongoCmd.Flags().StringVarP(&mongoExec, "cmd", "c", "", "mongo cmd to be executed, only used in exec mode")
	rootCmd.AddCommand(mongoCmd)
}
