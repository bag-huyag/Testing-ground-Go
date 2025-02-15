package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Movie struct {
	ID       string    `json:"id"`
	Isbn     string    `json:"isbn"`
	Title    string    `json:"title"`
	Director *Director `json:"director"`
}

type Director struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

var (
	movies    []Movie
	moviesMap map[string]*Movie
	mutex     sync.Mutex
)

func init() {
	moviesMap = make(map[string]*Movie)
}

func getMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mutex.Lock()
	defer mutex.Unlock()
	json.NewEncoder(w).Encode(movies)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	mutex.Lock()
	defer mutex.Unlock()
	for index, item := range movies {
		if item.ID == params["id"] {
			log.Printf("Deleting movie with ID: %s", item.ID)
			movies = append(movies[:index], movies[index+1:]...)
			delete(moviesMap, item.ID)
			break
		}
	}
	json.NewEncoder(w).Encode(movies)
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if movie.Title == "" || movie.Director == nil {
		http.Error(w, "Title and Director are required", http.StatusBadRequest)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	movie.ID = strconv.Itoa(rand.IntN(1000000000))
	movies = append(movies, movie)
	moviesMap[movie.ID] = &movie
	json.NewEncoder(w).Encode(movie)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	mutex.Lock()
	defer mutex.Unlock()
	if movie, ok := moviesMap[params["id"]]; ok {
		var updatedMovie Movie
		if err := json.NewDecoder(r.Body).Decode(&updatedMovie); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updatedMovie.ID = params["id"]
		*movie = updatedMovie
		json.NewEncoder(w).Encode(movie)
		return
	}
	http.Error(w, "Movie not found", http.StatusNotFound)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	mutex.Lock()
	defer mutex.Unlock()
	if movie, ok := moviesMap[params["id"]]; ok {
		json.NewEncoder(w).Encode(movie)
		return
	}
	http.Error(w, "Movie not found", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()

	movies = append(movies, Movie{ID: "1", Isbn: "438222", Title: "Movie One", Director: &Director{Firstname: "John", Lastname: "Doe"}})
	movies = append(movies, Movie{ID: "2", Isbn: "45455", Title: "Movie Two", Director: &Director{Firstname: "Steve", Lastname: "Smith"}})
	moviesMap["1"] = &movies[0]
	moviesMap["2"] = &movies[1]

	r.HandleFunc("/movies", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	r.HandleFunc("/movies", createMovie).Methods("POST")
	r.HandleFunc("/movies/{id}", updateMovie).Methods("PUT")
	r.HandleFunc("/movies/{id}", deleteMovie).Methods("DELETE")

	fmt.Printf("Starting server at port 8000\n")
	log.Fatal(http.ListenAndServe(":8000", r))
}
