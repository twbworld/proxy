package initialize

import (
	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
)

// 流量上下行的记录清零
func Clear() {
	err := dao.Tx(func(tx *sqlx.Tx) (e error) {
		e = dao.App.UsersDb.UpdateUsersClear(tx)
		return
	})

	if err != nil {
		panic(err)
	}
}
