package mongo

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	//cmd := "db.aaa.bbb.ccc.find({\"name\":{\"$regex\":\"菜鸟.*\"}})"
	cmd := "db.runoob.find().sort().sort().limit(1)"
	actioncmd := ParseActionCmd(cmd).(*ActionCmd)
	fmt.Println(actioncmd.Db)
	fmt.Println(actioncmd.Collection)
	fmt.Println(actioncmd.Action)

}
