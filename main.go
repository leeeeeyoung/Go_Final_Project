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
	// 資料庫連接字串
	// 替換成你的用戶名與密碼 -> dsn := "username:password@tcp(127.0.0.1:3306)/memo_app?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "username:password@tcp(127.0.0.1:3306)/memo_app?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自動遷移，建立表
	db.AutoMigrate(&User{}, &Memo{})

	router := mux.NewRouter()

	// 靜態檔案伺服器
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// 頁面連結
	router.HandleFunc("/", RedirectToLogin).Methods("GET")
	router.HandleFunc("/login", LoginPage).Methods("GET")
	router.HandleFunc("/register", RegisterPage).Methods("GET")
	router.HandleFunc("/index", AuthMiddleware(IndexPage)).Methods("GET")

	// 用户相關API
	router.HandleFunc("/api/register", RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", LoginHandler).Methods("POST")
	router.HandleFunc("/api/users", GetUsersHandler).Methods("GET")
	router.HandleFunc("/api/logout", LogoutHandler).Methods("POST")

	// 備忘錄相關API
	router.HandleFunc("/api/memos", AuthMiddleware(GetMemosHandler)).Methods("GET")
	router.HandleFunc("/api/memos", AuthMiddleware(CreateMemoHandler)).Methods("POST")
	router.HandleFunc("/api/memos/{id}", AuthMiddleware(UpdateMemoHandler)).Methods("PUT")
	router.HandleFunc("/api/memos/{id}", AuthMiddleware(DeleteMemoHandler)).Methods("DELETE")
	router.HandleFunc("/api/memos/{id}/complete", AuthMiddleware(CompleteMemoHandler)).Methods("POST")

	// 新增 API 端点：更新排序
	router.HandleFunc("/api/memos/sort", AuthMiddleware(UpdateMemosSortHandler)).Methods("POST")

	// 啟動提醒服務
	go ReminderService()

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", router)
}
