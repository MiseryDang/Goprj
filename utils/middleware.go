package utils

import (
	"log"
	"net/http"

	"github.com/MiseryDang/Goprj/go_crud/models"

	"github.com/dgrijalva/jwt-go"
)

func TokenValidChecked(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Thiếu Authorization", http.StatusUnauthorized)
			return
		}
		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})
		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			http.Error(w, "Token không hợp lệ", http.StatusUnauthorized)
			return
		}
		log.Printf("Validated token: %s", tokenString)
		db, err := ConnectToSQLite()
		if err != nil {
			http.Error(w, "Không thể kết nối csdl", http.StatusInternalServerError)
			return
		}
		var user models.Registration
		if err := db.Where("username = ?", claims.Username).First(&user).Error; err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		// Kiểm tra token tồn tại
		var storedToken models.TokenStore
		if err := db.Where("token = ?", tokenString).First(&storedToken).Error; err != nil {
			http.Error(w, "Token không tồn tại", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
