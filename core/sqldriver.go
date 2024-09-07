package core

import (
	"database/sql"
	"davinci/common/log"
	"github.com/c-bata/go-prompt"
	"strings"
)

type execSqlFunc func(cmd string) [][]string

func sqlShell(exexSql execSqlFunc) {

	pmt := prompt.New(func(in string) {},
		func(document prompt.Document) []prompt.Suggest {
			return nil
		})

	for {
		in := pmt.Input()
		if strings.EqualFold(in, "exit") || strings.EqualFold(in, "exit()") {
			break
		}
		if strings.TrimSpace(in) == "" {
			continue
		}
		result := exexSql(in)
		log.Output(result)
	}
}

func execute(conn *sql.DB, cmd string) [][]string {
	if conn == nil {
		return nil
	}
	var result [][]string
	defer func() {
		err := recover()
		if err != nil && err.(error).Error() == "runtime error: invalid memory address or nil pointer dereference" {
		}
	}()
	log.Info("execute sql: " + cmd)
	query, err := conn.Query(cmd)
	if err != nil {
		log.Error(err)
		return nil
	}

	cols, _ := query.Columns()
	result = append(result, cols)
	colNums := len(cols)
	values := make([][]byte, colNums)
	scans := make([]interface{}, colNums)
	for i := range values {
		scans[i] = &values[i]
	}
	for query.Next() {
		query.Scan(scans...)

		row := make([]string, colNums)

		for k, v := range values {
			row[k] = string(v)
		}
		result = append(result, row)
	}

	return result
}
