package cmd

import (
	"davinci/core"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	gsHost   string
	gsPort   int
	gsUser   string
	gsPasswd string
	gsDbName string
	gsSql    string

	gsCmd = &cobra.Command{
		Use:   "gaussdb [exec|shell|auto_gather]",
		Short: "gaussdb client",
		Long: "gaussdb client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       gaussdb interactive shell\n" +
			"  - auto_gather: automatically collect database information, including users, databases,\n" +
			"                 tables, table structures, and the first 5 rows of data for each table",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather"},
		RunE: func(command *cobra.Command, args []string) error {
			// gaussdb is compatible with postgresql
			service := &core.GaussDB{
				Pgsql: &core.Pgsql{
					Host:   gsHost,
					Port:   gsPort,
					User:   gsUser,
					Passwd: gsPasswd,
					DbName: gsDbName,
					Cmd:    gsSql,
				},
			}
			defer service.Close()
			switch args[0] {
			case "exec":
				if gsSql == "" {
					return fmt.Errorf("davinci pgsql exec must have cmd")
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

	gsCmd.Flags().StringVarP(&gsHost, "host", "H", "127.0.0.1", "gaussdb ip addr")
	gsCmd.Flags().IntVarP(&gsPort, "port", "P", 5432, "gaussdb port")
	gsCmd.Flags().StringVarP(&gsUser, "user", "u", "gaussdb", "username")
	gsCmd.Flags().StringVarP(&gsPasswd, "passwd", "p", "", "pasword")
	gsCmd.Flags().StringVarP(&gsDbName, "dbName", "d", "postgres", "database name,not require")
	gsCmd.Flags().StringVarP(&gsSql, "cmd", "c", "", "gaussdb cmd to be executed, only used in exec mode")
	gsCmd.MarkFlagRequired("passwd")
	rootCmd.AddCommand(gsCmd)
}
