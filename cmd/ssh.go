package cmd

import (
	"davinci/core"
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	sshHost             string
	sshPort             int
	sshUser             string
	sshPasswd           string
	sshPrivateKeyBase64 string
	sshPrivateKeyfile   string
	sshExec             string
	sshSourceFile       string
	sshTargetFile       string

	sshCmd = &cobra.Command{
		Use:   "ssh [exec|shell|auto_gather|download|upload]",
		Short: "ssh client",
		Long: "ssh client:\n" +
			"  - exec:        execute the command once and return directly\n" +
			"  - shell:       ssh interactive shell\n" +
			"  - auto_gather: execute some cmd to gather basic info,such ps/netstat/etc.\n" +
			"  - download:    scp download file (remote -> local)\n" +
			"  - upload:      scp upload file (local -> remote)\n" +
			"\n" +
			"Example: \n" +
			"  davinci ssh exec        -H 192.168.1.2 -P 22 -u root -p 123456 -c \"id\" \n" +
			"  davinci ssh shell       -H 192.168.1.2 -P 22 -u root -p 123456 \n" +
			"  davinci ssh auto_gather -H 192.168.1.2 -P 22 -u root -p 123456 \n" +
			"  davinci ssh download    -H 192.168.1.2 -P 22 -u root -p 123456 -s d:/tmp/ -t /tmp/1.zip \n" +
			"  davinci ssh uplload     -H 192.168.1.2 -P 22 -u root -p 123456 -s d:/tmp/1.zip -t /tmp/ \n" +
			"  ",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"exec", "shell", "auto_gather", "download", "upload"},
		RunE: func(command *cobra.Command, args []string) error {
			service := &core.Ssh{
				Host:   sshHost,
				Port:   sshPort,
				User:   sshUser,
				Passwd: sshPasswd,
				Cmd:    sshExec,
			}
			if sshPrivateKeyBase64 != "" {
				privateKey, err := base64.StdEncoding.DecodeString(sshPrivateKeyBase64)
				if err != nil {
					return fmt.Errorf("decoding private key error: ", err)
				}
				service.PrivateKey = privateKey
			} else if sshPrivateKeyfile != "" {
				privakey, err := ioutil.ReadFile(sshPrivateKeyfile)
				if err != nil {
					return fmt.Errorf("read private key error: ", err)
				}
				service.PrivateKey = privakey
			}

			defer service.Close()
			switch args[0] {
			case "exec":
				if sshExec == "" {
					return fmt.Errorf("davinci ssh exec must have cmd")
				}
				service.ExecuteOnce()
			case "shell":
				service.Shell()
			case "auto_gather":
				service.AutoGather()
			case "download":
				service.ScpDownload(sshSourceFile, sshTargetFile)
			case "upload":
				service.ScpUpload(sshSourceFile, sshTargetFile)
			}
			return nil
		},
	}
)

func init() {
	sshCmd.Flags().StringVarP(&sshHost, "host", "H", "127.0.0.1", "target ip addr")
	sshCmd.Flags().IntVarP(&sshPort, "port", "P", 22, "ssh port")
	sshCmd.Flags().StringVarP(&sshUser, "user", "u", "root", "username")
	sshCmd.Flags().StringVarP(&sshPasswd, "passwd", "p", "", "pasword")
	sshCmd.Flags().StringVarP(&sshPrivateKeyBase64, "prikey", "k", "", "base64 encoding of the private key,use for private key auth")
	sshCmd.Flags().StringVarP(&sshPrivateKeyfile, "prifile", "f", "", "private key file,use for private key auth")
	sshCmd.Flags().StringVarP(&sshExec, "cmd", "c", "", "cmd to be executed, only used in exec mode")
	sshCmd.Flags().StringVarP(&sshSourceFile, "source", "s", "", "source file,use for scp")
	sshCmd.Flags().StringVarP(&sshTargetFile, "target", "t", "", "target file,use for scp")

	rootCmd.AddCommand(sshCmd)
}
