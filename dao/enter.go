package dao

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

//PS:使用Get()需考虑sql.ErrNoRows的情况

var (
	DB      *sqlx.DB
	CanLock bool     //是否支持锁(FOR UPDATE)
	App     *DbGroup = new(DbGroup)
)

type DbGroup struct {
	SystemInfoDb
	UsersDb
}

func Tx(fc func(tx *sqlx.Tx) error) (err error) {
	tx, err := DB.Beginx()
	if err != nil {
		return fmt.Errorf("系统错误(开启事务失败): %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // 保留 Panic 原味抛出
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fc(tx)
	return
}
