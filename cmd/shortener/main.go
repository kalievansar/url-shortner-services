package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	urlStore = make(map[string]string)
	mu       sync.RWMutex
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano())) // Инициализация генератора случайных чисел
}

// Функция для генерации случайного сокращённого URL
func generateShortURL(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func createShortURL(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
	}
	originalURL := string(body)
	shortURL := generateShortURL(8) // Генерируем идентификатор длиной 8 символов

	mu.Lock()
	urlStore[shortURL] = originalURL
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("http://localhost:8080/" + shortURL))
}

func redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]

	mu.RLock()
	originalURL, ok := urlStore[shortURL]
	mu.RUnlock()

	if !ok {
		http.Error(w, "Short URL not found", http.StatusBadRequest)
		return
	}

	//w.Header().Set("Location", originalURL)
	_, err := w.Write([]byte("Location: " + originalURL))
	if err != nil {
		http.Error(w, "Error writing response: %v", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			redirectToOriginalURL(writer, request)
			return
		} else if request.Method == http.MethodPost {
			createShortURL(writer, request)
			return
		} else {
			http.Error(writer, "not current method", http.StatusBadRequest)
			return
		}
	})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
