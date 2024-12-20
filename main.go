// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// Book represents the model for a book
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	connStr := "postgres://postgres:postgres@postgres:5432/bookshop?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the books table if it doesn't exist
	createTableSQL := `
    create table if not exists books (
        id SERIAL primary key,
        title varchar(100) not null,
        author varchar(100) not null,
        isbn varchar(14) unique not null
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books", createBook).Methods("POST")
	router.HandleFunc("/books", updateBook).Methods("PUT") // ISBN query parameter route
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books/{id}", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// Handler functions
func getBooks(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var books []Book
	rows, err := db.Query("select id, title, author, isbn from books")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	var book Book

	err := db.QueryRow("select id, title, author, isbn from books where id = $1",
		params["id"]).Scan(&book.ID, &book.Title, &book.Author, &book.ISBN)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(book)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"insert into books (title, author, isbn) values ($1, $2, $3) returning id",
		book.Title, book.Author, book.ISBN,
	).Scan(&book.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result sql.Result
	var err error

	// Check if ISBN query parameter exists
	isbn := r.URL.Query().Get("isbn")
	if isbn != "" {
		// Update by ISBN
		result, err = db.Exec(
			"update books set title = $1, author = $2, isbn = $3 where isbn = $4",
			book.Title, book.Author, book.ISBN, isbn,
		)
	} else {
		// Update by ID
		params := mux.Vars(r)
		id := params["id"]
		if id == "" {
			http.Error(w, "Either ID in path or ISBN in query parameter is required", http.StatusBadRequest)
			return
		}
		result, err = db.Exec(
			"update books set title = $1, author = $2, isbn = $3 where id = $4",
			book.Title, book.Author, book.ISBN, id,
		)
	}

	if err != nil {
		// Check for unique constraint violation on ISBN
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, "Book with this ISBN already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Fetch the updated book to return the correct ID
	var updatedBook Book
	err = db.QueryRow(
		"select id, title, author, isbn from books where isbn = $1",
		book.ISBN,
	).Scan(&updatedBook.ID, &updatedBook.Title, &updatedBook.Author, &updatedBook.ISBN)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedBook)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	result, err := db.Exec("delete from books where id = $1", params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
