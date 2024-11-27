package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"github.com/MiseryDang/Goprj/go_crud/utils"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect test database: %v", err)
	}

	// Tạo bảng từ model Registration
	err = db.AutoMigrate(&models.Registration{}, &models.TokenStore{})
	if err != nil {
		log.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestRegister(t *testing.T) {
	db := setupTestDB()
	utils.DB = db // Đảm bảo kết nối đúng cơ sở dữ liệu test

	registration := models.Registration{
		Username: "testuser",
		Password: "password123",
	}
	payload, _ := json.Marshal(registration)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Register)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var result models.Registration
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Username != registration.Username {
		t.Errorf("expected %v, got %v", registration.Username, result.Username)
	}
}

func TestLogin(t *testing.T) {
	db := setupTestDB()
	utils.DB = db // Đảm bảo sử dụng cơ sở dữ liệu test

	// Chuẩn bị dữ liệu ban đầu
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.Registration{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	db.Create(&user)

	// Thực hiện yêu cầu đăng nhập
	loginPayload := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	payload, _ := json.Marshal(loginPayload)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&result)
	if _, ok := result["token"]; !ok {
		t.Errorf("expected token in response")
	}
}

func TestGetUsers(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	users := []models.Registration{
		{Username: "user1", Password: "password1"},
		{Username: "user2", Password: "password2"},
	}
	for _, user := range users {
		db.Create(&user)
	}

	req, _ := http.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetUsers)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result []models.Registration
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != len(users) {
		t.Errorf("expected %v users, got %v", len(users), len(result))
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	// Tạo người dùng để kiểm tra
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	registration := models.Registration{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	db.Create(&registration)

	// Tạo token JWT cho người dùng
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		UserID:   registration.ID,
		Username: registration.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(utils.JwtKey)

	req, _ := http.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result models.Registration
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Username != registration.Username {
		t.Errorf("expected %v, got %v", registration.Username, result.Username)
	}
}
