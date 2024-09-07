package cmd

import (
	"davinci/common"
	"davinci/common/log"
	"davinci/core"
	"encoding/base64"
	"encoding/json"
	"github.com/spf13/cobra"
	"net"
	"os"
	"strings"
)

const defaultBatchFile = ".davinci_batch.json"

type batchExec struct {
	CmdType string   `json:"cmd_type"`
	Hosts   []string `json:"hosts"`
	Port    int      `json:"port,omitempty"`
	User    string   `json:"user,omitempty"`
	Passwd  string   `json:"passwd,omitempty"`
	Cmds    []string `json:"cmds"`
}

var (
	file      string
	b64Config string

	batchCmd = &cobra.Command{
		Use:   "batch [export|exec]",
		Short: "batch execute",
		Long: "batch execute:\n" +
			"  - export:      export execute config template\n" +
			"  - exec  :      batch execute",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"export", "exec"},
		RunE: func(command *cobra.Command, args []string) error {
			switch args[0] {
			case "export":
				batchTemp1 := &batchExec{
					CmdType: "ssh",
					Hosts:   []string{"127.0.0.1", "192.168.83.1/24", "192.168.83.1-20"},
					Port:    22,
					User:    "root",
					Passwd:  "123456",
					Cmds:    []string{"ls -al /", "ifconfig"},
				}
				batchTemp2 := &batchExec{
					CmdType: "redis",
					Hosts:   []string{"127.0.0.1"},
					Port:    6379,
					Cmds:    []string{"dbsize"},
				}
				var batchTempList []*batchExec
				batchTempList = append(batchTempList, batchTemp1)
				batchTempList = append(batchTempList, batchTemp2)
				if jsonData, err := json.MarshalIndent(batchTempList, "", "    "); err != nil {
					log.Error(err)
				} else {
					if err := common.WriteFile(file, jsonData); err == nil {
						log.Info("export batch config template success: ", file)
					} else {
						log.Error(err)
						return err
					}
				}
			case "exec":
				var data []byte
				var err error
				if b64Config != "" {
					data, err = base64.StdEncoding.DecodeString(b64Config)
				} else {
					data, err = os.ReadFile(file)
				}
				if err != nil {
					log.Error(err)
					return err
				}
				var batchs []batchExec
				if err := json.Unmarshal(data, &batchs); err != nil {
					log.Error(err)
					return err
				}
				log.Info("load batch config success: ", file)
				batchExecute(batchs)
			}
			return nil
		},
	}
)

func batchExecute(batchs []batchExec) {
	for _, batch := range batchs {
		service := parseBatchExec(batch)
		if service == nil {
			continue
		}
		for _, ips := range batch.Hosts {
			ipList := common.ParseIps(ips)
			if ipList != nil {
				for _, ip := range ipList {
					service.SetHost(ip)
					log.Info("try batch execute in : ", ip)
					batchExecCmd(service, batch.Cmds)
				}
				continue
			}
			_, err := net.ResolveIPAddr("ip", ips)
			if err == nil {
				log.Info("try batch execute in : ", ips)
				service.SetHost(ips)
				batchExecCmd(service, batch.Cmds)
			}
		}

	}
}

func batchExecCmd(service core.Service, cmds []string) {
	for _, cmd := range cmds {
		service.SetCmd(cmd)
		service.ExecuteOnce()
	}
}

func parseBatchExec(batch batchExec) core.Service {
	var service core.Service
	batch.CmdType = strings.ToLower(batch.CmdType)
	switch batch.CmdType {
	case "ssh":
		if batch.Port == 0 {
			batch.Port = 22
		}
		service = &core.Ssh{
			Port:   batch.Port,
			User:   batch.User,
			Passwd: batch.Passwd,
		}
	case "mysql":
		if batch.Port == 0 {
			batch.Port = 3306
		}
		service = &core.Mysql{
			Port:   batch.Port,
			User:   batch.User,
			Passwd: batch.Passwd,
		}
	case "pgsql":
		if batch.Port == 0 {
			batch.Port = 5432
		}
		service = &core.Pgsql{
			Port:   batch.Port,
			User:   batch.User,
			Passwd: batch.Passwd,
		}
	case "gaussdb":
		if batch.Port == 0 {
			batch.Port = 5432
		}
		service = &core.GaussDB{
			Pgsql: &core.Pgsql{
				Port:   batch.Port,
				User:   batch.User,
				Passwd: batch.Passwd,
			},
		}
	case "clickhouse":
		if batch.Port == 0 {
			batch.Port = 9000
		}
		service = &core.ClickHouse{
			Port:   batch.Port,
			User:   batch.User,
			Passwd: batch.Passwd,
		}
	case "redis":
		if batch.Port == 0 {
			batch.Port = 6379
		}
		service = &core.Redis{
			Port:   batch.Port,
			User:   batch.User,
			Passwd: batch.Passwd,
		}
	case "mongo":
		if batch.Port == 0 {
			batch.Port = 27017
		}
		service = &core.MongoDb{
			Port:   batch.Port,
			User:   batch.User,
			Passwd: batch.Passwd,
		}
	}
	return service
}

func init() {
	batchCmd.Flags().StringVarP(&file, "file", "f", defaultBatchFile, "config file")
	batchCmd.Flags().StringVarP(&b64Config, "b64config", "b", "", "the config base64 encode")
	rootCmd.AddCommand(batchCmd)
}
