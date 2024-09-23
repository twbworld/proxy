package initialize

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/service"
)

// 过期用户处理
func Expiry() {
	var users []db.Users

	if err := dao.App.UsersDb.GetUsers(&users); err != nil {
		global.Log.Fatalf("查询失败[fsuojnv]: %v", err)
	}

	if len(users) < 1 {
		global.Log.Infoln("没有过期用户[gsfiod]")
		return
	}

	now := time.Now().In(global.Tz)
	t, err := time.ParseInLocation(time.DateTime, now.Format(time.DateOnly+" 00:00:01"), global.Tz)
	if err != nil {
		panic("时间出错[djaksofja]: " + err.Error())
	}
	t1, t2 := now.AddDate(0, 0, -7), time.Now().In(global.Tz).AddDate(0, 0, -5)

	ids := make([]uint, 0, len(users))
	for _, user := range users {
		if user.ExpiryDate == nil || *user.ExpiryDate == "" || user.Id < 1 {
			continue
		}
		ti, err := time.ParseInLocation(time.DateOnly, *user.ExpiryDate, global.Tz)
		if err != nil {
			continue
		}
		if t.After(ti) {
			ids = append(ids, user.Id)
		}

		if t1.After(ti) && t2.Before(ti) {
			go service.Service.UserServiceGroup.TgService.TgSend(user.Username + "快到期" + ti.Format(time.DateOnly))
		}
	}
	if len(ids) == 0 {
		global.Log.Infoln("没有过期用户[ofijsdfio]")
		return
	}

	err = dao.Tx(func(tx *sqlx.Tx) (e error) {
		return dao.App.UsersDb.UpdateUsersExpiry(ids, tx)
	})

	if err != nil {
		global.Log.Fatalf("更新用户过期状态失败: %v", err)
	}

	global.Log.Infoln("[Expiry]过期用户处理: ", ids)
}
