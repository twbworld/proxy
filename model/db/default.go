package db

import (
	"github.com/twbworld/proxy/utils"
)

// 所有数据库结构体 都需实现的接口
type Dbfunc interface {
	TableName() string
}

// 可能为null的字段, 用指针
type BaseField struct {
	Id         uint  `db:"id" json:"id"`
	AddTime    int64 `db:"add_time" json:"add_time"`
	UpdateTime int64 `db:"update_time" json:"-"`
}

func (b *BaseField) AddTimeFormat() string {
	return utils.TimeFormat(b.AddTime)
}

func (b *BaseField) UpdateTimeFormat() string {
	return utils.TimeFormat(b.UpdateTime)
}
