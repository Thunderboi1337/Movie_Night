package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

type Movie struct {
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	Genre       string  `json:"Genre"`
}

type APIResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

type TemplateData struct {
	Movies []Movie
}

type App struct {
	APIKey string
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {

	searchQuery := r.URL.Query().Get("search")
	newMovies := r.URL.Query().Get("addmovies")
	var url string

	if searchQuery != "" {
		// Use the search movies endpoint
		formattedQuery := strings.ReplaceAll(searchQuery, " ", "-")
		url = fmt.Sprintf("https://api.themoviedb.org/3/search/movie?query=%s&language=en-US&page=1&include_adult=false", formattedQuery)

	} else if newMovies != "" {

		url = "https://api.themoviedb.org/3/account/21472664/favorite/movies?language=en-US&page=1&sort_by=created_at.asc"

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

	baseImageURL := "https://image.tmdb.org/t/p/w500"
	for i := range apiResponse.Results {
		apiResponse.Results[i].PosterPath = baseImageURL + apiResponse.Results[i].PosterPath
	}

	var data TemplateData

	genres := []string{"Anime", "Animation", "Action", "Drama", "Comedy", "Random", "Werid", "Last Weeks Winner"}

	if len(apiResponse.Results) > 0 {
		data = TemplateData{
			Movies: apiResponse.Results[0:8],
		}
	} else {
		fmt.Println("No movies found in the response.")
	}

	for i := range genres {
		data.Movies[i].Genre = genres[i]
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
