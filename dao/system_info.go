package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/model/db"
)

type SystemInfoDb struct{}

func (s *SystemInfoDb) SaveSysVal(key, value string, tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[ios58ja]")
	}

	if err := s.CheckSysVal(key, tx); err != nil {
		return err
	}

	tn := db.SystemInfo{}.TableName()

	var sqlStr string
	if CanLock {
		sqlStr = fmt.Sprintf("SELECT `id` FROM `%s` WHERE `key`=? FOR UPDATE", tn)
		if _, err = tx.Exec(sqlStr, key); err != nil {
			return fmt.Errorf("[fui6u]%s", err)
		}
	}

	sqlStr = fmt.Sprintf("UPDATE `%s` SET `value` = ?, `updated_at` = ? WHERE `key` = ?", tn)
	_, err = tx.Exec(sqlStr, value, time.Now().Unix(), key)
	return
}

func (s *SystemInfoDb) CheckSysVal(key string, tx *sqlx.Tx) error {
	info := new(db.SystemInfo)
	if App.GetByCol(info, "key", key, tx) == sql.ErrNoRows {
		// 泛型方法自动推断出正确的表和结构
		return App.Insert[db.SystemInfo](map[string]any{"key": key}, tx)
	}
	return nil
}
