package dao

import (
	"database/sql"
	"sync"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
	"github.com/twbworld/proxy/utils"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

const (
	SysTimeOut int64   = 30         //流程码过期时间,单位s
	QuotaMax   float64 = 1073741824 //流量单位转换,入库需要, 1G*1024*1024*1024 = 1073741824byte
)

var mu sync.Mutex

func GetUsers(users *[]model.Users, where ...string) error {
	sql := "SELECT * FROM `users` WHERE 1=1"
	if len(where) > 0 {
		sql += " AND " + where[0]
	}

	return DB.Select(users, sql)
}

func GetUserNames(users *[]model.Users, where ...string) error {
	sql := "SELECT `id`, `username` FROM `users` WHERE 1=1"
	if len(where) > 0 {
		sql += " AND " + where[0]
	}

	return DB.Select(users, sql)
}

func GetUsersByUserName(users *model.Users, userName string) error {
	sql := "SELECT * FROM `users` WHERE `username`=?"
	return DB.Get(users, sql, userName)
}
func GetUsersByUserNameTx(tx *sqlx.Tx, users *model.Users, userName string) error {
	sql := "SELECT * FROM `users` WHERE `username`=?"
	return tx.Get(users, sql, userName)
}
func GetUsersByUserId(users *model.Users, id uint) error {
	sql := "SELECT * FROM `users` WHERE `id`=?"
	return DB.Get(users, sql, id)
}

func UpdateUsers(tx *sqlx.Tx, user *model.Users) (err error) {
	sql := "SELECT `id` FROM `users` WHERE `id`=? FOR UPDATE"
	if global.Config.Env.Db.Type == "sqlite" {
		sql = "SELECT `id` FROM `users` WHERE `id`=?"
	}
	if _, err = tx.Exec(sql, user.Id); err != nil {
		return
	}

	sql = "UPDATE `users` SET `username` = :username, `password` = :password, `passwordShow` = :passwordShow, `quota` = :quota, `download` = :download, `upload` = :upload, `useDays` = :useDays, `expiryDate` = :expiryDate WHERE `id` = :id"
	_, err = tx.NamedExec(sql, map[string]interface{}{
		"id":           user.Id,
		"username":     user.Username,
		"password":     user.Password,
		"passwordShow": user.PasswordShow,
		"quota":        user.Quota,
		"download":     user.Download,
		"upload":       user.Upload,
		"useDays":      user.UseDays,
		"expiryDate":   user.ExpiryDate,
	})

	return
}

func UpdateUsersClear(tx *sqlx.Tx) (err error) {
	mu.Lock()
	defer mu.Unlock()

	sql := "LOCK TABLE `users` WRITE"
	if global.Config.Env.Db.Type == "mysql" {
		if _, err = tx.Exec(sql); err != nil {
			return
		}
	}

	sql = "UPDATE `users` SET `download` = :download, `upload` = :upload"
	_, err = tx.NamedExec(sql, gin.H{"download": 0, "upload": 0})
	if err != nil {
		return
	}

	sql = "UNLOCK TABLES"
	if global.Config.Env.Db.Type == "mysql" {
		_, err = tx.Exec(sql)
	}

	return
}

func UpdateUsersExpiry(ids *[]uint, tx *sqlx.Tx) (err error) {
	sql := "SELECT `id` FROM `users` WHERE `id` IN (?) FOR UPDATE"
	if global.Config.Env.Db.Type == "sqlite" {
		sql = "SELECT `id` FROM `users` WHERE `id` IN (?)"
	}
	query, args, err := sqlx.In(sql, *ids)
	if err != nil {
		return
	}
	if _, err = tx.Exec(tx.Rebind(query), args...); err != nil {
		return
	}

	sql = "UPDATE `users` SET `quota` = 0 WHERE `id` IN (?)"
	query, args, err = sqlx.In(sql, *ids)
	if err != nil {
		return
	}
	_, err = tx.Exec(tx.Rebind(query), args...)
	return
}

func InsertEmptyUsers(tx *sqlx.Tx, userName string) (err error) {
	sql := "INSERT INTO `users`(`username`, `password`, `passwordShow`, `quota`, `download`, `upload`, `useDays`, `expiryDate`) VALUES(:username, :password, :passwordShow, :quota, :download, :upload, :useDays, :expiryDate)"

	_, err = tx.NamedExec(sql, map[string]interface{}{
		"username":     userName,
		"password":     utils.Hash(userName),
		"passwordShow": utils.Base64Encode(utils.Hash(userName)),
		"quota":        int(50 * QuotaMax),
		"expiryDate":   time.Now().AddDate(0, 1, 0).Format(time.DateOnly),
		"useDays":      30,
		"download":     0,
		"upload":       0,
	})
	return
}

func DelUsersHandle(id uint, tx *sqlx.Tx) (err error) {
	sql := "DELETE FROM `users` WHERE `id`=?"
	_, err = tx.Exec(sql, id)
	return
}

func GetSysValByKey(SystemInfo *model.SystemInfo, key string) error {
	sql := "SELECT * FROM `system_info` WHERE `key`=?"
	return DB.Get(SystemInfo, sql, key)
}

func SaveSysVal(tx *sqlx.Tx, key string, value string) (err error) {
	if err = CheckSysVal(tx, key); err != nil {
		return
	}

	sql := "SELECT `id` FROM `system_info` WHERE `key`=? FOR UPDATE"
	if global.Config.Env.Db.Type == "sqlite" {
		sql = "SELECT `id` FROM `system_info` WHERE `key`=?"
	}

	if _, err = tx.Exec(sql, key); err != nil {
		return
	}

	sql = "UPDATE `system_info` SET `value` = :value, `update_time` = CURRENT_TIMESTAMP WHERE `key` = :key"
	_, err = tx.NamedExec(sql, map[string]interface{}{
		"value": value,
		"key":   key,
	})
	return
}
func CheckSysVal(tx *sqlx.Tx, key string) (err error) {
	var info model.SystemInfo
	if err = GetSysValByKey(&info, key); err == sql.ErrNoRows {
		//空则创建
		sql := "INSERT INTO `system_info`(`key`) VALUES(:key)"
		_, err = tx.NamedExec(sql, map[string]interface{}{
			"key": key,
		})
	}

	return
}
