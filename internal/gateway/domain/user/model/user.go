package model

import "time"

type User struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`                 // 主键
	Username  string    `gorm:"column:username;NOT NULL"`                             // 账号名称
	Nickname  string    `gorm:"column:nickname;NOT NULL"`                             // 昵称
	Password  string    `gorm:"column:password;NOT NULL"`                             // 密码
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"` // 更新时间
}

func (m *User) TableName() string {
	return "user"
}
