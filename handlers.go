package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func IndexPage(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	json.NewDecoder(r.Body).Decode(&creds)

	var existingUser User
	result := db.Where("username = ?", creds.Username).First(&existingUser)
	if result.Error == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, _ := HashPassword(creds.Password)
	user := User{
		Username: creds.Username,
		Password: hashedPassword,
		Email:    creds.Email,
	}
	db.Create(&user)

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	json.NewDecoder(r.Body).Decode(&creds)

	var user User
	result := db.Where("username = ?", creds.Username).First(&user)
	if result.Error != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if !CheckPasswordHash(creds.Password, user.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateJWT(int(user.ID))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: tokenString,
		Path:  "/",
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusOK)
}

func GetMemosHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memos []Memo
	db.Where("user_id = ?", userID).Find(&memos)

	// 对 memos 进行排序
	sort.SliceStable(memos, func(i, j int) bool {
		// 首先比较事件类型，重要的在前
		if memos[i].Type != memos[j].Type {
			return memos[i].Type == "important"
		}

		// 然后在相同事件类型下，按照提醒时间从早到晚排序
		if memos[i].ReminderTime != nil && memos[j].ReminderTime != nil {
			return memos[i].ReminderTime.Before(*memos[j].ReminderTime)
		}
		// 只有一个有提醒时间，有提醒时间的在前
		if memos[i].ReminderTime != nil {
			return true
		}
		if memos[j].ReminderTime != nil {
			return false
		}
		// 都没有提醒时间，维持原有顺序
		return false
	})

	// 构建响应数据
	type MemoResponse struct {
		ID           uint    `json:"id"`
		UserID       uint    `json:"user_id"`
		Title        string  `json:"title"`
		Content      string  `json:"content"`
		Type         string  `json:"type"`
		ReminderTime *string `json:"reminder_time"`
	}

	var memosResponse []MemoResponse
	for _, memo := range memos {
		var reminderTimeStr *string
		if memo.ReminderTime != nil {
			reminderTime := memo.ReminderTime.Format("2006-01-02T15:04:05")
			reminderTimeStr = &reminderTime
		}
		memosResponse = append(memosResponse, MemoResponse{
			ID:           memo.ID,
			UserID:       memo.UserID,
			Title:        memo.Title,
			Content:      memo.Content,
			Type:         memo.Type,
			ReminderTime: reminderTimeStr,
		})
	}

	json.NewEncoder(w).Encode(memosResponse)
}

func CreateMemoHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var memo Memo
	json.NewDecoder(r.Body).Decode(&memo)
	memo.UserID = uint(userID)

	// 解析提醒时间
	if memo.ReminderTimeStr != "" {
		location, _ := time.LoadLocation("Local")
		reminderTime, err := time.ParseInLocation("2006-01-02T15:04", memo.ReminderTimeStr, location)
		if err != nil {
			http.Error(w, "Invalid reminder time format", http.StatusBadRequest)
			return
		}
		memo.ReminderTime = &reminderTime
	} else {
		memo.ReminderTime = nil
	}

	// 保存到数据库
	if err := db.Create(&memo).Error; err != nil {
		http.Error(w, "Error creating memo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func UpdateMemoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memo Memo
	result := db.First(&memo, id)
	if result.Error != nil || memo.UserID != uint(userID) {
		http.Error(w, "Memo not found", http.StatusNotFound)
		return
	}

	var updatedMemo Memo
	json.NewDecoder(r.Body).Decode(&updatedMemo)

	// 更新字段
	memo.Title = updatedMemo.Title
	memo.Content = updatedMemo.Content
	memo.Type = updatedMemo.Type

	// 解析提醒时间
	if updatedMemo.ReminderTimeStr != "" {
		location, _ := time.LoadLocation("Local")
		reminderTime, err := time.ParseInLocation("2006-01-02T15:04", updatedMemo.ReminderTimeStr, location)
		if err != nil {
			http.Error(w, "Invalid reminder time format", http.StatusBadRequest)
			return
		}
		memo.ReminderTime = &reminderTime
	} else {
		memo.ReminderTime = nil
	}

	// 保存到数据库
	if err := db.Save(&memo).Error; err != nil {
		http.Error(w, "Error updating memo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteMemoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memo Memo
	result := db.First(&memo, id)
	if result.Error != nil || memo.UserID != uint(userID) {
		http.Error(w, "Memo not found", http.StatusNotFound)
		return
	}

	db.Delete(&memo)
	w.WriteHeader(http.StatusOK)
}
