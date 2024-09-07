package common

import (
	"davinci/common/log"
	"fmt"
	"github.com/projectdiscovery/mapcidr"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var client *http.Client

func GetHttpClient() *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	return client
}

func Request(req *http.Request) *http.Response {
	client := GetHttpClient()
	res, err := client.Do(req)
	if err != nil {
		log.Warn(err)
		return nil
	}
	return res
}

func GetAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil

}

func ParseIps(ipDesc string) []string {
	// 合法ip
	if net.ParseIP(ipDesc) != nil {
		return []string{ipDesc}
	}
	// cidr
	if iplist, err := mapcidr.IPAddresses(ipDesc); err == nil {
		return delBroadcast(iplist)
	}
	// -
	var iplist []string
	index := strings.Index(ipDesc, "-")
	length := len(ipDesc)
	if index > 0 && index < length {
		ipStart := ipDesc[0:index]
		ipEndNumStr := ipDesc[index+1 : length]
		ipStartParse := net.ParseIP(ipStart)
		if ipStartParse != nil && ipStartParse.To4() != nil {
			ips := strings.Split(ipStart, ".")
			ipStartNumStr := ips[3]
			ipStartNum, err1 := strconv.Atoi(ipStartNumStr)
			ipEndNum, err2 := strconv.Atoi(ipEndNumStr)
			if err1 == nil && err2 == nil && ipStartNum <= ipEndNum {
				for i := ipStartNum; i <= ipEndNum; i++ {
					ip := fmt.Sprintf("%s.%s.%s.%d", ips[0], ips[1], ips[2], i)
					iplist = append(iplist, ip)
				}
				return delBroadcast(iplist)
			}
		}
	}
	log.Warn("parse ip fail: ", ipDesc)
	return []string{}
}

func delBroadcast(ips []string) []string {
	var newIps []string
	for _, ip := range ips {
		if strings.HasSuffix(ip, ".0") {
			log.Info("del broadcast in ip list: ", ip)
			continue
		}
		newIps = append(newIps, ip)
	}
	return newIps
}
