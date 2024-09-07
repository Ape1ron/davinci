package core

import (
	"davinci/common"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"
)

//var mysql = &Mysql{
//	Host:   "192.168.159.135",
//	Port:   3307,
//	User:   "root",
//	Passwd: "123456",
//	DbName: "",
//	Cmd:    "show databases",
//}
//
//func TestExecuteOnce(t *testing.T) {
//	//conn, _ = client.Connect("192.168.159.135:3307", "root", "123456", "")
//	//r, _ := conn.Execute("desc information_schema.CHARACTER_SETS")
//	//fmt.Println(r.RowNumber())
//	//fmt.Println(r.ColumnNumber())
//	//fmt.Println(fmt.Sprintf("%v", r.FieldNames))
//	//fmt.Println(string(r.Fields[0].Name))
//	//defer r.Close()
//	defer mysql.Close()
//	mysql.ExecuteOnce()
//}

func TestXName(t *testing.T) {
	file := "D:\\project\\GolangProject\\davinci\\lib\\redis\\linux\\x86\\eval_x86.so"
	if content, readErr := ioutil.ReadFile(file); readErr == nil {
		h := "0x" + hex.EncodeToString(content)
		fmt.Println(h)
	} else {
		fmt.Println(readErr)
	}
}

func TestCompareVersion(t *testing.T) {
	version := "5.7.40"
	fmt.Println(common.CompareVersion("5.0", version, "."))
	fmt.Println(common.CompareVersion("5.1", version, "."))
	fmt.Println(common.CompareVersion("5.7.39", version, "."))
	fmt.Println(common.CompareVersion("5.7.40", version, "."))
	fmt.Println(common.CompareVersion("5.7.41", version, "."))
	fmt.Println(common.CompareVersion("5.8", version, "."))
	fmt.Println(common.CompareVersion("8.8", version, "."))
	fmt.Println(common.CompareVersion("10.8", version, "."))
}
