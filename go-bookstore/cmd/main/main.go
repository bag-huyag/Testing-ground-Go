package main

import (
	"log"
	"net/http"

	"github.com/bag-huyag/go-bookstore/pkg/config"
	"github.com/bag-huyag/go-bookstore/pkg/models"
	"github.com/bag-huyag/go-bookstore/pkg/routes"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	config.Connect()
	models.InitDB(config.GetDB())

	r := mux.NewRouter()
	routes.RegisterBookStoreRoutes(r)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("localhost:9010", r))
}
