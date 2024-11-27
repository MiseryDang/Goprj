package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"github.com/MiseryDang/Goprj/go_crud/utils"
	"github.com/gorilla/mux"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

// func setupTestDB() *gorm.DB {
// 	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
// 	db.AutoMigrate(&models.Movie{}, &models.Director{})
// 	return db
// }

func TestCreateMovie(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	// Xóa bản ghi trước khi chèn mới
	db.Exec("DELETE FROM movies WHERE id = ?", 1)

	movie := models.Movie{Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}}
	payload, _ := json.Marshal(movie)

	req, _ := http.NewRequest("POST", "/movies", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNywidXNlcm5hbWUiOiJoYWhhaGFoYSIsImV4cCI6MTczMjY5MDYyM30.9tCidzN9pK8EihHhHWSOsbL0W5zfQpo2HlmsEniBMiI") // Thêm tiêu đề xác thực giả

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateMovie)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("trình xử lý trả về mã trạng thái sai: nhận được %v muốn %v", status, http.StatusCreated)
	}

	var result models.Movie
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Title != movie.Title {
		t.Errorf("mong đợi %v, nhận được %v", movie.Title, result.Title)
	}
}

func TestGetMovies(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	// Dọn dẹp dữ liệu trước khi thêm
	db.Exec("DELETE FROM movies")

	// Thêm dữ liệu mẫu
	movies := []models.Movie{
		{ID: "1", Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}},
		{ID: "2", Isbn: "45455", Title: "Movie Two", Director: &models.Director{FirstName: "Steve", LastName: "Smith"}},
	}
	for _, movie := range movies {
		db.Create(&movie)
	}

	// Gửi yêu cầu GET
	req, _ := http.NewRequest("GET", "/movies", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMovies)
	handler.ServeHTTP(rr, req)

	// Kiểm tra mã trạng thái
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Giải mã và kiểm tra kết quả
	var result []models.Movie
	err := json.NewDecoder(rr.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(result) != len(movies) {
		t.Errorf("expected %v movies, got %v", len(movies), len(result))
	}
}

func TestGetMovie(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	movie := models.Movie{ID: "1", Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}}
	db.Create(&movie)

	req, _ := http.NewRequest("GET", "/movies/1", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/movies/{id}", GetMovie).Methods("GET")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result models.Movie
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Title != movie.Title {
		t.Errorf("expected %v, got %v", movie.Title, result.Title)
	}
}

func TestUpdateMovie(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	// Xóa các bản ghi cũ và tạo một bản ghi mới
	db.Exec("DELETE FROM movies WHERE id = ?", 1)
	db.Create(&models.Movie{
		ID:    "1", // Đảm bảo ID là số nguyên
		Isbn:  "438222",
		Title: "Movie One",
		Director: &models.Director{
			FirstName: "John",
			LastName:  "Doe",
		},
	})

	// Tạo payload để cập nhật
	updatedMovie := models.Movie{Title: "Updated Movie"}
	payload, _ := json.Marshal(updatedMovie)

	// Gửi yêu cầu PUT
	req, _ := http.NewRequest("PUT", "/movies/1", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer your_fake_token_here") // Thêm tiêu đề xác thực giả
	rr := httptest.NewRecorder()

	// Định tuyến và gọi hàm xử lý
	router := mux.NewRouter()
	router.HandleFunc("/movies/{id}", UpdateMovie).Methods("PUT")
	router.ServeHTTP(rr, req)

	// Kiểm tra mã trạng thái trả về
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Kiểm tra giá trị đã được cập nhật trong cơ sở dữ liệu
	var result models.Movie
	db.First(&result, "id = ?", 1)
	if result.Title != updatedMovie.Title {
		t.Errorf("expected %v, got %v", updatedMovie.Title, result.Title)
	}
}

func TestDeleteMovie(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	movie := models.Movie{ID: "1", Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}}
	db.Create(&movie)

	req, _ := http.NewRequest("DELETE", "/movies/1", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/movies/{id}", DeleteMovie).Methods("DELETE")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	var result models.Movie
	err := db.First(&result, "id = ?", movie.ID).Error
	if err == nil {
		t.Errorf("expected movie to be deleted, but it still exists")
	}
}
func TestGetMoviesByCreator(t *testing.T) {
	db := setupTestDB()
	utils.DB = db

	// Xóa tất cả các bản ghi trước khi chèn mới
	db.Exec("DELETE FROM movies")

	movies := []models.Movie{
		{Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}, CreatedBy: 1},
		{Isbn: "45455", Title: "Movie Two", Director: &models.Director{FirstName: "Steve", LastName: "Smith"}, CreatedBy: 1},
		{Isbn: "12345", Title: "Movie Three", Director: &models.Director{FirstName: "Jane", LastName: "Doe"}, CreatedBy: 2},
	}
	for _, movie := range movies {
		db.Create(&movie)
	}

	req, _ := http.NewRequest("GET", "/movies/creator/1", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/movies/creator/{created_by}", GetMoviesByCreator).Methods("GET")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result []models.Movie
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 movies, got %v", len(result))
	}
}
