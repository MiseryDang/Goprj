package utils

import (
	"testing"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func cleanTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM movies")
	db.Exec("DELETE FROM directors")
}
func TestAddMovie(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.Movie{}, &models.Director{})
	cleanTestDB(db) // Xóa dữ liệu trước khi chạy test

	movie := models.Movie{
		ID:    "1",
		Isbn:  "438222",
		Title: "Movie One",
		Director: &models.Director{
			FirstName: "John",
			LastName:  "Doe",
		},
	}
	err = AddMovie(db, movie)
	if err != nil {
		t.Errorf("failed to add movie: %v", err)
	}

	var result models.Movie
	db.Preload("Director").First(&result, "id = ?", movie.ID)
	if result.Title != movie.Title {
		t.Errorf("expected %v, got %v", movie.Title, result.Title)
	}
}

func TestGetAllMovies(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.Movie{}, &models.Director{})

	movies := []models.Movie{
		{ID: "1", Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}},
		{ID: "2", Isbn: "45455", Title: "Movie Two", Director: &models.Director{FirstName: "Steve", LastName: "Smith"}},
	}
	for _, movie := range movies {
		db.Create(&movie)
	}
	result, err := GetAllMovies(db)
	if err != nil {
		t.Errorf("failed to get movies: %v", err)
	}
	if len(result) != len(movies) {
		t.Errorf("xpected %v, got %v", len(movies), len(result))
	}
}
