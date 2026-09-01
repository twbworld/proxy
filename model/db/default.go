package db

import (
	"reflect"
	"sync"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/utils"
)

// 所有表都需实现的抽象接口
type Dbfunc interface {
	TableName() string
}

// 可能为null的字段, 用指针
type BaseField struct {
	Id        uint  `db:"id" json:"id"`
	CreatedAt int64 `db:"created_at" json:"created_at"`
	UpdatedAt int64 `db:"updated_at" json:"-"`
}

func (b *BaseField) CreatedAtFormat() string {
	return utils.TimeFormat(b.CreatedAt, global.Tz)
}

func (b *BaseField) UpdateAtFormat() string {
	return utils.TimeFormat(b.UpdatedAt, global.Tz)
}

type dbTag struct {
	CreatedAtDbTag string
	UpdatedAtDbTag string
}

var GetBaseFieldDbTags = sync.OnceValue(func() dbTag {
	var info dbTag
	t := reflect.TypeFor[BaseField]()

	if field, found := t.FieldByName("CreatedAt"); found {
		info.CreatedAtDbTag = field.Tag.Get("db")
	}
	if field, found := t.FieldByName("UpdatedAt"); found {
		info.UpdatedAtDbTag = field.Tag.Get("db")
	}
	return info
})
