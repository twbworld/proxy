package task

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
	"golang.org/x/sync/errgroup"
)

// Expiry 过期用户处理
func (m *Manager) Expiry(ctx context.Context) error {
	var users []db.Users

	if err := dao.App.UsersDb.GetUsers(&users); err != nil {
		global.Log.Errorf("查询失败[fsuojnv]: %v", err)
		return err
	}

	if len(users) == 0 {
		global.Log.Infoln("没有过期用户[gsfiod]")
		return nil
	}

	now := time.Now().In(time.UTC)
	t, err := time.ParseInLocation(time.DateTime, now.Format(time.DateOnly+" 00:00:01"), time.UTC)
	if err != nil {
		global.Log.Errorf("时间出错[djaksofja]: %v", err)
		return err
	}

	t1, t2 := now.AddDate(0, 0, -7), now.AddDate(0, 0, -5)
	var ids []uint
	eg, egCtx := errgroup.WithContext(ctx)

	for _, user := range users {
		if user.ExpiryDate == "" || user.Id < 1 {
			continue
		}
		ti, err := time.ParseInLocation(time.DateOnly, user.ExpiryDate, time.UTC)
		if err != nil {
			continue
		}

		if t.After(ti) {
			ids = append(ids, user.Id)
		}

		if t1.After(ti) && t2.Before(ti) {
			userCopy := user // 避免迭代器变量捕获问题
			eg.Go(func() error {
				defer func() {
					if r := recover(); r != nil {
						global.Log.Errorf("TgSend协程发生Panic: %v", r)
					}
				}()
				// 响应并阻断任务取消信号
				select {
				case <-egCtx.Done():
					return egCtx.Err()
				default:
					return global.Notifier.Send(userCopy.Username + "快到期" + ti.Format(time.DateOnly))
				}
			})
		}
	}
	_ = eg.Wait()

	if len(ids) == 0 {
		global.Log.Infoln("没有过期用户[ofijsdfio]")
		return nil
	}

	err = dao.Tx(func(tx *sqlx.Tx) error {
		return dao.App.UsersDb.UpdateUsersExpiry(ids, tx)
	})

	if err != nil {
		global.Log.Errorln("更新用户过期状态失败[4r789s]", ids, err)
		return err
	}

	global.Log.Infoln("成功处理过期用户[weou89]: ", ids)
	return nil
}
