package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Movie struct {
	ID         string    `json:"id"`
	Isbn       string    `json:"isbn"`
	Title      string    `json:"title"`
	Director   *Director `json:"director"`
	DirectorID uint      `json:"director_id" gorm:"foreignKey:DirectorID"`
}

type Director struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

var movies []Movie

func connectToSQLite() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func addMovie(db *gorm.DB, movie Movie) error {
	return db.Create(&movie).Error
}

func getAllMovies(db *gorm.DB) ([]Movie, error) {
	var movies []Movie
	err := db.Preload("Director").Find(&movies).Error
	return movies, err
}

func getMovies(w http.ResponseWriter, r *http.Request) {
	db, err := connectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	movies, err := getAllMovies(db)
	if err != nil {
		http.Error(w, "failed to get movies", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(movies)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, item := range movies {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var movie Movie
	_ = json.NewDecoder(r.Body).Decode(&movie)
	movie.ID = strconv.Itoa(rand.Intn(100000000))
	movies = append(movies, movie)
	w.WriteHeader(http.StatusCreated) // Đặt mã trạng thái là 201 (Created)
	json.NewEncoder(w).Encode(movie)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range movies {
		if item.ID == params["id"] {
			movies = append(movies[:index], movies[index+1:]...)

			var movie Movie
			// Giải mã yêu cầu JSON thành đối tượng movie
			if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// Gán lại ID cho phim đã cập nhật
			movie.ID = params["id"]
			// Thêm phim đã cập nhật vào mảng movies
			movies = append(movies, movie)
			// Trả về phim đã cập nhật dưới dạng JSON
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(movie)
			return
		}
	}
	http.Error(w, "Movie not found", http.StatusNotFound)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for index, item := range movies {
		if item.ID == params["id"] {
			movies = append(movies[:index], movies[index+1:]...)
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		}
	}
	http.Error(w, "Movie not found", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	db, err := connectToSQLite()
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	err = db.AutoMigrate(&Movie{}, &Director{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
	// Xóa data cũ
	db.Exec("DELETE FROM movies")
	db.Exec("DELETE FROM directors")
	// Thêm data
	moviesToAdd := []Movie{
		{ID: "1", Isbn: "438222", Title: "Movie One", Director: &Director{FirstName: "John", LastName: "Doe"}},
		{ID: "2", Isbn: "45455", Title: "Movie Two", Director: &Director{FirstName: "Steve", LastName: "Smith"}},
	}
	for _, movie := range moviesToAdd {
		if err := addMovie(db, movie); err != nil {
			log.Fatal("failed to add movie: ", err)
		}
	}
	// Đọc dữ liệu từ cơ sở dữ liệu
	movies, err := getAllMovies(db)
	if err != nil {
		log.Fatal("failed to get movies: ", err)
	}
	// In ra dữ liệu để kiểm tra
	for _, movie := range movies {
		fmt.Printf("ID: %s, Title: %s, Director: %s %s\n", movie.ID, movie.Title, movie.Director.FirstName, movie.Director.LastName)
	}

	db.Find(&movies)
	r.HandleFunc("/movies", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	r.HandleFunc("/movies", createMovie).Methods("POST")
	r.HandleFunc("/movies/{id}", updateMovie).Methods("PUT")
	r.HandleFunc("/movies/{id}", deleteMovie).Methods("DELETE")
	fmt.Printf("Starting server at port 8009\n")
	log.Fatal(http.ListenAndServe(":8009", r))
}
