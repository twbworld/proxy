package initialize

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/service"
)

var tx *sqlx.Tx

func transaction() func() {
	var err error
	if tx, err = dao.DB.Beginx(); err != nil {
		panic(err)
	}

	return func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if e := tx.Commit(); e != nil {
			tx.Rollback()
			panic(e)
		} else {
			global.Log.Println("成功")
		}
	}
}

// 流量上下行的记录清零
func Clear() {

	defer transaction()()

	if err := dao.UpdateUsersClear(tx); err != nil {
		panic(err)
	}

}

// 过期用户处理
func Expiry() {
	var (
		users []model.Users
	)

	if err := dao.GetUsers(&users, "`quota` != 0 AND `useDays` != 0"); err != nil {
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
	t, err := time.ParseInLocation(time.DateTime, time.Now().In(tz).Format(time.DateOnly+" 00:00:01"), tz)
	if err != nil {
		panic("时间出错[djaksofja]: " + err.Error())
	}
	t1, t2 := time.Now().In(tz).AddDate(0, 0, -7), time.Now().In(tz).AddDate(0, 0, -5)

	ids := make([]uint, 0, len(users))
	for _, value := range users {
		if *value.ExpiryDate == "" || value.Id < 1 {
			continue
		}
		ti, _ := time.ParseInLocation(time.DateOnly, *value.ExpiryDate, tz)
		if err != nil {
			continue
		}
		if t.After(ti) {
			ids = append(ids, value.Id)
		}

		if t1.After(ti) && t2.Before(ti) {
			service.TgSend(value.Username + "快到期" + ti.Format(time.DateOnly))
		}

	}
	if len(ids) < 1 {
		global.Log.Infoln("没有过期用户[ofijsdfio]")
		return
	}

	defer transaction()()

	if err := dao.UpdateUsersExpiry(&ids, tx); err != nil {
		panic(err)
	}

	global.Log.Infoln("[Expiry]过期用户处理: ", ids)
}
