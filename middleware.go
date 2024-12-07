// middleware.go

package main

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 取得用戶的 JWT Token
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tokenStr := cookie.Value

		// 驗證 JWT Token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// 若 Token 無效或解析錯誤，重新導向到登入頁面
		if err != nil || !token.Valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// 解析 Token 內的用戶 ID
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID := int(userIDFloat)

		// 將用戶 ID 存入請求上下文，並執行下一個處理器
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
