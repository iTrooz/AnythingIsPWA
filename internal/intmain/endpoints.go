package intmain

import (
	"anythingispwa/internal/websiteinfos"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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

type UserManifestData struct {
	Name      string
	ShortName string
	StartURL  string
	IconURL   string
}

func CreateUserManifestData(query url.Values) (*UserManifestData, error) {
	startURL, err := url.Parse(query.Get("start_url"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse start_url: %w", err)
	}

	if startURL.Scheme == "" {
		// assume https
		startURL.Scheme = "https"
	} else if startURL.Scheme != "http" && startURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid start_url scheme: %v", startURL.Scheme)
	}

	return &UserManifestData{
		Name:      query.Get("name"),
		ShortName: query.Get("short_name"),
		StartURL:  startURL.String(),
		IconURL:   query.Get("icon_url"),
	}, nil

}

func manifestHandler(w http.ResponseWriter, r *http.Request) {

	userManifestData, err := CreateUserManifestData(r.URL.Query())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user manifest data: %v", err), http.StatusBadRequest)
		return
	}

	shortName := userManifestData.ShortName
	if shortName == "" {
		shortName = userManifestData.Name
	}

	iconURL := userManifestData.IconURL
	if iconURL == "" {
		iconURL = "/app/icon.png"
	}

	manifest := Manifest{
		Name:      userManifestData.Name,
		ShortName: shortName,
		StartURL:  "/redirect?url=" + userManifestData.StartURL,
		Icons: []Icon{
			{
				Src:   iconURL,
				Sizes: "512x512", // may not be accurate
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
	_, err = w.Write(manifestBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func iconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/icon.png")
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	for _, key := range []string{"name", "start_url"} {
		if !r.URL.Query().Has(key) {
			http.Error(w, fmt.Sprintf("Required parameter %v is missing", key), http.StatusBadRequest)
			return
		}
	}

	tmpl, err := template.ParseFiles("templates/app.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type AppInput struct {
		UserManifestData
		ParamsStr template.URL
	}

	userManifestData, err := CreateUserManifestData(r.URL.Query())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user manifest data: %v", err), http.StatusBadRequest)
		return
	}

	paramsStr := r.URL.Query().Encode()
	input := AppInput{
		UserManifestData: *userManifestData,
		ParamsStr:        template.URL(paramsStr), // to avoid escaping
	}
	err = tmpl.Execute(w, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

// rate limit: maxTokens requests per refillInterval
func withRateLimit(handler http.HandlerFunc, maxTokens int, refillInterval time.Duration) http.HandlerFunc {
	tokens := maxTokens
	var mu sync.Mutex
	ticker := time.NewTicker(refillInterval)
	go func() {
		for range ticker.C {
			mu.Lock()
			tokens = maxTokens
			mu.Unlock()
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if tokens > 0 {
			tokens--
			mu.Unlock()
			handler(w, r)
		} else {
			mu.Unlock()
			http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
		}
	}
}

// get a website title and icon
func getWebsiteInfoHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	infos, err := websiteinfos.Get(url)
	if err != nil {
		// for security, do not expose the error message in this case
		logrus.Errorf("Failed to get website infos: %v", err)
		http.Error(w, "Failed to get website infos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(infos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
