package core

import (
	"io/ioutil"
	"testing"
)

var s = &Ssh{
	Host:   "192.168.83.129",
	Port:   22,
	User:   "root",
	Cmd:    "whoami",
	Passwd: "123qwe",
}

func TestSsh_ExecuteOnce(t *testing.T) {
	s.ExecuteOnce()
}

func TestSsh_ExecuteOnce_by_PrivateKey(t *testing.T) {
	privakey, _ := ioutil.ReadFile("D:/tmp/id_rsa")
	s.PrivateKey = privakey
	s.ExecuteOnce()
}

func TestSsh_Shell(t *testing.T) {
	s.Shell()
}

func TestSsh_AutoGather(t *testing.T) {
	s.AutoGather()
}

func TestSsh_ScpDownload(t *testing.T) {
	privakey, _ := ioutil.ReadFile("D:/tmp/id_rsa")
	s.PrivateKey = privakey
	s.ExecuteOnce()

	s.ScpDownload("/root/.ssh/authorized_keys", "d:/tmp/ssh_test/authorized_keys")
}

func TestSsh_ScpUpload(t *testing.T) {
	privakey, _ := ioutil.ReadFile("D:/tmp/id_rsa")
	s.PrivateKey = privakey
	s.ScpUpload("d:/tmp/ssh_test/authorized_keys", "/root/.ssh/aaa")
}
