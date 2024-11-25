package routes

import (
	"net/http"

	"github.com/MiseryDang/Goprj/go_crud/controllers"
	"github.com/MiseryDang/Goprj/go_crud/utils"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/movies", controllers.GetMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", controllers.GetMovie).Methods("GET")
	r.Handle("/movies", utils.TokenValidChecked(http.HandlerFunc(controllers.CreateMovie))).Methods("POST")
	r.Handle("/movies/{id}", utils.TokenValidChecked(http.HandlerFunc(controllers.UpdateMovie))).Methods("PUT")
	r.Handle("/movies/{id}", utils.TokenValidChecked(http.HandlerFunc(controllers.DeleteMovie))).Methods("DELETE")
	r.HandleFunc("/register", controllers.Register).Methods("POST")
	r.HandleFunc("/login", controllers.Login).Methods("POST")
	r.HandleFunc("/users", controllers.GetUsers).Methods("GET")
	r.HandleFunc("/usersToken", controllers.GetUser).Methods("GET")
	r.HandleFunc("/movies/creator/{created_by}", controllers.GetMoviesByCreator).Methods("GET")
}
