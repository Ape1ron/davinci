package core

import (
	"database/sql"
	"davinci/common"
	"davinci/common/log"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"strings"
)

var excludeClickHouseDB = []string{
	"information_schema",
	"default",
	"system",
}

type ClickHouse struct {
	conn   *sql.DB
	Host   string
	Port   int
	User   string
	Passwd string
	DbName string
	Cmd    string
}

func (c *ClickHouse) connect() error {
	var err error
	if c.conn == nil {
		dataSrc := fmt.Sprintf("tcp://%s:%d?username=%s&password=%s&database=%s", c.Host, c.Port, c.User, c.Passwd, c.DbName)

		if c.conn, err = sql.Open("clickhouse", dataSrc); err != nil {
			log.Error(err)
		} else if err = c.conn.Ping(); err != nil {
			log.Error(err)
		}
	}
	return err
}

func (c *ClickHouse) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *ClickHouse) ExecuteOnce() {
	if c.connect() != nil {
		return
	}

	result := c.execute(c.Cmd)
	log.Output(result)
}

func (c *ClickHouse) AutoGather() {
	if c.connect() != nil {
		return
	}

	log.Output(c.getVersion())
	log.Output(c.getUsers())
	databases := c.getDatabases()
	log.Output(databases)
	log.Info(fmt.Sprintf("exclude databbase(built-in): %s", strings.Join(excludeClickHouseDB, ",")))
	for _, database := range common.GetColumnData(databases, 0) {
		if common.Contains(excludeClickHouseDB, strings.ToLower(database)) {
			continue
		}
		tables := c.getTables(database)
		log.Output(tables)
		for _, table := range common.GetColumnData(tables, 0) {
			tableStruct := c.getTableStruct(database, table)
			log.Output(tableStruct)
			dataSum := c.getSumRows(database, table)
			log.Output(dataSum)
			datas := c.getFirst5Rows(database, table)
			log.Output(datas)
		}
	}
}

func (c *ClickHouse) Shell() {

	if c.connect() != nil {
		return
	}
	sqlShell(c.execute)
}

func (c *ClickHouse) SetHost(host string) {
	c.Host = host
}
func (c *ClickHouse) SetPort(port int) {
	c.Port = port
}

func (c *ClickHouse) SetCmd(cmd string) {
	c.Cmd = cmd
}

func (c *ClickHouse) execute(cmd string) [][]string {
	return execute(c.conn, cmd)
}

func (c *ClickHouse) getDatabases() [][]string {
	sql := "show databases;"
	log.Info("get databases")
	return c.execute(sql)
}

func (c *ClickHouse) getTables(dbName string) [][]string {
	sql := fmt.Sprintf("SELECT table_name FROM information_schema.tables where table_schema='%s';", dbName)
	log.Info(fmt.Sprintf("get tables in %s ", dbName))
	return c.execute(sql)
}

func (c *ClickHouse) getColumns(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("SELECT column_name,columns_type FROM information_schema.columns where table_schema='%s' and table_name = '%s';", dbName, tbName)
	return c.execute(sql)
}

func (c *ClickHouse) getTableStruct(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("DESC %s.%s;", dbName, tbName)
	log.Info(fmt.Sprintf("get table [%s] struct", tbName))
	return c.execute(sql)
}

func (c *ClickHouse) getSumRows(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("select count(*) from %s.%s;", dbName, tbName)
	log.Info(fmt.Sprintf("get total number of rows in table [%s]", tbName))
	return c.execute(sql)
}

func (c *ClickHouse) getFirst5Rows(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("select * from %s.%s limit 5;", dbName, tbName)
	log.Info(fmt.Sprintf("get first 5 rows in table [%s]", tbName))
	return c.execute(sql)
}

func (c *ClickHouse) getUsers() [][]string {
	sql := "select * from system.users;"
	log.Info("get users")
	return c.execute(sql)
}

func (c *ClickHouse) getVersion() [][]string {
	sql := "select version();"
	log.Info("get version")
	return c.execute(sql)
}
