package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type CacheEntry struct {
	Response *http.Response
	Body     []byte
}

type Cache struct {
	entries map[string]CacheEntry
	mutex   sync.RWMutex
}

// The NewCache function creates and returns a new Cache instance with an empty map of entries.
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
	}
}

// The `Set` method in the `Cache` struct is used to set a cache entry in the cache map.
func (c *Cache) Set(key string, entry CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[key] = entry
}

// The `Get` method in the `Cache` struct is used to retrieve a cache entry based on a given key.
func (c *Cache) Get(key string) (CacheEntry, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	entry, ok := c.entries[key]
	return entry, ok
}

// The `Debug()` method in the `Cache` struct is used to retrieve debug information from the cache. It
// iterates over all entries in the cache, extracts relevant information from each entry (such as URL,
// HTTP method, response status, and response body size), and stores this information in a map with
// string keys and interface{} values. This map is then returned as the debug information.
func (c *Cache) Debug() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	debug := make(map[string]interface{})
	for key, entry := range c.entries {
		debug[key] = map[string]interface{}{
			"URL":    entry.Response.Request.URL.String(),
			"Method": entry.Response.Request.Method,
			"Status": entry.Response.Status,
			"Size":   len(entry.Body),
		}
	}
	return debug
}

var cache = NewCache()

// The `proxyHandler` function serves as a proxy that forwards HTTP requests to a target server, caches
// responses, and forwards the responses back to the client.
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	targetURLParam := r.URL.Query().Get("target")
	if targetURLParam == "" {
		usage := " Usage: ?target=<URL> (e.g., ?target=https://example.com)"
		http.Error(w, "Up and running!"+usage, http.StatusBadRequest)
		return
	}

	targetURL, err := url.Parse(targetURLParam)
	if err != nil {
		http.Error(w, "Invalid 'target' URL", http.StatusBadRequest)
		return
	}

	// Check if the response is cached
	cacheKey := r.Method + " " + targetURL.String() + " " + r.Header.Get("Content-Type") + " " + r.Header.Get("Authorization")
	if cachedEntry, ok := cache.Get(cacheKey); ok {
		log.Printf("Serving cached response for %s\n", targetURL.String())

		// Copy headers from cached response
		for k, v := range cachedEntry.Response.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(cachedEntry.Response.StatusCode)
		w.Write(cachedEntry.Body)
		return
	}

	resp := &http.Response{}
	contentType := r.Header.Get("Content-Type")
	// Forward the request to the target server
	if r.Method == "GET" {
		log.Printf("Forwarding request to %s\n", targetURLParam)

		// forward headers to target
		req, err := http.NewRequest("GET", targetURL.String(), nil)
		if err != nil {
			http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Error forwarding request: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "POST" {
		log.Printf("Forwarding request to %s\n", targetURLParam)

		// forward headers to target
		req, err := http.NewRequest("POST", targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header
		req.Header.Set("Content-Type", contentType)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Error forwarding request: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the response
	cache.Set(cacheKey, CacheEntry{
		Response: resp,
		Body:     body,
	})

	// Forward the response to the client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// The debugHandler function retrieves debug information from a cache and encodes it into JSON format
// to be sent as a response.
func debugHandler(w http.ResponseWriter, r *http.Request) {
	debug := cache.Debug()
	json.NewEncoder(w).Encode(debug)
}

// The main function sets up HTTP handlers for a proxy, health check, and debug endpoints, and starts a
// server listening on port 8080.
func main() {
	http.HandleFunc("/", withCors(proxyHandler))
	http.Handle("/health", withCors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})))
	http.HandleFunc("/debug", withCors(debugHandler))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// withCors is a middleware function that adds CORS headers to the response.
func withCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		next.ServeHTTP(w, r)
	}
}
