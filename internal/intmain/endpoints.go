package intmain

import (
	"encoding/json"
	"html/template"
	"net/http"
)

type UserManifestData struct {
	Name      string
	ShortName string
	StartURL  string
}

func manifestHandler(w http.ResponseWriter, r *http.Request) {

	userManifestData := UserManifestData{
		Name:      r.URL.Query().Get("name"),
		ShortName: r.URL.Query().Get("short_name"),
		StartURL:  r.URL.Query().Get("start_url"),
	}

	manifest := Manifest{
		Name:      userManifestData.Name,
		ShortName: userManifestData.ShortName,
		StartURL:  userManifestData.ShortName,
		Icons: []Icon{
			{
				Src:   "/app/icon.png",
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

func appHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("app.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	paramsData := r.URL.Query().Encode()
	err = tmpl.Execute(w, template.URL(paramsData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
