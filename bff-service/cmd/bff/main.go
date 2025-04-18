package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var (
	httpClient *http.Client
	services   map[string]string
)

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

	for header, values := range res.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.WriteHeader(res.StatusCode)
	if _, err := io.Copy(w, res.Body); err != nil {
		http.Error(w, "Cannot process request", http.StatusBadGateway)
		return
	}
}

func main() {
	server := http.NewServeMux()
	httpClient = &http.Client{}
	services = map[string]string{
		"cart":    os.Getenv("CART_SERVICE_URL"),
		"product": os.Getenv("PRODUCT_SERVICE_URL"),
	}

	server.HandleFunc("/", proxyHandler)

	http.ListenAndServeTLS(":443", "cert/server.pem", "cert/server.key", server)
}
