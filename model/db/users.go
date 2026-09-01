package db

type Users struct {
	BaseField
	Username   string `db:"username" json:"username"`
	Password   string `db:"password" json:"password"`
	Quota      int    `db:"quota" json:"quota"`
	Download   uint   `db:"download" json:"download"`
	Upload     uint   `db:"upload" json:"upload"`
	ExpiryDate string `db:"expiryDate" json:"expiryDate"`
}

func (Users) TableName() string {
	return `users`
}
