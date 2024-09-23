package dao

import (
	"strings"

	"github.com/twbworld/proxy/model/db"
)

type dbUtils struct{}

func (u *dbUtils) getInsertSql(d db.Dbfunc, data map[string]interface{}) (string, []interface{}) {
	if len(data) < 1 {
		return ``, []interface{}{}
	}

	var (
		fields strings.Builder
		values strings.Builder
		sql    strings.Builder
		args   []interface{} = make([]interface{}, 0, len(data))
	)

	//注意map是无序的
	for k, v := range data {
		fields.WriteString("`")
		fields.WriteString(k)
		fields.WriteString("`,")
		values.WriteString("?,")
		args = append(args, v)
	}

	sql.WriteString("INSERT INTO `")
	sql.WriteString(d.TableName())
	sql.WriteString("`(")
	sql.WriteString(strings.TrimRight(fields.String(), `,`))
	sql.WriteString(`) VALUES(`)
	sql.WriteString(strings.TrimRight(values.String(), `,`))
	sql.WriteString(`)`)

	return sql.String(), args
}

func (u *dbUtils) getUpdateSql(d db.Dbfunc, id uint, data map[string]interface{}) (string, []interface{}) {
	if len(data) < 1 {
		return ``, []interface{}{}
	}

	var (
		fields strings.Builder
		sql    strings.Builder
		args   []interface{} = make([]interface{}, 0, len(data))
	)

	for k, v := range data {
		fields.WriteString(" `")
		fields.WriteString(k)
		fields.WriteString("` = ?,")
		args = append(args, v)
	}

	sql.WriteString("UPDATE `")
	sql.WriteString(d.TableName())
	sql.WriteString("` SET")
	sql.WriteString(strings.TrimRight(fields.String(), ","))
	sql.WriteString(" WHERE `id` = ?")
	args = append(args, id)

	return sql.String(), args
}
