package core

import (
	"davinci/common"
	"davinci/common/log"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Ssh struct {
	conn       *ssh.Client
	session    *ssh.Session
	Host       string
	Port       int
	User       string
	Passwd     string
	PrivateKey []byte
	Cmd        string
}

func (s *Ssh) connect() error {
	var err error
	if s.conn == nil {
		var authList []ssh.AuthMethod
		// 私钥认证
		if s.PrivateKey != nil {

			if signer, err := ssh.ParsePrivateKey(s.PrivateKey); err != nil {
				log.Warn(err)
			} else {
				authList = append(authList, ssh.PublicKeys(signer))
			}
		}
		// 密码认证
		if s.Passwd != "" {
			authList = append(authList, ssh.Password(s.Passwd))
		}
		config := &ssh.ClientConfig{
			User:            s.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            authList,
			Timeout:         5 * time.Second,
		}
		if s.conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), config); err != nil {
			log.Error(err)
		}
	}
	return err
}

func (s *Ssh) Close() {
	if s.session != nil {
		s.session.Close()
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *Ssh) execute(cmd string) string {
	session, err := s.conn.NewSession()
	if err != nil {
		log.Warn(err)
		return ""
	}
	log.Info("execute: " + cmd)
	result, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Warn(err)
	}
	return string(result)
}

func (s *Ssh) ExecuteOnce() {
	if s.connect() != nil {
		return
	}

	result := s.execute(s.Cmd)
	log.Output(result)
}

func (s *Ssh) Shell() {
	if s.connect() != nil {
		return
	}

	session, err := s.conn.NewSession()
	if err != nil {
		log.Error(err)
		return
	}
	defer session.Close()
	log.Info("start ssh shell")
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Error(fmt.Errorf("MakeRaw failed: %v", err))
		return
	}
	defer terminal.Restore(fd, oldState)

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		log.Warn(fmt.Sprintf("Unable to get terminal size: %v,use default setting", err))
		termWidth = 40
		termHeight = 80
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 禁用本地回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度
		ssh.TTY_OP_OSPEED: 14400, // 输出速度
	}
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		log.Error("Request for pseudo terminal failed: ", err)
		return
	}
	if err := session.Shell(); err != nil {
		log.Error("Failed to start shell: ", err)
		return
	}

	// 等待交互式shell会话结束
	if err := session.Wait(); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitErr.ExitStatus())
		}
		log.Warn("Session wait failed: ", err)
	}
}

func (s *Ssh) AutoGather() {
	if s.connect() != nil {
		return
	}
	cmds := []string{
		"ls -al /",
		"ps -ef",
		"netstat -atlnp",
		"cat /etc/passwd",
		"uname -a",
		"env",
	}

	for _, cmd := range cmds {
		result := s.execute(cmd)
		log.Output(result)
	}
}

func (s *Ssh) SetHost(host string) {
	s.Host = host
}
func (s *Ssh) SetPort(port int) {
	s.Port = port
}

func (s *Ssh) SetCmd(cmd string) {
	s.Cmd = cmd
}

func (s *Ssh) ScpDownload(sourceFile, targetFile string) {
	if s.connect() != nil {
		return
	}
	if sourceFile == "" {
		log.Warn("please set source file")
		return
	}

	if targetFile == "" {
		targetFile = "./"
	}
	if common.IsDir(targetFile) {
		targetFile += filepath.Base(sourceFile)
	}

	log.Info("source file(remote): ", sourceFile)
	log.Info("target file(local) : ", targetFile)
	client, err := sftp.NewClient(s.conn)
	if err != nil {
		log.Error("create sftp error: ", err)
		return
	}
	defer client.Close()

	source, err := client.Open(sourceFile)
	if err != nil {
		log.Error("open (remote)source file error: ", err)
		return
	}
	defer source.Close()

	target, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Error("open (local)target file error: ", err)
		return
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	if err != nil {
		log.Warn("download file error: ", err)
	} else {
		log.Info("download file success")
	}
}

func (s *Ssh) ScpUpload(sourceFile, targetFile string) {
	if s.connect() != nil {
		return
	}
	if targetFile == "" {
		log.Warn("please set target file")
		return
	}
	if common.IsDir(targetFile) {
		targetFile += filepath.Base(sourceFile)
	}

	if !common.IsFileExist(sourceFile) {
		log.Warn("source file is not exist")
		return
	}
	if common.IsDir(sourceFile) {
		log.Warn("source file cannot be a dir")
		return
	}

	log.Info("source file(local) : ", sourceFile)
	log.Info("target file(remote): ", targetFile)
	client, err := sftp.NewClient(s.conn)
	if err != nil {
		log.Error("create sftp error: ", err)
		return
	}

	source, err := os.Open(sourceFile)
	if err != nil {
		log.Error("open (local)source file error: ", err)
		return
	}
	defer source.Close()

	defer client.Close()
	target, err := client.Create(targetFile)
	if err != nil {
		log.Error("open (remote)target file error: ", err)
		return
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	if err != nil {
		log.Error("upload file error: ", err)
	} else {
		log.Info("upload file success")
	}

}
