package core

import (
	"davinci/common/log"
	mongo2 "davinci/core/mongo"
	"encoding/json"
	"fmt"
	"github.com/siddontang/go/bson"
	"reflect"
	"runtime/debug"
	"testing"
)

func TestName(t *testing.T) {
	mongo := &MongoDb{
		Host:   "192.168.159.135",
		Port:   27017,
		User:   "",
		Passwd: "",
		DbName: "",
		Cmd:    "",
	}
	defer mongo.Close()
	//mongo.execute("")
	databases := mongo.getDatabases()
	log.Output(databases)
	tables := mongo.getTables("admin")
	log.Output(tables)
	sums := mongo.getSumDocuments("runoob", "runoob")
	log.Output(sums)
	singleResult := mongo.findOne("qqqqq", "runoob", bson.M{})
	log.Output(singleResult)
	users := mongo.getUsers()
	log.Output(users)
	datas := mongo.getFirst5Docs("runoob", "runoob")
	log.Output(datas)
	//finddatas := mongo.find("runoob", "runoob", bson.M{"name": bson.M{"$regex": "菜鸟.*"}})
	var filter bson.M
	err := json.Unmarshal([]byte("{\"$or\":[{\"name\": \"菜鸟教程\"}, {\"title\": \"MongoDB 教程\"}]}"), &filter)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(filter)
	//finddatas := mongo.find("runoob", "runoob", bson.M{"$or": []bson.M{{"name": "菜鸟教程"}, {"title": "MongoDB 教程"}}})
	finddatas := mongo.find("runoob", "runoob", filter)
	log.Output(finddatas)
}

func TestFind(t *testing.T) {
	//cmd := "db.runoob.find({\"name\":{\"$regex\":\"菜鸟.*\"}}).sort({\"_id\":-1}).limit(2).sort({\"_id\":1})"
	cmd := "runoob.runoob.find()"
	actioncmd := mongo2.ParseActionCmd(cmd)
	mongo := &MongoDb{
		Host:   "192.168.159.135",
		Port:   27017,
		User:   "",
		Passwd: "",
		DbName: "",
		Cmd:    "",
	}
	mongo.getDatabases()
	defer mongo.Close()
	//action := actioncmd.Action.(*mongo2.FindAction)
	switch actioncmd.(type) {
	case *mongo2.ActionCmd:
		actioncmd2 := actioncmd.(*mongo2.ActionCmd)
		action := actioncmd2.Action.(*mongo2.FindAction)
		datas := mongo.find(actioncmd2.Db, actioncmd2.Collection, action.Filter, action.Opts...)
		log.Output(datas)
	case error:
		fmt.Println(fmt.Sprintf("error: %v", actioncmd))
	default:
		fmt.Println("default")
		fmt.Println(reflect.TypeOf(actioncmd))
		fmt.Println(actioncmd)

	}

}

func TestFindOne(t *testing.T) {
	cmd := "db.runoob.findOne({\"name\":{\"$regex\":\"菜鸟.*\"}}).sort({\"_id\":1}).sort()"
	//cmd := "db.runoob.find().sort().sort().limit(2)"
	actioncmd := mongo2.ParseActionCmd(cmd)
	mongo := &MongoDb{
		Host:   "192.168.159.135",
		Port:   27017,
		User:   "",
		Passwd: "",
		DbName: "",
		Cmd:    "",
	}
	mongo.getDatabases()
	defer mongo.Close()
	switch actioncmd.(type) {
	case *mongo2.ActionCmd:
		actioncmd2 := actioncmd.(*mongo2.ActionCmd)
		action := actioncmd2.Action.(*mongo2.FindOneAction)
		datas := mongo.findOne("runoob", actioncmd2.Collection, action.Filter, action.Opts...)
		log.Output(datas)
	case error:
		fmt.Printf("error: %v\n", actioncmd)
		fmt.Println(string(debug.Stack()))
	default:
		fmt.Println("default")
		fmt.Println(reflect.TypeOf(actioncmd))
		fmt.Println(actioncmd)

	}

}

//func TestEval(t *testing.T) {
//	mongo := &MongoDb{
//		Host:   "192.168.159.135",
//		Port:   27020,
//		User:   "",
//		Passwd: "",
//		DbName: "",
//		Cmd:    "",
//	}
//	mongo.getDatabases()
//	defer mongo.Close()
//	cmd := "db.runoob.find();"
//	db := "admin"
//	result := mongo.eval(db, cmd)
//	fmt.Println(result)
//}

func TestVersion(t *testing.T) {
	mongo := &MongoDb{
		Host:   "192.168.159.135",
		Port:   27020,
		User:   "",
		Passwd: "",
		DbName: "",
		Cmd:    "",
	}
	mongo.connect()
	version := mongo.getVersion()
	log.Output(version)
}
