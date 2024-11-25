package controllers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"github.com/MiseryDang/Goprj/go_crud/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	movies, err := utils.GetAllMovies(db)
	if err != nil {
		http.Error(w, "failed to get movies", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(movies)
}

func GetMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	var movie models.Movie
	if err := db.Preload("Director").First(&movie, "id = ?", params["id"]).Error; err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(movie)
}

func CreateMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var movie models.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Dữ liệu đầu vào không hợp lệ", http.StatusBadRequest)
		return
	}
	movie.ID = strconv.Itoa(rand.Intn(100000000))

	// Lấy ID người dùng từ token
	tokenString := r.Header.Get("Authorization")
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
	movie.CreatedBy = claims.UserID

	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "Không thể kết nối cơ sở dữ liệu", http.StatusInternalServerError)
		return
	}
	if err := utils.AddMovie(db, movie); err != nil {
		http.Error(w, "Không thể thêm phim", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)
}

func UpdateMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	var movie models.Movie
	if err := db.First(&movie, "id = ?", params["id"]).Error; err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.Save(&movie).Error; err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(movie)
}

func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	if err := db.Delete(&models.Movie{}, "id = ?", params["id"]).Error; err != nil {
		http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetMoviesByCreator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	creatorID, err := strconv.Atoi(params["created_by"])
	if err != nil {
		http.Error(w, "Invalid creator ID", http.StatusBadRequest)
		return
	}

	db, err := utils.ConnectToSQLite()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	var movies []models.Movie
	if err := db.Where("created_by = ?", creatorID).Preload("Director").Find(&movies).Error; err != nil {
		http.Error(w, "Failed to get movies", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movies)
}
