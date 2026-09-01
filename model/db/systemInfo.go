package db

type SystemInfo struct {
	BaseField
	Key   string `db:"key" json:"key"`
	Value string `db:"value" json:"value"`
}

func (SystemInfo) TableName() string {
	return `system_info`
}
