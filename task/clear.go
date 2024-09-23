package task

import (
	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
)

// 流量上下行的记录清零
func Clear() error {
	err := dao.Tx(func(tx *sqlx.Tx) (e error) {
		e = dao.App.UsersDb.UpdateUsersClear(tx)
		return
	})

	if err != nil {
		global.Log.Errorln("清除失败[weij0]")
		return err
	}

	global.Log.Infoln("成功清零流量")
	return nil
}
