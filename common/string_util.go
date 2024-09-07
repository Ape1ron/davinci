package common

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"time"
)

var escapeMap = map[string]string{
	"\\a":  "\a",
	"\\b":  "\b",
	"\\f":  "\f",
	"\\n":  "\n",
	"\\r":  "\r",
	"\\t":  "\t",
	"\\v":  "\v",
	"\\\\": "\\",
	"\\'":  "'",
	"\\\"": "\"",
	"\\?":  "?",
	"\\0":  "",
}

func findCharNotInStr(content string, target rune) []int {
	stack := make([]rune, 0)
	var result []int
	for i := 1; i < len(content); i++ {
		c := content[i]
		var top rune
		if len(stack) == 0 {
			top = rune(0)
		} else {
			top = stack[len(stack)-1]
		}

		switch c {
		case '"':
			if top == '"' {
				stack = stack[:len(stack)-1]
			} else if top != '\'' {
				stack = append(stack, '"')
			}
		case '\'':
			if top == '\'' {
				stack = stack[:len(stack)-1]
			} else if top != '"' {
				stack = append(stack, '\'')
			}
		case uint8(target):
			if len(stack) == 0 {
				result = append(result, i)
			}
		case '\\':
			i++
		}
	}
	return result
}

func FindFirstCharNotInStr(content string, target rune) int {
	indexs := findCharNotInStr(content, target)
	if len(indexs) == 0 {
		return -1
	}
	return indexs[0]
}

func FindLastCharNotInStr(content string, target rune) int {
	indexs := findCharNotInStr(content, target)
	if len(indexs) == 0 {
		return -1
	}
	return indexs[len(indexs)-1]
}

func SplitCmd(content string, target rune) []string {
	indexs := findCharNotInStr(content, target)
	if len(indexs) == 0 {
		return []string{content}
	}
	var result []string
	start := 0
	for _, index := range indexs {
		result = append(result, content[start:index])
		start = index + 1
	}
	result = append(result, content[start:])
	return result
}

func DelElement(array []string, element string) []string {
	var result []string
	for _, val := range array {
		if val != element {
			result = append(result, val)
		}
	}
	return result
}

func DelEmptyEle(array []string) []string {
	return DelElement(array, "")
}

func GetRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	return random(n, letters)
}

func GetRandomAlapha(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	return random(n, letters)
}

func GetRandomNum(n int) string {
	var letters = []rune("0123456789")
	return random(n, letters)
}

func random(n int, letters []rune) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// return version > version2
func CompareVersion(version1, version2, split string) bool {
	version1Array := strings.Split(version1, split)
	version2Array := strings.Split(version2, split)

	len1 := len(version1Array)
	len2 := len(version2Array)
	len := len1
	if len1 > len2 {
		len = len2
	}

	for i := 0; i < len; i++ {
		if version1Array[i] > version2Array[i] {
			return true
		}
	}
	if strings.HasPrefix(version1, version2) && len1 > len2 {
		return true
	}
	return false
}

func ResolveEscapeCharacters(str string) string {
	if str[0] == '"' {
		str = str[1:]
	}
	length := len(str)
	if str[length-1] == '"' {
		str = str[:length-1]
	}

	for key, value := range escapeMap {
		str = strings.ReplaceAll(str, key, value)
	}
	return str
}

func FormatJson(content string) string {
	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, []byte(content), "", "  "); err != nil {
		return content
	} else {
		return prettyJson.String()
	}
}

func GetColumnData(table [][]string, column int) []string {
	var result []string
	if len(table) <= 1 {
		return []string{}
	}
	for _, v := range table[1:] {
		result = append(result, v[column])
	}

	return result
}

func Contains(str_array []string, target string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}
