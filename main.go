package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Manifest struct {
	Name            string `json:"name"`
	ShortName       string `json:"short_name"`
	StartURL        string `json:"start_url"`
	Icons           []Icon `json:"icons"`
	Display         string `json:"display"`
	ThemeColor      string `json:"theme_color"`
	BackgroundColor string `json:"background_color"`
}

type Icon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

func main() {
	http.HandleFunc("/manifest.json", manifestHandler)
	http.HandleFunc("/icon.png", iconHandler)
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/redirect", redirectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on port %s...\n", port)
	http.ListenAndServe(":"+port, nil)
}

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	manifest := Manifest{
		Name:      "My PWA",
		ShortName: "PWA",
		StartURL:  "/",
		Icons: []Icon{
			{
				Src:   "/icon.png",
				Sizes: "192x192",
				Type:  "image/png",
			},
		},
		Display:         "standalone",
		ThemeColor:      "#ffffff",
		BackgroundColor: "#ffffff",
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(manifestBytes)
}

func iconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "icon.jpg")
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
