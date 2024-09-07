package mongo

import (
	"davinci/common"
	"fmt"
	"testing"
)

func TestFindCloseBrackets(t *testing.T) {
	cmd := "({\"name\":'\"菜鸟教程\\\"123\"456'}).sort({\"_id\":-1}).limit(2)"
	fmt.Println(cmd)
	end, err := findCloseBrackets(cmd)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(end)
		fmt.Println(cmd[:end+1])
	}
}

func TestSplitNotInStr(t *testing.T) {
	content := "({\"name\":{\"$regex\":\"菜鸟.*\"}}).limit(1)"
	array := common.SplitCmd(content, '.')
	fmt.Println(array)
	fmt.Println(len(array))
}
