package controllers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"github.com/MiseryDang/Goprj/go_crud/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	db := utils.DB
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
	db := utils.DB
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
		log.Printf("Token không hợp lệ: %v", err)
		http.Error(w, "Token không hợp lệ", http.StatusUnauthorized)
		return
	}
	movie.CreatedBy = claims.UserID

	db := utils.DB
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
	// Kết nối cơ sở dữ liệu
	db := utils.DB
	var movie models.Movie
	// Tìm movie theo ID
	if err := db.First(&movie, "id = ?", params["id"]).Error; err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	// Decode dữ liệu từ request vào struct tạm thời
	var updatedData models.Movie
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Chỉ cập nhật các trường có giá trị mới
	if updatedData.Title != "" {
		movie.Title = updatedData.Title
	}
	if updatedData.Isbn != "" {
		movie.Isbn = updatedData.Isbn
	}
	if updatedData.Director != nil {
		movie.DirectorID = updatedData.Director.ID // Cập nhật ID của Director
	}

	// Lưu lại dữ liệu đã cập nhật
	if err := db.Save(&movie).Error; err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}

	// Trả về đối tượng movie đã được cập nhật
	json.NewEncoder(w).Encode(movie)
}

func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	db := utils.DB
	if err := db.Delete(&models.Movie{}, "id = ?", params["id"]).Error; err != nil {
		http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetMoviesByCreator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	createdBy, err := strconv.Atoi(params["created_by"])
	if err != nil {
		http.Error(w, "Invalid creator ID", http.StatusBadRequest)
		return
	}

	var movies []models.Movie
	db := utils.DB
	if err := db.Where("created_by = ?", createdBy).Find(&movies).Error; err != nil {
		http.Error(w, "Failed to retrieve movies", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movies)
}
