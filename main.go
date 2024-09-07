package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var tmpl = template.Must(template.ParseFiles("templates/index.html"))

// Déclaration de la structure Router et de la fonction NewRouter
type Router struct {
	routes          map[string]http.HandlerFunc
	notFoundHandler http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes:          make(map[string]http.HandlerFunc),
		notFoundHandler: http.NotFound,
	}
}

func (r *Router) Handle(path string, handler http.HandlerFunc) {
	r.routes[path] = handler
}

func (r *Router) SetNotFoundHandler(handler http.HandlerFunc) {
	r.notFoundHandler = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handler, ok := r.routes[req.URL.Path]; ok {
		handler(w, req)
		return
	}
	r.notFoundHandler(w, req)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Créer la table "items" si elle n'existe pas
	createTable()

	router := NewRouter()

	// Définir vos routes
	router.Handle("/", indexHandler)
	router.Handle("/items", itemsHandler)
	router.Handle("/create-item", createItemHandler)
	router.Handle("/delete-item", deleteItemHandler)
	router.Handle("/edit-item", editItemHandler)

	// Définir le gestionnaire pour les pages non trouvées
	router.SetNotFoundHandler(notFoundHandler)

	// Création d'un nouveau multiplexeur pour gérer toutes les requêtes
	mux := http.NewServeMux()

	// Gestion des fichiers statiques
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Gestion du service worker
	mux.HandleFunc("/service-worker.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./service-worker.js")
	})

	// Utilisation du routeur personnalisé pour toutes les autres routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") && r.URL.Path != "/service-worker.js" {
			router.ServeHTTP(w, r)
		}
	})

	log.Println("Serving on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func createTable() {
	query := `
    CREATE TABLE IF NOT EXISTS items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT
    );`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Erreur lors de la création de la table:", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	if !isHTMXRequest(r) {
		notFoundHandler(w, r)
		return
	}

	rows, err := db.Query("SELECT id, name FROM items")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type Item struct {
		ID   int
		Name string
	}
	var items []Item
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.ID, &item.Name)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}

	tmpl, _ := template.ParseFiles("templates/items.html")
	tmpl.Execute(w, items)
}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
	if !isHTMXRequest(r) {
		notFoundHandler(w, r)
		return
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		if name != "" {
			_, err := db.Exec("INSERT INTO items (name) VALUES (?)", name)
			if err != nil {
				log.Fatal(err)
			}
		}
		itemsHandler(w, r)
	}
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	if !isHTMXRequest(r) {
		notFoundHandler(w, r)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	itemsHandler(w, r)
}

func editItemHandler(w http.ResponseWriter, r *http.Request) {
	if !isHTMXRequest(r) {
		notFoundHandler(w, r)
		return
	}

	if r.Method == http.MethodPost {
		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		if name != "" {
			_, err := db.Exec("UPDATE items SET name = ? WHERE id = ?", name, id)
			if err != nil {
				log.Fatal(err)
			}
		}
		itemsHandler(w, r)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	tmpl, _ := template.ParseFiles("templates/404.html")
	tmpl.Execute(w, nil)
}
