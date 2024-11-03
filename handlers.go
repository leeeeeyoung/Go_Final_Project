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

	if _, exists := users[creds.Username]; exists {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, _ := HashPassword(creds.Password)
	user := User{
		ID:       len(users) + 1,
		Username: creds.Username,
		Password: hashedPassword,
		Email:    creds.Email,
	}
	users[creds.Username] = user

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	json.NewDecoder(r.Body).Decode(&creds)

	user, exists := users[creds.Username]
	if !exists || !CheckPasswordHash(creds.Password, user.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateJWT(user.ID)
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

	type MemoResponse struct {
		ID           int     `json:"id"`
		UserID       int     `json:"user_id"`
		Title        string  `json:"title"`
		Content      string  `json:"content"`
		Type         string  `json:"type"`
		ReminderTime *string `json:"reminder_time"`
	}

	var memosResponse []MemoResponse
	for _, memo := range memos {
		if memo.UserID == userID {
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
	}

	// 对 memosResponse 进行排序
	sort.SliceStable(memosResponse, func(i, j int) bool {
		// 首先比较事件类型，重要的在前
		if memosResponse[i].Type != memosResponse[j].Type {
			return memosResponse[i].Type == "important"
		}

		// 然后在相同事件类型下，按照提醒时间从早到晚排序
		var timeI, timeJ time.Time
		var hasTimeI, hasTimeJ bool
		if memosResponse[i].ReminderTime != nil {
			timeI, _ = time.Parse("2006-01-02T15:04:05", *memosResponse[i].ReminderTime)
			hasTimeI = true
		}
		if memosResponse[j].ReminderTime != nil {
			timeJ, _ = time.Parse("2006-01-02T15:04:05", *memosResponse[j].ReminderTime)
			hasTimeJ = true
		}

		// 都有提醒时间
		if hasTimeI && hasTimeJ {
			return timeI.Before(timeJ)
		}
		// 只有一个有提醒时间，有提醒时间的在前
		if hasTimeI {
			return true
		}
		if hasTimeJ {
			return false
		}
		// 都没有提醒时间，维持原有顺序
		return false
	})

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
	memo.ID = len(memos) + 1
	memo.UserID = userID

	// 解析提醒时间
	location, _ := time.LoadLocation("Local")
	if memo.ReminderTimeStr != "" {
		reminderTime, err := time.ParseInLocation("2006-01-02T15:04", memo.ReminderTimeStr, location)
		if err != nil {
			http.Error(w, "Invalid reminder time format", http.StatusBadRequest)
			return
		}
		memo.ReminderTime = &reminderTime
	} else {
		memo.ReminderTime = nil
	}

	memos[memo.ID] = memo
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

	memo, exists := memos[id]
	if !exists || memo.UserID != userID {
		http.Error(w, "Memo not found", http.StatusNotFound)
		return
	}

	var updatedMemo Memo
	json.NewDecoder(r.Body).Decode(&updatedMemo)
	updatedMemo.ID = id
	updatedMemo.UserID = userID

	// 解析提醒时间
	location, _ := time.LoadLocation("Local")
	if updatedMemo.ReminderTimeStr != "" {
		reminderTime, err := time.ParseInLocation("2006-01-02T15:04", updatedMemo.ReminderTimeStr, location)
		if err != nil {
			http.Error(w, "Invalid reminder time format", http.StatusBadRequest)
			return
		}
		updatedMemo.ReminderTime = &reminderTime
	} else {
		updatedMemo.ReminderTime = nil
	}

	memos[id] = updatedMemo
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

	memo, exists := memos[id]
	if !exists || memo.UserID != userID {
		http.Error(w, "Memo not found", http.StatusNotFound)
		return
	}

	delete(memos, id)
	w.WriteHeader(http.StatusOK)
}
