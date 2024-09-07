package cmd

import (
	"davinci/core"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	chHost   string
	chPort   int
	chUser   string
	chPasswd string
	chDbName string
	chSql    string

	chCmd = &cobra.Command{
		Use:   "clickhouse [exec|shell|auto_gather]",
		Short: "clickhouse client",
		Long: "clickhouse client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       clickhouse interactive shell\n" +
			"  - auto_gather: automatically collect database information, including users, databases,\n" +
			"                 tables, table structures, and the first 5 rows of data for each table",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.ClickHouse{
				Host:   chHost,
				Port:   chPort,
				User:   chUser,
				Passwd: chPasswd,
				DbName: chDbName,
				Cmd:    chSql,
			}
			defer service.Close()
			switch args[0] {
			case "exec":
				if chSql == "" {
					return fmt.Errorf("davinci clickhouse exec must have cmd")
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

	chCmd.Flags().StringVarP(&chHost, "host", "H", "127.0.0.1", "clickhouse ip addr")
	chCmd.Flags().IntVarP(&chPort, "port", "P", 9000, "clickhouse port")
	chCmd.Flags().StringVarP(&chUser, "user", "u", "default", "username")
	chCmd.Flags().StringVarP(&chPasswd, "passwd", "p", "", "pasword")
	chCmd.Flags().StringVarP(&chDbName, "dbName", "d", "", "database name,not require")
	chCmd.Flags().StringVarP(&chSql, "cmd", "c", "", "clickhouse cmd to be executed, only used in exec mode")
	chCmd.MarkFlagRequired("passwd")
	rootCmd.AddCommand(chCmd)
}
