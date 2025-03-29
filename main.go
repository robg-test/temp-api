package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "myapi/docs" // Change to your actual project path

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title My API
// @version 1.0
// @description This is a sample server.
// @host localhost:8080
// @BasePath /

var db *sql.DB

func initDB() {
	var err error
	dsn := "mydatabase.db"
	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    value TEXT NOT NULL
  )`)
	if err != nil {
		log.Fatal(err)
	}
}

type Item struct {
	ID    int    `json:"id" swaggerignore:"true"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// @Summary Get an item by ID
// @Description Get a single item from the database by ID
// @Tags items
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} Item
// @Router /items/{id} [get]
func getItemByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var item Item
	err := db.QueryRow("SELECT id, name, value FROM items WHERE id = ?", id).Scan(&item.ID, &item.Name, &item.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// @Summary Create an item
// @Description Create a new item
// @Tags items
// @Accept json
// @Produce json
// @Success 201 {object} Item
// @Router /items [post]
func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO items (name, value) VALUES (?, ?)", item.Name, item.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// @Summary Update an item
// @Description Update an existing item
// @Tags items
// @Accept json
// @Param id path int true "Item ID"
// @Param item body Item true "Item to update"
// @Success 204
// @Router /items/{id} [put]
func updateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE items SET name = ?, value = ? WHERE id = ?", item.Name, item.Value, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/items/{id}", getItemByID).Methods("GET")
	r.HandleFunc("/items", createItem).Methods("POST")
	r.HandleFunc("/items/{id}", updateItem).Methods("PUT")

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	if os.Getenv("Production") == "true" {
		log.Println("Running in production mode with HTTPS")
		log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", r))
	} else {
		log.Println("Running in development mode with HTTP")
		log.Fatal(http.ListenAndServe(":8080", r))
	}
}
