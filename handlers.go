// handlers.go

package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
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
	log.Printf("Attempting to find user: %s\n", creds.Username)
	result := db.Where("username = ?", creds.Username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("User not found: %s\n", creds.Username)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Database error: %v\n", result.Error)
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

// GetUsers 函數
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetMemosHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 取得排序參數
	sortBy := r.URL.Query().Get("sort_by") // "time", "importance", "custom"
	var memos []Memo

	switch sortBy {
	case "time":
		// 按提醒时间排序，如果提醒时间相同，按重要程度排序
		db.Where("user_id = ?", userID).
			Order("reminder_time ASC").
			Order("CASE WHEN type = 'important' THEN 1 ELSE 2 END ASC").
			Find(&memos)
	case "importance":
		// 按重要程度排序，如果重要程度相同，按提醒时间排序
		db.Where("user_id = ?", userID).
			Order("CASE WHEN type = 'important' THEN 1 ELSE 2 END ASC").
			Order("reminder_time ASC").
			Find(&memos)
	case "custom":
		// 按 SortOrder 排序
		db.Where("user_id = ?", userID).
			Order("sort_order ASC").
			Find(&memos)
	default:
		// 預設排序，依提醒時間排序
		db.Where("user_id = ?", userID).
			Order("reminder_time ASC").
			Order("CASE WHEN type = 'important' THEN 1 ELSE 2 END ASC").
			Find(&memos)
	}

	// 建構響應數據
	type MemoResponse struct {
		ID           uint    `json:"id"`
		UserID       uint    `json:"user_id"`
		Title        string  `json:"title"`
		Content      string  `json:"content"`
		Type         string  `json:"type"`
		ReminderTime *string `json:"reminder_time"`
		Completed    bool    `json:"completed"`
		SortOrder    int     `json:"sort_order"`
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
			Completed:    memo.Completed,
			SortOrder:    memo.SortOrder,
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
	if strings.TrimSpace(memo.ReminderTimeStr) != "" {
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

	// 設定 SortOrder 為目前使用者的最大 SortOrder + 1
	var maxSortOrder int
	db.Model(&Memo{}).Where("user_id = ?", userID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSortOrder)
	memo.SortOrder = maxSortOrder + 1

	// 儲存到資料庫
	if err := db.Create(&memo).Error; err != nil {
		http.Error(w, "Error creating memo", http.StatusInternalServerError)
		log.Printf("Error creating memo: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func UpdateMemoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid memo ID", http.StatusBadRequest)
		return
	}
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memo Memo
	result := db.First(&memo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Memo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Database error: %v", result.Error)
		return
	}

	if memo.UserID != uint(userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var updatedMemo Memo
	json.NewDecoder(r.Body).Decode(&updatedMemo)

	// 更新字段
	memo.Title = updatedMemo.Title
	memo.Content = updatedMemo.Content
	memo.Type = updatedMemo.Type

	// 解析提醒时间
	if strings.TrimSpace(updatedMemo.ReminderTimeStr) != "" {
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

	// 儲存到資料庫
	if err := db.Save(&memo).Error; err != nil {
		http.Error(w, "Error updating memo", http.StatusInternalServerError)
		log.Printf("Error updating memo: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteMemoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid memo ID", http.StatusBadRequest)
		return
	}
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memo Memo
	result := db.First(&memo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Memo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Database error: %v", result.Error)
		return
	}

	if memo.UserID != uint(userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	db.Delete(&memo)
	w.WriteHeader(http.StatusOK)
}

// 新增函數：切換完成狀態
func CompleteMemoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid memo ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var memo Memo
	result := db.First(&memo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Memo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Database error: %v", result.Error)
		return
	}

	if memo.UserID != uint(userID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Toggle Completed status
	memo.Completed = !memo.Completed

	if err := db.Save(&memo).Error; err != nil {
		http.Error(w, "Failed to update memo", http.StatusInternalServerError)
		log.Printf("Failed to update memo: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"completed": memo.Completed,
	})
}

func UpdateMemosSortHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var sortData []struct {
		ID        uint `json:"id"`
		SortOrder int  `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&sortData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 確保資料一致性
	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	for _, item := range sortData {
		var memo Memo
		if err := tx.Where("id = ? AND user_id = ?", item.ID, userID).First(&memo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				http.Error(w, "Memo not found or unauthorized", http.StatusNotFound)
				return
			}
			tx.Rollback()
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		memo.SortOrder = item.SortOrder
		if err := tx.Save(&memo).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to update sort order", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
