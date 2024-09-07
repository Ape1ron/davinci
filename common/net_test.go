package common

import (
	"davinci/common/log"
	"fmt"
	"os"
	"testing"
)

func init() {
	log.AddLogWriter(os.Stdout)
}

func TestParseIp(t *testing.T) {
	ip := "192.158.1.25"
	fmt.Println(ParseIps(ip))
	ip = "192.168.1.1-22"
	fmt.Println(ParseIps(ip))
	ip = "192.168.1.0/23"
	fmt.Println(ParseIps(ip))
	ip = "192.255.256.1-20"
	fmt.Println(ParseIps(ip))
}
