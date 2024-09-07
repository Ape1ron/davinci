package core

import (
	"context"
	"davinci/common"
	"davinci/common/log"
	mongoutil "davinci/core/mongo"
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/siddontang/go/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

var excludeMongoDB = []string{
	"admin",
	"local",
	"config",
}

type MongoDb struct {
	conn    *mongo.Client
	Host    string
	Port    int
	User    string
	Passwd  string
	DbName  string
	Cmd     string
	context context.Context
}

func (m *MongoDb) connect() error {
	var err error
	if m.conn == nil {
		log.Info("connecting target...")
		m.context = context.Background()
		var dataSrc string
		if m.User == "" || m.Passwd == "" {
			dataSrc = fmt.Sprintf(fmt.Sprintf("mongodb://%s:%d/", m.Host, m.Port))
		} else {
			dataSrc = fmt.Sprintf("mongodb://%s:%s@%s:%d/", m.User, m.Passwd, m.Host, m.Port)
		}
		if m.conn, err = mongo.Connect(m.context, options.Client().ApplyURI(dataSrc).SetConnectTimeout(5*time.Second)); err != nil {
			log.Error(err)
		} else if err = m.conn.Ping(m.context, nil); err != nil {
			log.Error(err)
		}
	}
	return err
}

func (m *MongoDb) Close() {
	if m.conn != nil {
		m.conn.Disconnect(m.context)
		m.conn = nil
	}
}

func (m *MongoDb) ExecuteOnce() {
	if m.connect() != nil {
		return
	}
	result := m.execute(m.Cmd)
	log.Output(result)
	if _, ok := result.(error); ok {
		help()
	}
}

func (m *MongoDb) AutoGather() {
	if m.connect() != nil {
		return
	}
	log.Output(m.getVersion())
	log.Output(m.getUsers())
	databases := m.getDatabases()
	log.Output(databases)
	log.Info(fmt.Sprintf("exclude databbase(built-in): %s", strings.Join(excludeMongoDB, ",")))
	for _, database := range common.GetColumnData(databases, 0) {
		if common.Contains(excludeMongoDB, strings.ToLower(database)) {
			continue
		}
		collections := m.getTables(database)
		log.Output(collections)
		for _, collection := range collections[1:] {
			dataSum := m.getSumDocuments(database, collection)
			log.Output(dataSum)
			datas := m.getFirst5Docs(database, collection)
			log.Output(datas)
		}
	}
}

func (m *MongoDb) Shell() {
	if m.connect() != nil {
		return
	}
	pmt := prompt.New(func(in string) {},
		func(document prompt.Document) []prompt.Suggest {
			return []prompt.Suggest{
				prompt.Suggest{Text: "show {something}", Description: "show some resource info"},
				prompt.Suggest{Text: "use {db_name}", Description: "set current database"},
				prompt.Suggest{Text: "db", Description: "db action"},
				prompt.Suggest{Text: "version", Description: "show version"},
				prompt.Suggest{Text: "exit", Description: "exit shell"},
				prompt.Suggest{Text: "help", Description: "help info"},
			}
		})

	for {
		in := strings.Trim(pmt.Input(), " ")
		if in == "" {
			continue
		}
		if strings.EqualFold(in, "exit") || strings.EqualFold(in, "exit()") {
			break
		} else {
			result := m.execute(in)
			if result != nil {
				log.Output(result)
				if _, ok := result.(error); ok {
					help()
				}
			}
		}

	}
}

func (m *MongoDb) SetHost(host string) {
	m.Host = host
}
func (m *MongoDb) SetPort(port int) {
	m.Port = port
}

func (m *MongoDb) SetCmd(cmd string) {
	m.Cmd = cmd
}

func (m *MongoDb) execute(in string) (result interface{}) {
	for strings.HasSuffix(in, ";") {
		in = in[:len(in)-1]
	}
	if strings.EqualFold(in, "help") {
		help()
	} else if strings.EqualFold(in, "db") {
		result = []string{"result", m.DbName}
	} else if strings.EqualFold(in, "version") {
		result = m.getVersion()
	} else if strings.HasPrefix(in, "show ") {
		resource := strings.Trim(in[5:], " ")
		switch resource {
		case "databases", "dbs":
			result = m.getDatabases()
		case "collections", "tables":
			result = m.getTables(m.DbName)
		case "users":
			result = m.getUsers()
		default:
			result = "not define: " + resource
		}
	} else if strings.HasPrefix(in, "use ") {
		db := strings.Trim(in[4:], " ")
		m.DbName = db
		result = "current database : " + m.DbName
	} else {
		action := mongoutil.ParseActionCmd(in)
		switch action.(type) {
		case *mongoutil.ActionCmd:
			actionCmd := action.(*mongoutil.ActionCmd)
			switch actionCmd.Action.(type) {
			case *mongoutil.SizeAction:
				sizeAction := actionCmd.Action.(*mongoutil.SizeAction)
				result = m.size(actionCmd.Db, actionCmd.Collection, sizeAction.Filter, sizeAction.Opts...)
			case *mongoutil.FindAction:
				findAction := actionCmd.Action.(*mongoutil.FindAction)
				result = m.find(actionCmd.Db, actionCmd.Collection, findAction.Filter, findAction.Opts...)
			case *mongoutil.FindOneAction:
				findOneAction := actionCmd.Action.(*mongoutil.FindOneAction)
				result = m.findOne(actionCmd.Db, actionCmd.Collection, findOneAction.Filter, findOneAction.Opts...)
			default:
				result = fmt.Errorf("the command cannot be parsed")
			}
		case error:
			result = action
		default:
			result = fmt.Errorf("the command cannot be parsed")
		}
	}
	return
}

func help() {
	info := "customize mongodb shell,support special mongo command use for info gather\n"
	info += "show {something}                 'show databases'/'show dbs': Print a list of all available databases.\n"
	info += "                                 'show collections'/'show tables': Print a list of all collections for current database.\n"
	info += "                                 'show users': Print a list of all users for current database.\n"
	info += "use {db_name}                    set current database\n"
	info += "db                               get current database\n"
	info += "db.{col_name}.find()             mongsh find command,support sort,limit and size options\n"
	info += "                                 e.g. db.mycol.find()\n"
	info += "                                 e.g. db.mycol.find().size()\n"
	info += "                                 e.g. db.mycol.find({\"field_name\":{\"$regex\": \"my_regex\"}})\n"
	info += "                                 e.g. db.mycol.find().limit(5)\n"
	info += "                                 e.g. db.mycol.find().sort({\"_id\":-1}).limit(5)\n"
	info += "db.{col_name}.findOne()          mongsh findOne command,support sort options\n"
	info += "                                 e.g. db.mycol.findOne()\n"
	info += "                                 e.g. db.mycol.findOne({\"field_name\":{\"$regex\": \"my_regex\"}})\n"
	info += "                                 e.g. db.mycol.findOne().sort({\"_id\":-1})\n"
	info += "version                          mongodb version\n"
	info += "help                             help info\n"
	info += "exit                             exit shell\n"
	fmt.Println(info)
}

func (m *MongoDb) currentDatabaseCur() *mongo.Database {
	return m.conn.Database(m.DbName)
}

func (m *MongoDb) setCurrentDatabase(dbName string) {
	m.DbName = dbName
}

func (m *MongoDb) getDatabaseCur(dbName string) *mongo.Database {
	return m.conn.Database(dbName)
}

func (m *MongoDb) getDatabases() [][]string {
	m.connect()
	log.Info("get databases: ")
	if listDatabasesResult, err := m.conn.ListDatabases(m.context, bson.M{}); err != nil {
		log.Error(err)
		return nil
	} else {
		var result = [][]string{{"name", "size"}}
		for _, database := range listDatabasesResult.Databases {
			name := database.Name
			size := float64(database.SizeOnDisk) / float64(1024)
			result = append(result, []string{name, strconv.FormatFloat(size, 'f', 2, 64) + "kb"})
		}
		return result
	}
}

func (m *MongoDb) getTables(dbName string) []string {
	m.connect()
	db := m.getDatabaseCur(dbName)
	log.Info(fmt.Sprintf("get collections: [%s]", dbName))
	if tables, err := db.ListCollectionNames(m.context, bson.M{}); err != nil {
		log.Error(err)
		return nil
	} else {
		var result = []string{"name"}
		for _, table := range tables {
			result = append(result, table)
		}
		return result
	}
}

func (m *MongoDb) getSumDocuments(dbName, tbName string) []string {
	log.Info(fmt.Sprintf("get documents size  [%s]", tbName))
	return m.size(dbName, tbName, bson.M{}, []*options.FindOptions{}...)
}

func (m *MongoDb) size(dbName, tbName string, filter bson.M, opts ...*options.FindOptions) []string {
	if dbName == "" {
		dbName = m.DbName
	}
	cur := m.conn.Database(dbName).Collection(tbName)
	if res, err := cur.Find(m.context, filter, opts...); err != nil {
		log.Error(err)
		return nil
	} else {
		defer res.Close(m.context)
		var result = []string{"size", strconv.Itoa(res.RemainingBatchLength())}
		return result
	}
}

func (m *MongoDb) findOne(dbName, tbName string, filter bson.M, opts ...*options.FindOneOptions) []string {
	if dbName == "" {
		dbName = m.DbName
	}
	cur := m.conn.Database(dbName).Collection(tbName)
	singleResult := cur.FindOne(m.context, filter, opts...)
	if result, err := singleResult.Raw(); err != nil {
		log.Error(err)
		return nil
	} else {
		return []string{"result", result.String()}
	}
}

func (m *MongoDb) find(dbName, tbName string, filter bson.M, opts ...*options.FindOptions) []string {
	if dbName == "" {
		dbName = m.DbName
	}
	res := m.conn.Database(dbName).Collection(tbName)
	if cur, err := res.Find(m.context, filter, opts...); err != nil {
		log.Error(err)
		return nil
	} else {
		defer cur.Close(m.context)
		var result = []string{"result"}
		result = append(result, m.getDocuments(cur)...)
		return result
	}
}

func (m *MongoDb) getDocuments(cur *mongo.Cursor) []string {
	defer cur.Close(m.context)
	defer cur.Close(m.context)
	var result = []string{}
	for cur.Next(m.context) {
		var document bson.M
		if err := cur.Decode(&document); err != nil {
			log.Error(err)
		} else {
			jsonBytes, _ := json.Marshal(document)
			result = append(result, string(jsonBytes))
		}
	}
	return result
}

func (m *MongoDb) getFirst5Docs(dbName, tbName string) []string {
	log.Info(fmt.Sprintf("get first 5 documents [%s]", tbName))
	limit := int64(5)
	return m.find(dbName, tbName, bson.M{}, &options.FindOptions{
		Limit: &limit,
	})
}

func (m *MongoDb) getUsers() []string {
	log.Info(fmt.Sprintf("get users"))
	res := m.conn.Database("admin").Collection("system.users")
	//return m.find("admin", "system.users", bson.M{}, nil)
	if cur, err := res.Find(m.context, bson.M{}); err != nil {
		log.Error(err)
		return nil
	} else {
		defer cur.Close(m.context)
		var result = []string{"user"}
		result = append(result, m.getDocuments(cur)...)
		return result
	}
}

func (m *MongoDb) getVersion() []string {
	log.Info("get version")
	db := m.conn.Database("test")
	var result = []string{"version"}
	var document bson.M
	var command = bson.M{}
	command["buildInfo"] = 1
	if err := db.RunCommand(m.context, command).Decode(&document); err != nil {
		log.Error(err)
	} else {
		if version, ok := document["version"]; ok {
			result = append(result, version.(string))
		}

	}
	return result
}

//func (m *MongoDb) eval(dbName, cmd string) []string {
//	fmt.Println(fmt.Sprintf("[info] use eval command to execute javascript (eval command remove before mongodb 4.0)"))
//	db := m.conn.Database(dbName)
//	var result = []string{}
//	var document bson.M
//	var command = bson.M{}
//	command["eval"] = cmd
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	if err := db.RunCommand(ctx, command).Decode(&document); err != nil {
//		fmt.Println(err)
//	} else {
//		jsonBytes, _ := json.Marshal(document)
//		result = append(result, string(jsonBytes))
//	}
//	return result
//}
