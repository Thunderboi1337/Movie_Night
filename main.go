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

type App struct {
	APIKey string
}

type APIResponse struct {
	Results []Movie `json:"results"`
}
type TemplateData struct {
	Movies       []Movie
	SearchMovies []Movie
}

type Movie struct {
	Title       string  `json:"title"`
	Id          int     `json:"id"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	Genre       string  `json:"Genre"`
}

var storedMovies TemplateData

func getStoredMovies() {

	// Open the JSON file
	jsonFile, err := os.ReadFile("m.json")
	if err != nil {
		fmt.Println("Error opening JSON file:", err)

	}

	fmt.Println("Successfully opened m.json")

	// Unmarshal JSON data into the storedMovies variable
	err = json.Unmarshal(jsonFile, &storedMovies.Movies)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)

	}

}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {

	var data TemplateData
	var search_data TemplateData

	genres := []string{"Anime", "Animation", "Action", "Drama", "Comedy", "Random", "Weird", "Last Weeks Winner"}

	var url string
	search := false

	searchQuery := r.URL.Query().Get("search")
	trailer := r.URL.Query().Get("trailer")
	movieId := r.URL.Query().Get("movieId")

	if movieId != "" {

		// Logic to handle the movie addition by its ID
		fmt.Println("Movie ID to add:", movieId)

		// Respond to the client, or redirect to a success page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else if searchQuery != "" {
		// Use the search movies endpoint
		formattedQuery := strings.ReplaceAll(searchQuery, " ", "-")
		url = fmt.Sprintf("https://api.themoviedb.org/3/search/movie?query=%s&language=en-US&page=1&include_adult=false", formattedQuery)
		search = true

	} else if trailer != "" {
		url = "https://api.themoviedb.org/3/movie/293660/videos?language=en-US" //deadpool
	} else {
		// Default to fetching favorite movies
		url = "https://api.themoviedb.org/3/account/21472664/favorite/movies?language=en-US&page=1&sort_by=created_at.asc"
	}

	if search {
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

		if search {
			if len(apiResponse.Results) > 0 {
				search_data = TemplateData{
					SearchMovies: apiResponse.Results[0:8],
				}
			} else {
				fmt.Println("No movies found in the response.")
			}

			for i := range genres {
				search_data.SearchMovies[i].Genre = genres[i]
			}

			t, err := template.ParseFiles("index.html")
			if err != nil {
				log.Fatal(err)
			}

			t.Execute(w, search_data)
		}
	} else {

		if len(storedMovies.Movies) > 0 {
			data = TemplateData{
				Movies: storedMovies.Movies[0:8],
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

}

func main() {

	getStoredMovies()

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
