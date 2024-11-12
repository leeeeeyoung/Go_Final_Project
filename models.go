package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtKey []byte

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"password"`
	Email    string `gorm:"type:varchar(100);not null" json:"email"`
	Memos    []Memo `json:"memos"`
}

type Memo struct {
	gorm.Model
	UserID          uint       `gorm:"not null" json:"user_id"`
	Title           string     `gorm:"type:varchar(200);not null" json:"title"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	Type            string     `gorm:"type:varchar(50);not null" json:"type"`
	ReminderTime    *time.Time `json:"reminder_time,omitempty"`
	ReminderTimeStr string     `gorm:"-" json:"reminder_time_str,omitempty"`
	Completed       bool       `gorm:"default:false" json:"completed"`
	SortOrder       int        `gorm:"default:0;index" json:"sort_order"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ParseJWT 用來解析並驗證 JWT token，並返回 user ID
func ParseJWT(tokenString string) (int, error) {
	// 解析並驗證 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 驗證使用的簽名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	// 從 claims 中提取 user_id
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
