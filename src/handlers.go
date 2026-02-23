package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ---- Small helpers ----

func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func matchesQuery(p Product, qLower string) bool {
	return strings.Contains(strings.ToLower(p.Name), qLower) ||
		strings.Contains(strings.ToLower(p.Category), qLower)
}

func startIndexForQuery(q string) int {
	if len(productIDs) == 0 {
		return 0
	}
	return int(hashToUint32(q) % uint32(len(productIDs)))
}

// ---- Handlers ----

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func productSearchHandler(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	startTime := time.Now()
	qLower := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	start := startIndexForQuery(qLower)

	checked := 0
	totalFound := 0
	results := make([]Product, 0, MaxResults)

	// Bounded iteration: EXACTLY CheckPerSearch inspections per request
	for i := 0; i < CheckPerSearch; i++ {
		id := productIDs[(start+i)%len(productIDs)]

		// Count every inspected item as work (even if not found)
		checked++

		v, ok := products.Load(id)
		if !ok {
			continue
		}

		// Empty query: still do bounded work, but no matches
		if qLower == "" {
			continue
		}

		p := v.(Product)

		if matchesQuery(p, qLower) {
			totalFound++
			if len(results) < MaxResults {
				results = append(results, p)
			}
		}
	}

	resp := map[string]any{
		"products":       results,
		"total_found":    totalFound,
		"checked":        checked,
		"search_time_ms": time.Since(startTime).Milliseconds(),
	}

	writeJSON(w, resp)
}

func listProductsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	all := make([]Product, 0, TotalProducts)
	products.Range(func(_, value any) bool {
		all = append(all, value.(Product))
		return true
	})

	writeJSON(w, all)
}

func productByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	idStr := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/products/"))
	if idStr == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	v, ok := products.Load(id)
	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	writeJSON(w, v.(Product))
}