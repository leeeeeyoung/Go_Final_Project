package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var users = make(map[string]User)
var memos = make(map[int]Memo)
var jwtKey = []byte("your_secret_key")

type User struct {
	ID       int
	Username string
	Password string
	Email    string
}

type Memo struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	Type            string     `json:"type"` // 新增的事件类型字段
	ReminderTime    *time.Time `json:"reminder_time,omitempty"`
	ReminderTimeStr string     `json:"reminder_time_str,omitempty"`
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

func GetMemosByUserID(userID int) []Memo {
	var userMemos []Memo
	for _, memo := range memos {
		if memo.UserID == userID {
			userMemos = append(userMemos, memo)
		}
	}
	return userMemos
}
