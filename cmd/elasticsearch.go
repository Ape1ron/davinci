package cmd

import (
	"davinci/core"
	"github.com/spf13/cobra"
)

var (
	esUrl    string
	esUser   string
	esPasswd string
	force    bool

	esCmd = &cobra.Command{
		Use:   "es [auto_gather]",
		Short: "elasticsearch client",
		Long: "elasticsearch client:\n" +
			"  - auto_gather: automatically collect es information, including nodes,users,\n" +
			"                 total amount of data, indices,first 5 data of each index",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"auto_gather"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.ElasticSearch{
				Url:    esUrl,
				User:   esUser,
				Passwd: esPasswd,
				Check:  !force,
			}
			switch args[0] {
			case "auto_gather":
				service.AutoGather()
			}
			return nil
		},
	}
)

func init() {

	esCmd.Flags().StringVarP(&esUrl, "url", "U", "http://127.0.0.1:9200", "elasticsearch addr")
	esCmd.Flags().StringVarP(&esUser, "user", "u", "", "username")
	esCmd.Flags().StringVarP(&esPasswd, "passwd", "p", "", "pasword")
	esCmd.Flags().BoolVar(&force, "force", false, "do not check elasticsearch api enable,force request")
	rootCmd.AddCommand(esCmd)
}
