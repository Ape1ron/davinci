package common

import (
	"os"
	"strings"
)

func SplitFilePath(file string) (string, string) {
	var dir, filename string
	if file[len(file)-1] == '/' {
		file = file[:len(file)-1]
	}
	index := strings.LastIndex(file, "/")
	if index >= 0 {
		dir = file[:index+1]
		filename = file[index+1:]
	} else {
		dir = ""
		filename = file
	}
	return dir, filename
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func IsDir(file string) bool {
	fileInfo, err := os.Stat(file)
	if err == nil && fileInfo.IsDir() {
		return true
	}
	return false
}

func WriteFile(fileName string, content []byte) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}
