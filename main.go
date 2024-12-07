// main.go

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// 資料庫連接設定
	// 替換成你的用戶名與密碼 -> dsn := "username:password@tcp(127.0.0.1:3306)/memo_app?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "username:password@tcp(127.0.0.1:3306)/memo_app?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自動建立資料庫表格
	db.AutoMigrate(&User{}, &Memo{})

	router := mux.NewRouter()

	// 靜態檔案伺服器
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// 設定頁面連結
	router.HandleFunc("/", RedirectToLogin).Methods("GET")
	router.HandleFunc("/login", LoginPage).Methods("GET")
	router.HandleFunc("/register", RegisterPage).Methods("GET")
	router.HandleFunc("/index", AuthMiddleware(IndexPage)).Methods("GET")

	// 設定用戶相關 API
	router.HandleFunc("/api/register", RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", LoginHandler).Methods("POST")
	router.HandleFunc("/api/user", GetUserHandler).Methods("GET")
	router.HandleFunc("/api/logout", LogoutHandler).Methods("POST")

	// 設定備忘錄相關 API
	router.HandleFunc("/api/memos", AuthMiddleware(GetMemosHandler)).Methods("GET")
	router.HandleFunc("/api/memos", AuthMiddleware(CreateMemoHandler)).Methods("POST")
	router.HandleFunc("/api/memos/{id}", AuthMiddleware(UpdateMemoHandler)).Methods("PUT")
	router.HandleFunc("/api/memos/{id}", AuthMiddleware(DeleteMemoHandler)).Methods("DELETE")
	router.HandleFunc("/api/memos/{id}/complete", AuthMiddleware(CompleteMemoHandler)).Methods("POST")

	// 設定更新備忘錄排序 API
	router.HandleFunc("/api/memos/sort", AuthMiddleware(UpdateMemosSortHandler)).Methods("POST")

	// 啟動提醒服務（背景執行）
	go ReminderService()

	// 啟動伺服器
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", router)
}
