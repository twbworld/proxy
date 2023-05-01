package initialize

import (
	"encoding/json"
	"log"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/service"
	"github.com/twbworld/proxy/utils"
	"time"
)

//流量上下行的记录清零
func Clear() {

	tx, err := dao.DB.Beginx()
	if err != nil {
		log.Fatalln("开启事务失败[ijhdfakkaop]: ", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Println("事务回滚[orfiujojnmg]: ", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = dao.UpdateUsersClear(tx)
	if err != nil {
		log.Fatalln("清除失败[gpodk]: ", err)
	}
}

//过期用户处理
func Expiry() {
	var (
		users       []model.Users
		usersHandle []model.Users
		ids         []uint
	)

	err := dao.GetUsers(&users, "`quota` != 0 AND `useDays` != 0")

	if err != nil {
		log.Fatalln("查询失败[fsuojnv]: ", err)
	}

	if len(users) < 1 {
		log.Println("没有过期用户[gsfiod]")
		return
	}

	tz, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalln("时区设置错误[pgohkf]: ", err)
	}
	t, err := time.Parse(time.DateTime, time.Now().In(tz).Format(time.DateOnly+" 00:00:01"))
	if err != nil {
		log.Fatalln("时间出错[djaksofja]: ", err)
	}

	for _, value := range users {
		if *value.ExpiryDate == "" || value.Id < 1 {
			continue
		}
		t1, _ := time.Parse(time.DateOnly, *value.ExpiryDate)
		if err != nil {
			continue
		}
		if t.After(t1) {
			usersHandle = append(usersHandle, value)
			ids = append(ids, value.Id)
		}

	}
	if len(usersHandle) < 1 {
		log.Println("没有过期用户[ofijsdfio]")
		return
	}

	tx, err := dao.DB.Beginx()
	if err != nil {
		log.Fatalln("开启事务失败: ", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Println("事务回滚[orjdnmg]: ", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = dao.UpdateUsersExpiry(ids, tx)
	if err != nil {
		log.Fatalln("更新失败[fofiwjm]: ", err)
	}

	log.Println("过期用户处理: ", ids)
	global.Log.Info("过期用户处理: ", ids)

}

//处理用户信息,对mysql数据库作处理
func Handle() {
	var (
		usersJson  []model.UsersInfo
		usersMysql []model.Users
		res        map[string][]string = map[string][]string{"update": {}, "add": {}, "del": {}}
	)

	err := dao.GetUsersByJson(global.Config.AppConfig.UsersPath, &usersJson)
	if err != nil {
		log.Fatalln("读取用户文件错误[odnoskjk]: ", err)
	}
	if len(usersJson) < 1 {
		log.Fatalln("用户文件有误[udldfjos]: ", err)
	}

	err = dao.GetUsers(&usersMysql)
	if err != nil {
		log.Fatalln("查询失败[opfiskjj]: ", err)
	}

	usersMysqlName := utils.ListToMap(usersMysql, "Username")

	tx, err := dao.DB.Beginx()
	if err != nil {
		log.Fatalln("开启事务失败: ", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Println("事务回滚[uirf]: ", err)
			tx.Rollback()
		} else {
			err = tx.Commit()

			log.Printf("%s", res)
			jsonstr, err := json.Marshal(res)
			if err == nil {
				global.Log.Info("汇总: ", string(jsonstr))
				service.TgSend("处理用户" + string(jsonstr))
			}

			Expiry()
		}
	}()

	d1 := 0
	d2 := 30
	for _, user := range usersJson {
		user.Users.UseDays = &d2
		if user.UsersJson.Level > 0 {
			user.Users.UseDays = &d1
		}

		name, ok := usersMysqlName[user.UsersJson.Username]
		if ok {
			//存在于数据库
			if name.(model.Users).PasswordShow != user.Users.PasswordShow || name.(model.Users).Quota != user.UsersJson.Quota || *name.(model.Users).UseDays != *user.Users.UseDays || *name.(model.Users).ExpiryDate != user.UsersJson.ExpiryDate {

				user.Users.Id = name.(model.Users).Id
				err = dao.UpdateUsersHandle(user, tx)

				if err != nil {
					log.Fatalln("更新用户失败[oisdfsm]: ", err)
				}
				jsonstr, err := json.Marshal(name)
				jsonstr2, err2 := json.Marshal(user)
				if err == nil && err2 == nil {
					global.Log.Info("更新用户: ", string(jsonstr)+"=>"+string(jsonstr2))
				}
				res["update"] = append(res["update"], user.UsersJson.Username)
			}

			delete(usersMysqlName, user.UsersJson.Username)
		} else {
			err = dao.AddUsersHandle(user, tx)
			if err != nil {
				log.Fatalln("新增用户失败[poertiflmgo]: ", err)
			}
			jsonstr, err := json.Marshal(user)
			if err == nil {
				global.Log.Info("新增用户: ", string(jsonstr))
			}
			res["add"] = append(res["add"], user.UsersJson.Username)
		}
	}

	if len(usersMysqlName) > 0 {
		for _, value := range usersMysqlName {
			err = dao.DelUsersHandle(value.(model.Users).Id, tx)
			if err != nil {
				log.Fatalln("新增用户失败[poertiflmgo]: ", err)
			}
			jsonstr, err := json.Marshal(value.(model.Users))
			if err == nil {
				global.Log.Info("删除用户: ", string(jsonstr))
			}
			res["del"] = append(res["del"], value.(model.Users).Username)
		}
	}

}
