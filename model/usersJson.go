package model

type UsersJson struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Quota      int    `json:"quota"`
	Enable     bool   `json:"enable"`
	Level      uint   `json:"level"`
	ExpiryDate string `json:"expiryDate"`
}

type UsersInfo struct {
	UsersJson UsersJson
	Users     Users
}
