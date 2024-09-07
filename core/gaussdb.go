package core

import (
	"davinci/common"
	"davinci/common/log"
	"fmt"
	"strings"
)

var excludeGsDB = []string{
	"template1",
	"template0",
}

var excludeGsSchema = []string{
	"pg_toast",
	"cstore",
	"pkg_service",
	"dbe_perf",
	"snapshot",
	"blockchain",
	"pg_catalog",
	"sqladvisor",
	"dbe_pldebugger",
	"dbe_pldeveloper",
	"dbe_sql_util",
	"information_schema",
	"db4ai",
}

type GaussDB struct {
	*Pgsql
}

func (g *GaussDB) AutoGather() {
	if g.connect() != nil {
		return
	}

	log.Output(g.getVersion())
	log.Output(g.getUsers())
	databases := g.getDatabases()
	log.Output(databases)

	log.Info(fmt.Sprintf("exclude database(built-in): %s", strings.Join(excludeGsDB, ",")))
	for _, db := range common.GetColumnData(databases, 0) {
		if common.Contains(excludeGsDB, strings.ToLower(db)) {
			continue
		}
		g.gatherInOneDb(db)
	}
}

//
//func (g *GaussDB) SetHost(host string) {
//	g.Host = host
//}
//func (g *GaussDB) SetPort(port int) {
//	g.Port = port
//}
//
//func (g *GaussDB) SetCmd(cmd string) {
//	g.Cmd = cmd
//}

func (g *GaussDB) gatherInOneDb(dbName string) {
	g.Close()
	g.DbName = dbName
	if g.connect() != nil {
		return
	}
	currentDb := g.getCurrentDb()
	log.Output(currentDb)
	dbSzie := g.getDatabaseSize(dbName)
	log.Output(dbSzie)
	schemas := g.getSchemas()
	log.Output(schemas)

	log.Info(fmt.Sprintf("exclude schemas(built-in): %s", strings.Join(excludeGsSchema, ",")))
	for _, schema := range common.GetColumnData(schemas, 0) {
		if common.Contains(excludeGsSchema, strings.ToLower(schema)) {
			continue
		}
		tables := g.getTables(schema)
		log.Output(tables)
		for _, table := range common.GetColumnData(tables, 0) {
			tableStruct := g.getTableStruct(schema, table)
			log.Output(tableStruct)
			dataSum := g.getSumRows(schema, table)
			log.Output(dataSum)
			datas := g.getFirst5Rows(schema, table)
			log.Output(datas)
		}
	}
}
