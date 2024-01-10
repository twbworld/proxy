package initialize

import (
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/service"
	"time"
)

// 流量上下行的记录清零
func Clear() {

	tx, err := dao.DB.Beginx()
	if err != nil {
		global.Log.Panicln("开启事务失败[ijhdfakkaop]: ", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			global.Log.Warn("事务回滚[orfiujojnmg]: ", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = dao.UpdateUsersClear(tx)
	if err != nil {
		global.Log.Panicln("清除失败[gpodk]: ", err)
	}
}

// 过期用户处理
func Expiry() {
	var (
		users       []model.Users
		usersHandle []model.Users
		ids         []uint
	)

	err := dao.GetUsers(&users, "`quota` != 0 AND `useDays` != 0")

	if err != nil {
		global.Log.Panicln("查询失败[fsuojnv]: ", err)
	}

	if len(users) < 1 {
		global.Log.Info("没有过期用户[gsfiod]")
		return
	}

	tz, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		global.Log.Panicln("时区设置错误[pgohkf]: ", err)
	}
	t, err := time.Parse(time.DateTime, time.Now().In(tz).Format(time.DateOnly+" 00:00:01"))
	if err != nil {
		global.Log.Panicln("时间出错[djaksofja]: ", err)
	}
	t1 := time.Now().In(tz).AddDate(0, 0, -7)
	t2 := time.Now().In(tz).AddDate(0, 0, -5)

	for _, value := range users {
		if *value.ExpiryDate == "" || value.Id < 1 {
			continue
		}
		ti, _ := time.Parse(time.DateOnly, *value.ExpiryDate)
		if err != nil {
			continue
		}
		if t.After(ti) {
			usersHandle = append(usersHandle, value)
			ids = append(ids, value.Id)
		}

		if t1.After(ti) && t2.Before(ti) {
			service.TgSend(value.Username + "快到期" + ti.In(tz).Format(time.DateOnly))
		}

	}
	if len(usersHandle) < 1 {
		global.Log.Info("没有过期用户[ofijsdfio]")
		return
	}

	tx, err := dao.DB.Beginx()
	if err != nil {
		global.Log.Panicln("开启事务失败: ", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			global.Log.Warn("事务回滚[orjdnmg]: ", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = dao.UpdateUsersExpiry(ids, tx)
	if err != nil {
		global.Log.Panicln("更新失败[fofiwjm]: ", err)
	}

	global.Log.Info("过期用户处理: ", ids)

}
