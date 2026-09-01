package dao

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/model/db"
)

// DAO 层内部维护的反射字段缓存
var columnCache sync.Map

// getColumns 解析结构体的 db tag 字段, 避免 SELECT *
func (g *DbGroup) getColumns[T db.Dbfunc]() string {
	typ := reflect.TypeFor[T]()
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if val, ok := columnCache.Load(typ); ok {
		return val.(string)
	}

	var cols []string
	for field := range typ.Fields() {
		// 处理内嵌结构体（如 BaseField）
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			for subField := range field.Type.Fields() {
				if tag := subField.Tag.Get("db"); tag != "" && tag != "-" {
					cols = append(cols, "`"+tag+"`")
				}
			}
			continue
		}
		if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
			cols = append(cols, "`"+tag+"`")
		}
	}

	colStr := strings.Join(cols, ", ")
	columnCache.Store(typ, colStr)
	return colStr
}

func (g *DbGroup) GetByCol[T db.Dbfunc](dest *T, col string, val any, tx ...*sqlx.Tx) error {
	var t T
	query := fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=? LIMIT 1", g.getColumns[T](), t.TableName(), col)
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Get(dest, query, val)
	}
	return DB.Get(dest, query, val)
}

// SelectWhere 通用条件批量查询
func (g *DbGroup) SelectWhere[T db.Dbfunc](dest *[]T, where string, args []any, tx ...*sqlx.Tx) error {
	var t T
	query := fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", g.getColumns[T](), t.TableName(), where)
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Select(dest, query, args...)
	}
	return DB.Select(dest, query, args...)
}

func (g *DbGroup) Insert[T db.Dbfunc](data map[string]any, tx *sqlx.Tx) error {
	if len(data) < 1 {
		return nil
	}

	// 应用代码层统一注入创建与更新时间
	tags := db.GetBaseFieldDbTags()
	now := time.Now().Unix()
	if _, ok := data[tags.CreatedAtDbTag]; !ok && tags.CreatedAtDbTag != "" {
		data[tags.CreatedAtDbTag] = now
	}
	if _, ok := data[tags.UpdatedAtDbTag]; !ok && tags.UpdatedAtDbTag != "" {
		data[tags.UpdatedAtDbTag] = now
	}

	var (
		t      T
		fields strings.Builder
		values strings.Builder
		args   = make([]any, 0, len(data))
	)

	for k, v := range data {
		fields.WriteString("`")
		fields.WriteString(k)
		fields.WriteString("`,")
		values.WriteString("?,")
		args = append(args, v)
	}

	query := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s)",
		t.TableName(),
		strings.TrimRight(fields.String(), ","),
		strings.TrimRight(values.String(), ","),
	)

	_, err := tx.Exec(query, args...)
	return err
}

func (g *DbGroup) UpdateById[T db.Dbfunc](id uint, data map[string]any, tx *sqlx.Tx) error {
	if len(data) < 1 {
		return nil
	}

	tags := db.GetBaseFieldDbTags()
	if _, ok := data[tags.UpdatedAtDbTag]; !ok && tags.UpdatedAtDbTag != "" {
		data[tags.UpdatedAtDbTag] = time.Now().Unix()
	}

	var (
		t      T
		fields strings.Builder
		args   = make([]any, 0, len(data)+1)
	)

	for k, v := range data {
		fmt.Fprintf(&fields, "`%s` = ?,", k)
		args = append(args, v)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `id` = ?",
		t.TableName(),
		strings.TrimRight(fields.String(), ","),
	)

	_, err := tx.Exec(query, args...)
	return err
}

// DeleteById 泛型按主键删除
func (g *DbGroup) DeleteById[T db.Dbfunc](id uint, tx *sqlx.Tx) error {
	var t T
	query := fmt.Sprintf("DELETE FROM `%s` WHERE `id`=?", t.TableName())
	_, err := tx.Exec(query, id)
	return err
}

// LockById 泛型行级排他锁 (FOR UPDATE)
func (g *DbGroup) LockById[T db.Dbfunc](id uint, tx *sqlx.Tx) error {
	if !CanLock {
		return nil
	}
	var t T
	query := fmt.Sprintf("SELECT `id` FROM `%s` WHERE `id`=? FOR UPDATE", t.TableName())
	_, err := tx.Exec(query, id)
	return err
}
