package model

// User 用户信息模型
type User struct {
	ID       uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username string `gorm:"default:'admin';size:20;column:username" json:"username"`
	Password string `gorm:"default:'adminadmin';column:password" json:"password"`
}

// UserUpdate 用户更新信息
type UserUpdate struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

// UserLogin 用户登录信息
type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Token JWT令牌信息
type Token struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

// TokenData JWT令牌解析数据
type TokenData struct {
	Username string `json:"username"`
}
