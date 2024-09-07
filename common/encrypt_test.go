package common

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestName(t *testing.T) {
	prikey, _ := ioutil.ReadFile("d://tmp/x.txt")
	y, err := EncryptRSAPrivateKey(prikey, "1234567")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(y))

	x, err := DecryptRSAPrivateKey([]byte(y), "1234567")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(x))
}

func Test2(t *testing.T) {
	fmt.Println(GetAvailablePort())
}
