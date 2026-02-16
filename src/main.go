package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Product struct {
	ProductID string  `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
}

var (
	mu    sync.Mutex
	store = make(map[string]Product)
)

func main() {
	http.HandleFunc("/products", productsHandler)     // GET list, POST create
	http.HandleFunc("/products/", productByIDHandler) // GET by id

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// POST /products
// GET  /products
func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if p.ProductID == "" || p.Name == "" || p.Price <= 0 {
			http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
			return
		}

		mu.Lock()
		if _, exists := store[p.ProductID]; exists {
			mu.Unlock()
			http.Error(w, "Product already exists", http.StatusConflict)
			return
		}
		store[p.ProductID] = p
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(p)

	case http.MethodGet:
		mu.Lock()
		products := make([]Product, 0, len(store))
		for _, v := range store {
			products = append(products, v)
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(products)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /products/{productId}
func productByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" {
		http.Error(w, "Missing productId", http.StatusBadRequest)
		return
	}

	mu.Lock()
	p, ok := store[id]
	mu.Unlock()

	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(p)
}