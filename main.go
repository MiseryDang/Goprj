package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Movie struct {
	ID         string    `json:"id"`
	Isbn       string    `json:"isbn"`
	Title      string    `json:"title"`
	Director   *Director `json:"director"`
	DirectorID uint      `json:"director_id" gorm:"foreignKey:DirectorID"`
	CreatedBy  uint      `json:"created_by"`
}

type Director struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type Registration struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique"`
	FullName string `json:"fullname"`
	Password string `json:"password"`
}

var jwtKey = []byte("secret_key")

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

type TokenStore struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Token string `json:"token" gorm:"unique"`
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
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Token không hợp lệ", http.StatusUnauthorized)
		return
	}
	movie.CreatedBy = claims.UserID

	db, err := connectToSQLite()
	if err != nil {
		http.Error(w, "Không thể kết nối cơ sở dữ liệu", http.StatusInternalServerError)
		return
	}
	if err := addMovie(db, movie); err != nil {
		http.Error(w, "Không thể thêm phim", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
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

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var registration Registration
	_ = json.NewDecoder(r.Body).Decode(&registration)

	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	registration.Password = string(hashedPassword)

	db, err := connectToSQLite()
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

// get hết
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := connectToSQLite()
	if err != nil {
		http.Error(w, "failed to connect database", http.StatusInternalServerError)
		return
	}
	var users []Registration
	if err := db.Find(&users).Error; err != nil {
		http.Error(w, "failed to get users", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

// kiểm tra belong
func getMoviesByCreator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	creatorID, err := strconv.Atoi(params["created_by"])
	if err != nil {
		http.Error(w, "Invalid creator ID", http.StatusBadRequest)
		return
	}

	db, err := connectToSQLite()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	var movies []Movie
	if err := db.Where("created_by = ?", creatorID).Preload("Director").Find(&movies).Error; err != nil {
		http.Error(w, "Failed to get movies", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movies)
}

// get riêng từng token
func getUser(w http.ResponseWriter, r *http.Request) {
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
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Token không hợp lệ", http.StatusUnauthorized)
		return
	}
	db, err := connectToSQLite()
	if err != nil {
		http.Error(w, "Không thể kết nối csdl", http.StatusInternalServerError)
		return
	}
	var user Registration
	if err := db.Where("username = ?", claims.Username).First(&user).Error; err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var credentials Registration
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	db, err := connectToSQLite()
	if err != nil {
		http.Error(w, "Không thể kết nối cơ sở dữ liệu", http.StatusInternalServerError)
		return
	}
	var registration Registration
	if err := db.Where("username = ?", credentials.Username).First(&registration).Error; err != nil {
		http.Error(w, "Không tìm thấy người dùng", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(registration.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Mật khẩu không hợp lệ", http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   registration.ID,
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Lỗi khi tạo token: %v", err)
		http.Error(w, "Không thể tạo token", http.StatusInternalServerError)
		return
	}
	log.Printf("Token đã tạo: %s", tokenString)
	// Lưu token vào cơ sở dữ liệu
	tokenStore := TokenStore{Token: tokenString}
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

func TokenValidChecked(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Thiếu Authorization", http.StatusUnauthorized)
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			http.Error(w, "Token không hợp lệ", http.StatusUnauthorized)
			return
		}
		log.Printf("Validated token: %s", tokenString)
		db, err := connectToSQLite()
		if err != nil {
			http.Error(w, "Không thể kết nối csdl", http.StatusInternalServerError)
			return
		}
		var user Registration
		if err := db.Where("username = ?", claims.Username).First(&user).Error; err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		// Kiểm tra token tồn tại
		var storedToken TokenStore
		if err := db.Where("token = ?", tokenString).First(&storedToken).Error; err != nil {
			http.Error(w, "Token không tồn tại", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func enableCors(next http.Handler) http.Handler {
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

func main() {
	r := mux.NewRouter()
	db, err := connectToSQLite()
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	err = db.AutoMigrate(&Movie{}, &Director{}, &Registration{}, &TokenStore{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	// Xóa data cũ
	db.Exec("DELETE FROM movies")
	db.Exec("DELETE FROM directors")
	db.Exec("DELETE FROM registrations")
	db.Exec("DELETE FROM token_stores")

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
	r.Handle("/movies", TokenValidChecked(http.HandlerFunc(createMovie))).Methods("POST")
	r.Handle("/movies/{id}", TokenValidChecked(http.HandlerFunc(updateMovie))).Methods("PUT")
	r.Handle("/movies/{id}", TokenValidChecked(http.HandlerFunc(deleteMovie))).Methods("DELETE")
	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/usersToken", getUser).Methods("GET")
	r.HandleFunc("/movies/creator/{created_by}", getMoviesByCreator).Methods("GET")

	// Middleware cho CORS
	http.Handle("/", enableCors(r))
	fmt.Printf("Starting server at port 8016\n")
	log.Fatal(http.ListenAndServe(":8016", nil))
}
