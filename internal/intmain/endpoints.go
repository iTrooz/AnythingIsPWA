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
}

func CreateUserManifestData(query url.Values) UserManifestData {
	return UserManifestData{
		Name:      query.Get("name"),
		ShortName: query.Get("short_name"),
		StartURL:  query.Get("start_url"),
	}

}

func manifestHandler(w http.ResponseWriter, r *http.Request) {

	userManifestData := CreateUserManifestData(r.URL.Query())

	manifest := Manifest{
		Name:      userManifestData.Name,
		ShortName: userManifestData.ShortName,
		StartURL:  "/redirect?url=" + userManifestData.StartURL,
		Icons: []Icon{
			{
				Src:   "/app/icon.png",
				Sizes: "512x512",
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

	paramsStr := r.URL.Query().Encode()
	input := AppInput{
		UserManifestData: CreateUserManifestData(r.URL.Query()),
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
