package dao

import (
	"errors"
	"fmt"

	"database/sql"
	"github.com/twbworld/proxy/model/db"

	"github.com/jmoiron/sqlx"
)

type SystemInfoDb struct{}

func (s *SystemInfoDb) GetSysValByKey(SystemInfo *db.SystemInfo, key string, tx ...*sqlx.Tx) error {
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE `key` = ?", SystemInfo.TableName())
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Get(SystemInfo, sql, key)
	}
	return DB.Get(SystemInfo, sql, key)
}

func (s *SystemInfoDb) SaveSysVal(key string, value string, tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[ios58ja]")
	}

	if err := s.CheckSysVal(key, tx); err != nil {
		return err
	}

	tn := db.SystemInfo{}.TableName()

	var sql string
	if CanLock {
		sql = fmt.Sprintf("SELECT `id` FROM `%s` WHERE `key`=? FOR UPDATE", tn)
		if _, err = tx.Exec(sql, key); err != nil {
			return fmt.Errorf("[fui6u]%s", err)
		}
	}

	sql = fmt.Sprintf("UPDATE `%s` SET `value` = ?, `update_time` = CURRENT_TIMESTAMP WHERE `key` = ?", tn)
	_, err = tx.Exec(sql, value, key)
	return
}

func (s *SystemInfoDb) CheckSysVal(key string, tx *sqlx.Tx) (err error) {
	var info db.SystemInfo
	if s.GetSysValByKey(&info, key, tx) == sql.ErrNoRows {
		sql, args := utils.getInsertSql(db.SystemInfo{}, map[string]interface{}{
			"key": key,
		})
		_, err = tx.Exec(sql, args...)
	}

	return
}
