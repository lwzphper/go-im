package user

type RegisterReq struct {
	Username string `binding:"required,min=3,max=20" form:"username" json:"username" xml:"username" label:"账号"`
	Nickname string `binding:"max=20" form:"nickname" json:"nickname" xml:"nickname" label:"昵称"`
	Password string `binding:"required,min=6,max=30,alphanumunicode" form:"password" json:"password" xml:"password" label:"密码"`
}

type RegisterResult struct {
	UserId uint64 `json:"user_id"`
}

type LoginReq struct {
	Username string `binding:"required,min=3,max=20" form:"username" json:"username" xml:"username" label:"账号"`
	Password string `binding:"required,min=6,max=30,alphanumunicode" form:"password" json:"password" xml:"password" label:"密码"`
}

type UserLoginInfo struct {
	Id       uint64 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
}

type LoginResult struct {
	Id            uint64 `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	ServerAddress string `json:"server_addr"` // websocket 地址
	Token         string `json:"token"`
}

type ImServerResult struct {
	ServerAddress string `json:"server_addr"` // websocket 地址
}
