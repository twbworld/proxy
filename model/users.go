package model
//可能为null的字段, 用指针
type Users struct {
	Id uint `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
	PasswordShow string `db:"passwordShow" json:"passwordShow"`
	Quota int `db:"quota" json:"quota"`
	Download uint `db:"download" json:"download"`
	Upload uint `db:"upload" json:"upload"`
	UseDays *int `db:"useDays" json:"useDays"`
	ExpiryDate *string `db:"expiryDate" json:"expiryDate"`
}
