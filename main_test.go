package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/movies", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	r.HandleFunc("/movies", createMovie).Methods("POST")
	r.HandleFunc("/movies/{id}", updateMovie).Methods("PUT")
	r.HandleFunc("/movies/{id}", deleteMovie).Methods("DELETE")
	return r
}

func setupDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Movie{}, &Director{})
	return db, nil
}

func TestGetMovies(t *testing.T) {
	movies = []Movie{}

	req, err := http.NewRequest("GET", "/movies", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)
	// Kiểm tra mã trạng thái HTTP trả về
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Giải mã JSON phản hồi và so sánh với mảng rỗng
	var got []Movie
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	// So sánh kết quả với mảng rỗng mong đợi
	if !reflect.DeepEqual(got, []Movie{}) {
		t.Errorf(" body: got %v want %v", got, []Movie{})
	}
}

func TestCreateMovie(t *testing.T) {
	db, _ := setupDatabase()
	defer db.Exec("DELETE FROM movies")
	defer db.Exec("DELETE FROM directors")

	movie := Movie{Isbn: "123456", Title: "New Movie", Director: &Director{FirstName: "Jane", LastName: "Doe"}}
	payload, _ := json.Marshal(movie)
	req, _ := http.NewRequest("POST", "/movies", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var createdMovie Movie
	json.Unmarshal(rr.Body.Bytes(), &createdMovie)
	if createdMovie.Title != movie.Title {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), movie.Title)
	}
}

func TestUpdateMovie(t *testing.T) {
	// Tạo cơ sở dữ liệu thử nghiệm (mảng movies)
	movies = []Movie{{ID: "1", Isbn: "123456", Title: "New Movie", Director: &Director{FirstName: "Jane", LastName: "Doe"}}}
	// Phim cần cập nhật
	updatedMovie := Movie{Title: "Updated Movie", Director: &Director{FirstName: "Jane", LastName: "Doe"}}

	payload, err := json.Marshal(updatedMovie)
	if err != nil {
		t.Fatalf("failed to marshal updated movie: %v", err)
	}
	// Tạo yêu cầu HTTP PUT để cập nhật phim
	req, err := http.NewRequest("PUT", "/movies/1", bytes.NewBuffer(payload)) // Sử dụng ID "1"
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Tạo bộ ghi thử nghiệm và gửi yêu cầu
	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)
	// Kiểm tra mã trạng thái HTTP trả về
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Ghi log và kiểm tra nội dung phản hồi
	t.Logf("Response body: %v", rr.Body.String())
	// Giải mã phản hồi JSON để kiểm tra tiêu đề phim đã cập nhật
	var responseMovie Movie
	if err := json.Unmarshal(rr.Body.Bytes(), &responseMovie); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// So sánh tiêu đề phim để xác minh cập nhật thành công
	if responseMovie.Title != updatedMovie.Title {
		t.Errorf("handler returned unexpected title: got %v want %v", responseMovie.Title, updatedMovie.Title)
	}
}

func TestDeleteMovie(t *testing.T) {
	// Thiết lập mảng movies với phim ban đầu
	movies = []Movie{{ID: "1", Title: "Movie 1"}}
	// Tạo yêu cầu DELETE cho phim có ID "1"
	req, err := http.NewRequest("DELETE", "/movies/1", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	// Tạo bộ ghi thử nghiệm và gửi yêu cầu
	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)
	// Kiểm tra mã trạng thái HTTP trả về
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
	// Kiểm tra xem mảng movies có còn phim hay không
	if len(movies) != 0 {
		t.Errorf("handler did not delete the movie: got %v want %v", len(movies), 0)
	}
}
