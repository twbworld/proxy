package db

type SystemInfo struct {
	Id         uint   `db:"id" json:"id"`
	Key        string `db:"key" json:"key"`
	Value      string `db:"value" json:"value"`
	UpdateTime string `db:"update_time" json:"update_time"`
}

func (SystemInfo) TableName() string {
	return `system_info`
}
