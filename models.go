// models.go

package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// JWT 金鑰
var jwtKey []byte

// 用戶資料結構
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"password"`
	Email    string `gorm:"type:varchar(100);not null" json:"email"`
	Memos    []Memo `json:"memos"`
}

// 備忘錄資料結構
type Memo struct {
	gorm.Model
	UserID          uint       `gorm:"not null" json:"user_id"`
	User            User       `gorm:"foreignKey:UserID" json:"user"`
	Title           string     `gorm:"type:varchar(200);not null" json:"title"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	Type            string     `gorm:"type:varchar(50);not null" json:"type"`
	ReminderTime    *time.Time `json:"reminder_time,omitempty"`
	ReminderTimeStr string     `gorm:"-" json:"reminder_time_str,omitempty"`
	Completed       bool       `gorm:"default:false" json:"completed"`
	SortOrder       int        `gorm:"default:0;index" json:"sort_order"`
}

// 用戶登入/註冊的認證資料
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// HashPassword 將密碼進行加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 驗證密碼是否正確
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT 生成用戶的 JWT token
func GenerateJWT(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ParseJWT 解析並驗證 JWT token，返回用戶 ID
func ParseJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
