package main

import (
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type CacheItem struct {
	Data       []byte
	Headers    http.Header
	StatusCode int
	Expiration time.Time
}

var (
	httpClient *http.Client
	services   map[string]string
	cache      sync.Map
	// cache keys gotten by request path
	cacheKeys map[string]string
)

func startCacheCleanup() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			now := time.Now()
			cache.Range(func(key, value any) bool {
				item := value.(CacheItem)
				if now.After(item.Expiration) {
					cache.Delete(key)
				}
				return true
			})
		}
	}()
}

func getRecipientUrl(r *http.Request) (string, error) {
	pathParts := strings.Split(r.URL.Path, "/")
	serviceName := pathParts[1]
	recipientUrl, err := url.Parse(services[serviceName])
	if err != nil {
		return "", err
	}
	recipientUrl.Path = path.Join(recipientUrl.Path, strings.Join(pathParts[2:], "/"))
	recipientUrl.RawQuery = r.URL.RawQuery

	return recipientUrl.String(), nil
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	recipientUrl, err := getRecipientUrl(r)
	if err != nil {
		http.Error(w, "Cannot process request", http.StatusBadGateway)
		return
	}

	cacheKey := cacheKeys[r.URL.Path]
	if r.Method == "GET" && cacheKey != "" {
		if data, ok := cache.Load(cacheKey); ok {
			item := data.(CacheItem)
			if time.Now().Before(item.Expiration) {
				maps.Copy(w.Header(), item.Headers)
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(item.StatusCode)
				w.Write(item.Data)
				return
			}
			cache.Delete(cacheKey)
		}
	}

	req, err := http.NewRequest(r.Method, recipientUrl, r.Body)
	if err != nil {
		http.Error(w, "Cannot process request", http.StatusBadGateway)
		return
	}

	for header, values := range r.Header {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "Cannot process request", http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Cannot process request", http.StatusBadGateway)
		return
	}

	if r.Method == "GET" && cacheKey != "" && res.StatusCode == http.StatusOK {
		cacheItem := CacheItem{
			Data:       body,
			Headers:    res.Header.Clone(),
			StatusCode: res.StatusCode,
			Expiration: time.Now().Add(2 * time.Minute),
		}
		cache.Store(cacheKey, cacheItem)
	}

	for header, values := range res.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	if r.Method == "GET" && cacheKey != "" {
		w.Header().Set("X-Cache", "MISS")
	}

	w.WriteHeader(res.StatusCode)
	w.Write(body)
}

func main() {
	server := http.NewServeMux()
	httpClient = &http.Client{}
	services = map[string]string{
		"cart":    os.Getenv("CART_SERVICE_URL"),
		"product": os.Getenv("PRODUCT_SERVICE_URL"),
	}
	cacheKeys = map[string]string{
		"/product/products": "products",
	}

	startCacheCleanup()

	server.HandleFunc("/", proxyHandler)

	http.ListenAndServeTLS(":443", "cert/server.pem", "cert/server.key", server)
}
