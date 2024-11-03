package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// 靜態檔案伺服器
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// 頁面路由
	router.HandleFunc("/", RedirectToLogin).Methods("GET")
	router.HandleFunc("/login", LoginPage).Methods("GET")
	router.HandleFunc("/register", RegisterPage).Methods("GET")
	router.HandleFunc("/index", AuthMiddleware(IndexPage)).Methods("GET")

	// 用戶相關API
	router.HandleFunc("/api/register", RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", LoginHandler).Methods("POST")
	router.HandleFunc("/api/logout", LogoutHandler).Methods("POST")

	// 備忘錄相關API
	router.HandleFunc("/api/memos", AuthMiddleware(GetMemosHandler)).Methods("GET")
	router.HandleFunc("/api/memos", AuthMiddleware(CreateMemoHandler)).Methods("POST")
	router.HandleFunc("/api/memos/{id}", AuthMiddleware(UpdateMemoHandler)).Methods("PUT")
	router.HandleFunc("/api/memos/{id}", AuthMiddleware(DeleteMemoHandler)).Methods("DELETE")

	// 啟動提醒服務
	go ReminderService()

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", router)
}
