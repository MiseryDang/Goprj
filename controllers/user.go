package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"github.com/MiseryDang/Goprj/go_crud/utils"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var registration models.Registration
	_ = json.NewDecoder(r.Body).Decode(&registration)

	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	registration.Password = string(hashedPassword)

	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	if err := db.Create(&registration).Error; err != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registration)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var credentials models.Registration
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "Không thể kết nối cơ sở dữ liệu", http.StatusInternalServerError)
		return
	}
	var registration models.Registration
	if err := db.Where("username = ?", credentials.Username).First(&registration).Error; err != nil {
		http.Error(w, "Không tìm thấy người dùng", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(registration.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Mật khẩu không hợp lệ", http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{UserID: registration.ID,
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(utils.JwtKey)
	if err != nil {
		log.Printf("Lỗi khi tạo token: %v", err)
		http.Error(w, "Không thể tạo token", http.StatusInternalServerError)
		return
	}
	log.Printf("Token đã tạo: %s", tokenString)
	// Lưu token vào cơ sở dữ liệu
	tokenStore := models.TokenStore{Token: tokenString}
	if err := db.Create(&tokenStore).Error; err != nil {
		log.Printf("Lỗi khi lưu token: %v", err)
		http.Error(w, "Không thể lưu token", http.StatusInternalServerError)
		return
	}
	log.Printf("Token đã lưu: %s", tokenString)

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	var users []models.Registration
	if err := db.Find(&users).Error; err != nil {
		http.Error(w, "failed to get users", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Lấy token từ tiêu đề Authorization
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Thiếu Authorization", http.StatusUnauthorized)
		return
	}
	// Loại bỏ tiền tố "Bearer " nếu có
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Token không hợp lệ", http.StatusUnauthorized)
		return
	}
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "Không thể kết nối csdl", http.StatusInternalServerError)
		return
	}
	var user models.Registration
	if err := db.Where("username = ?", claims.Username).First(&user).Error; err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}
