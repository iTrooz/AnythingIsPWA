package intmain

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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
