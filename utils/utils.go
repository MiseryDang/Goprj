package utils

import (
	"net/http"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var JwtKey = []byte("secret_key")

func ConnectToSQLite() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func EnableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AddMovie(db *gorm.DB, movie models.Movie) error {
	return db.Create(&movie).Error
}

func GetAllMovies(db *gorm.DB) ([]models.Movie, error) {
	var movies []models.Movie
	err := db.Preload("Director").Find(&movies).Error
	return movies, err
}
