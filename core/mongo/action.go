package mongo

import (
	"davinci/common"
	"fmt"
	"strings"
)

const (
	FindActionName    = "find"
	FindOneActionName = "findOne"
)

const (
	Sort  = "sort"
	Limit = "limit"
	Size  = "size"
)

type Action interface {
}

type ActionCmd struct {
	Db         string
	Collection string
	Action     Action
}

func ParseActionCmd(cmd string) (result interface{}) {

	firstActionArea := common.FindFirstCharNotInStr(cmd, '(')
	if firstActionArea == -1 {
		return cmd
	}
	rscDesc := cmd[0:firstActionArea]
	if strings.Count(rscDesc, ".") < 2 {
		result = fmt.Errorf("not found command")
		return
	}
	startActionIndex := strings.LastIndex(rscDesc, ".")
	splitDbIndex := strings.Index(rscDesc, ".")

	db := cmd[0:splitDbIndex]
	if db == "db" {
		db = ""
	}
	collection := cmd[splitDbIndex+1 : startActionIndex]
	actionName := cmd[startActionIndex+1 : firstActionArea]
	actionCmd := cmd[firstActionArea:]
	//log.Info("database:" + db)
	//log.Info("collection:" + collection)
	//log.Info("actionName:" + actionName)
	//log.Info("actionCmd:" + actionCmd)

	var action Action
	defer func() {
		if err := recover(); err != nil {
			result = err
		}
	}()
	switch actionName {
	case FindActionName:
		action = parseFindActions(actionCmd)
	case FindOneActionName:
		action = parseFindOneActions(actionCmd)
	default:
		result = fmt.Errorf("do not find action: %s, please check command", actionName)
	}

	result = &ActionCmd{
		Db:         db,
		Collection: collection,
		Action:     action,
	}
	return
}
