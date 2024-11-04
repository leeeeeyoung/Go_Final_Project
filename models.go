// models.go

package main

import (
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
