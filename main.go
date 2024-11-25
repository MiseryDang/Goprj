package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MiseryDang/Goprj/go_crud/routes"
	"github.com/MiseryDang/Goprj/go_crud/utils"

	"github.com/MiseryDang/Goprj/go_crud/models"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	db, err := utils.ConnectToSQLite()
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	err = db.AutoMigrate(&models.Movie{}, &models.Director{}, &models.Registration{}, &models.TokenStore{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	// Xóa data cũ
	db.Exec("DELETE FROM movies")
	db.Exec("DELETE FROM directors")
	db.Exec("DELETE FROM registrations")
	db.Exec("DELETE FROM token_stores")

	// Thêm data
	moviesToAdd := []models.Movie{
		{ID: "1", Isbn: "438222", Title: "Movie One", Director: &models.Director{FirstName: "John", LastName: "Doe"}},
		{ID: "2", Isbn: "45455", Title: "Movie Two", Director: &models.Director{FirstName: "Steve", LastName: "Smith"}},
	}
	for _, movie := range moviesToAdd {
		if err := utils.AddMovie(db, movie); err != nil {
			log.Fatal("failed to add movie: ", err)
		}
	}

	// Đọc dữ liệu từ cơ sở dữ liệu
	movies, err := utils.GetAllMovies(db)
	if err != nil {
		log.Fatal("failed to get movies: ", err)
	}

	// In ra dữ liệu để kiểm tra
	for _, movie := range movies {
		fmt.Printf("ID: %s, Title: %s, Director: %s %s\n", movie.ID, movie.Title, movie.Director.FirstName, movie.Director.LastName)
	}

	routes.RegisterRoutes(r)

	// Middleware cho CORS
	http.Handle("/", utils.EnableCors(r))
	fmt.Printf("Starting server at port 8016\n")
	log.Fatal(http.ListenAndServe(":8016", nil))
}
