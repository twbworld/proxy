package initialize

import (
	"time"

	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/service"
)

// 流量上下行的记录清零
func Clear() {

	tx, err := dao.DB.Beginx()
	if err != nil {
		panic("开启事务失败[ijhdfakkaop]: " + err.Error())
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
			panic("事务回滚[orfiujojnmg]: " + err.Error())
		} else {
			err = tx.Commit()
			if err != nil {
				panic("错误[sgfjhios]: " + err.Error())
			}
			global.Log.Println("[Clear]成功")
		}
	}()

	err = dao.UpdateUsersClear(tx)

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
		panic("查询失败[fsuojnv]: " + err.Error())
	}

	if len(users) < 1 {
		global.Log.Infoln("没有过期用户[gsfiod]")
		return
	}

	tz, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic("时区设置错误[pgohkf]: " + err.Error())
	}
	t, err := time.Parse(time.DateTime, time.Now().In(tz).Format(time.DateOnly+" 00:00:01"))
	if err != nil {
		panic("时间出错[djaksofja]: " + err.Error())
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
		global.Log.Infoln("没有过期用户[ofijsdfio]")
		return
	}

	tx, err := dao.DB.Beginx()
	if err != nil {
		panic("开启事务失败[woiosd]: " + err.Error())
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
			panic("事务回滚[orjdnmg]: " + err.Error())
		} else {
			err = tx.Commit()
			if err != nil {
				panic("错误[opiakjf]: " + err.Error())
			}
			global.Log.Infoln("[Expiry]过期用户处理: ", ids)
		}
	}()

	err = dao.UpdateUsersExpiry(ids, tx)


}
