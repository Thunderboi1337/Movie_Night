package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
)

type Movie struct {
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
}

type APIResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

type TemplateData struct {
	Movies      []Movie
	Title       string
	Overview    string
	PosterPath  string
	ReleaseDate string
	VoteAverage float64
}

type App struct {
	APIKey string
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	var url string

	if searchQuery != "" {
		// Use the search movies endpoint
		url = fmt.Sprintf("https://api.themoviedb.org/3/search/movie?query=%s&language=en-US&page=1&include_adult=false", searchQuery)
	} else {
		// Default to fetching favorite movies
		url = "https://api.themoviedb.org/3/account/21472664/favorite/movies?language=en-US&page=1&sort_by=created_at.asc"
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	var data TemplateData

	if len(apiResponse.Results) > 0 {
		data = TemplateData{
			Movies: apiResponse.Results[:4],
		}
	} else {
		fmt.Println("No movies found in the response.")
	}

	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	t.Execute(w, data)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is not set")
	}

	app := &App{APIKey: apiKey}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", app.indexHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
