package dao

import (
	"errors"
	"fmt"
	"time"

	"sync"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"

	tool "github.com/twbworld/proxy/utils"

	"github.com/jmoiron/sqlx"
)

type UsersDb struct{}

const (
	SysTimeOut int64   = 30         //流程码过期时间,单位s
	QuotaMax   float64 = 1073741824 //流量单位转换,入库需要, 1G*1024*1024*1024 = 1073741824byte
)

var mu sync.Mutex

func (u *UsersDb) GetUsers(users *[]db.Users, tx ...*sqlx.Tx) error {
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE `quota` != 0 AND `useDays` != 0", db.Users{}.TableName())
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Select(users, sql)
	}

	return DB.Select(users, sql)
}

func (u *UsersDb) GetUserNames(users *[]db.Users, tx ...*sqlx.Tx) error {
	sql := fmt.Sprintf("SELECT `id`, `username` FROM `%s`", db.Users{}.TableName())
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Select(users, sql)
	}

	return DB.Select(users, sql)
}

func (d *UsersDb) GetUsersByUserName(users *db.Users, userName string, tx ...*sqlx.Tx) error {
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE `username`=?", users.TableName())
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Get(users, sql, userName)
	}
	return DB.Get(users, sql, userName)
}

func (u *UsersDb) GetUsersByUserId(users *db.Users, id uint, tx ...*sqlx.Tx) error {
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE `id`=?", users.TableName())
	if len(tx) > 0 && tx[0] != nil {
		return tx[0].Get(users, sql, id)
	}
	return DB.Get(users, sql, id)

}

func (u *UsersDb) UpdateUsers(user *db.Users, tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[odshja]")
	}
	if user.Id == 0 {
		return errors.New("用户ID不能为空[odssihja]")
	}

	var sql string
	if CanLock {
		sql = fmt.Sprintf("SELECT `id` FROM `%s` WHERE `id`=? FOR UPDATE", user.TableName())
		if _, err = tx.Exec(sql, user.Id); err != nil {
			return fmt.Errorf("[fuisdku]%s", err)
		}
	}

	sql, args := utils.getUpdateSql(user, user.Id, map[string]interface{}{
		"username":     user.Username,
		"password":     user.Password,
		"passwordShow": user.PasswordShow,
		"quota":        user.Quota,
		"download":     user.Download,
		"upload":       user.Upload,
		"useDays":      user.UseDays,
		"expiryDate":   user.ExpiryDate,
	})
	_, err = tx.Exec(sql, args...)

	return
}

func (u *UsersDb) UpdateUsersClear(tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[odshlja]")
	}

	mu.Lock()
	defer mu.Unlock()

	tn := db.Users{}.TableName()

	var sql string
	if CanLock {
		sql = fmt.Sprintf("LOCK TABLE `%s` WRITE", tn)
		if _, err = tx.Exec(sql); err != nil {
			return fmt.Errorf("[fuiswsdku]%s", err)
		}
	}

	sql = fmt.Sprintf("UPDATE `%s` SET `download` = ?, `upload` = ?", tn)
	_, err = tx.Exec(sql, 0, 0)
	if err != nil {
		return
	}

	if CanLock {
		sql = "UNLOCK TABLES"
		if _, err = tx.Exec(sql); err != nil {
			return fmt.Errorf("[fuis9u]%s", err)
		}
	}

	return
}

func (u *UsersDb) UpdateUsersExpiry(ids []uint, tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[dgkhja]")
	}

	tn := db.Users{}.TableName()

	var sql string
	if CanLock {
		sql = fmt.Sprintf("SELECT `id` FROM `%s` WHERE `id` IN (?) FOR UPDATE", tn)
		query, args, e := sqlx.In(sql, ids)
		if e != nil {
			return e
		}
		if _, err = tx.Exec(tx.Rebind(query), args...); err != nil {
			return
		}
	}

	sql = fmt.Sprintf("UPDATE `%s` SET `quota` = 0 WHERE `id` IN (?)", tn)
	q, a, err := sqlx.In(sql, ids)
	if err != nil {
		return
	}
	_, err = tx.Exec(tx.Rebind(q), a...)
	return
}

func (u *UsersDb) InsertEmptyUsers(userName string, tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[iosdhja]")
	}

	sql, args := utils.getInsertSql(db.Users{}, map[string]interface{}{
		"username":     userName,
		"password":     tool.Hash(userName),
		"passwordShow": tool.Base64Encode(tool.Hash(userName)),
		"quota":        int(50 * QuotaMax),
		"expiryDate":   time.Now().In(global.Tz).AddDate(0, 1, 0).Format(time.DateOnly),
		"useDays":      30,
		"download":     0,
		"upload":       0,
	})
	_, err = tx.Exec(sql, args...)
	return
}

func (u *UsersDb) DelUsersHandle(id uint, tx *sqlx.Tx) (err error) {
	if tx == nil {
		return errors.New("请使用事务[ios8ja]")
	}
	sql := fmt.Sprintf("DELETE FROM `%s` WHERE `id`=?", db.Users{}.TableName())
	_, err = tx.Exec(sql, id)
	return
}
