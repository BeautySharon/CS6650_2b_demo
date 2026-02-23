// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"hash/fnv"
// 	"log"
// 	"net/http"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"
// )

// type Product struct {
// 	ID          int    `json:"id"`
// 	Name        string `json:"name"`
// 	Category    string `json:"category"`
// 	Description string `json:"description"`
// 	Brand       string `json:"brand"`
// }

// // Requirements
// const (
// 	TotalProducts  = 100_000
// 	CheckPerSearch = 100
// 	MaxResults     = 20
// )

// // Thread-safe store
// var products sync.Map // key: int (id), value: Product

// // For bounded iteration we also keep a stable ID list (0..99999)
// var productIDs []int

// func main() {
// 	// Generate 100k products at startup (one-time)
// 	generateProducts()

// 	http.HandleFunc("/health", healthHandler)
// 	http.HandleFunc("/products/search", productSearchHandler)

// 	// Optional debug endpoints (keep if you want)
// 	http.HandleFunc("/products", listProductsHandler)      // GET list (heavy)
// 	http.HandleFunc("/products/", productByIDHandler)      // GET by id

// 	log.Println("Server running on :8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func healthHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	_, _ = w.Write([]byte("ok"))
// }

// func generateProducts() {
// 	brands := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Omega", "Nova"}
// 	categories := []string{"Electronics", "Books", "Home", "Sports", "Beauty", "Toys", "Grocery", "Clothing"}

// 	productIDs = make([]int, 0, TotalProducts)

// 	for i := 1; i <= TotalProducts; i++ {
// 		brand := brands[(i-1)%len(brands)]
// 		category := categories[(i-1)%len(categories)]

// 		p := Product{
// 			ID:          i,
// 			Name:        fmt.Sprintf("Product %s %d", brand, i),
// 			Category:    category,
// 			Description: fmt.Sprintf("This is a sample description for product %d in %s.", i, category),
// 			Brand:       brand,
// 		}

// 		products.Store(i, p)
// 		productIDs = append(productIDs, i)
// 	}

// 	log.Printf("[INIT] Generated %d products\n", TotalProducts)
// }

// // GET /products/search?q={query}
// //
// // Critical requirement: each search checks exactly 100 products then stops.
// // Also: increment counter for EVERY product checked (not just matches).
// func productSearchHandler(w http.ResponseWriter, r *http.Request) {
// 	startTime := time.Now()

// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	q := strings.TrimSpace(r.URL.Query().Get("q"))
// 	qLower := strings.ToLower(q)

// 	// Deterministic start index based on query hash (stable behavior for testing)
// 	start := int(hashToUint32(q) % uint32(len(productIDs)))

// 	checked := 0
// 	totalFound := 0
// 	results := make([]Product, 0, MaxResults)

// 	// Always check EXACTLY 100 products
// 	for i := 0; i < CheckPerSearch; i++ {
// 		idx := (start + i) % len(productIDs)
// 		id := productIDs[idx]

// 		v, ok := products.Load(id)
// 		if !ok {
// 			// Should not happen, but still count as checked "work"
// 			checked++
// 			continue
// 		}

// 		p := v.(Product)
// 		checked++

// 		// case-insensitive match on name/category
// 		if qLower == "" {
// 			// If empty query: treat as no match (keeps behavior stable)
// 			continue
// 		}

// 		nameLower := strings.ToLower(p.Name)
// 		catLower := strings.ToLower(p.Category)

// 		if strings.Contains(nameLower, qLower) || strings.Contains(catLower, qLower) {
// 			totalFound++
// 			if len(results) < MaxResults {
// 				results = append(results, p)
// 			}
// 		}
// 	}

// 	resp := map[string]any{
// 		"products":       results,
// 		"total_found":    totalFound,
// 		"checked":        checked, // super useful as evidence
// 		"search_time_ms": time.Since(startTime).Milliseconds(),
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	_ = json.NewEncoder(w).Encode(resp)
// }

// func hashToUint32(s string) uint32 {
// 	h := fnv.New32a()
// 	_, _ = h.Write([]byte(s))
// 	return h.Sum32()
// }

// // Optional: GET /products  (warning: returns 100k products, heavy)
// func listProductsHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	all := make([]Product, 0, TotalProducts)
// 	products.Range(func(_, value any) bool {
// 		all = append(all, value.(Product))
// 		return true
// 	})

// 	w.Header().Set("Content-Type", "application/json")
// 	_ = json.NewEncoder(w).Encode(all)
// }

// // Optional: GET /products/{id}
// func productByIDHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
// 	idStr = strings.TrimSpace(idStr)
// 	if idStr == "" {
// 		http.Error(w, "Missing id", http.StatusBadRequest)
// 		return
// 	}

// 	// support both "1" and "001" etc
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil || id <= 0 {
// 		http.Error(w, "Invalid id", http.StatusBadRequest)
// 		return
// 	}

// 	v, ok := products.Load(id)
// 	if !ok {
// 		http.Error(w, "Product not found", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	_ = json.NewEncoder(w).Encode(v.(Product))
// }

package main

import (
	"log"
	"net/http"
)

func main() {
	// Load data once at startup
	generateProducts()

	// Routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/products/search", productSearchHandler)

	// Optional debug endpoints
	http.HandleFunc("/products", listProductsHandler)
	http.HandleFunc("/products/", productByIDHandler)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}