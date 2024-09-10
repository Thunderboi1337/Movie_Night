package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type App struct { // API KEY
	APIKey string
}

type TemplateData struct { // Storing Template information for CLIENT
	Movies       []Movie
	WinnerMovie  Movie
	SearchMovies []Movie
	Trailer      []Trailer
	AboutMovie   Movie
}

type MovieAPIResponse struct { // Response from TMDB API
	MovieResults []Movie `json:"results"`
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

type TrailerAPIResponse struct { // Response from TMDB API
	TrailerResults []Trailer `json:"results"`
}

type Trailer struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Offical bool   `json:"official"`
	Key     string `json:"key"`
	Site    string `json:"site"`
}

var tpl *template.Template

var storedMovies TemplateData

func init() {

	tpl = template.Must(template.ParseGlob("*.html"))

}

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

func storeMovies() {

	file, err := json.MarshalIndent(storedMovies.Movies, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	// Write the marshaled data to the JSON file (m.json)
	err = os.WriteFile("m.json", file, 0644)
	if err != nil {
		fmt.Println("Error writing to JSON file:", err)
		return
	}

}

func (app *App) getMovie(w http.ResponseWriter, r *http.Request) {

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		return
	}

	// Retrieve the category value
	category := r.PostFormValue("category")
	movieID := r.PostFormValue("mov_id")
	log.Printf("Category: %s, Movie ID: %s", category, movieID)

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?language=en-US", movieID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var movie Movie

	baseImageURL := "https://image.tmdb.org/t/p/w500"

	err = json.Unmarshal(body, &movie)
	if err != nil {
		fmt.Println("Error:", err)

	}

	movie.Genre = category
	movie.PosterPath = baseImageURL + movie.PosterPath

	updated := false
	for i, m := range storedMovies.Movies {
		if m.Genre == category {
			storedMovies.Movies[i] = movie
			updated = true
			break
		}
	}

	if !updated {
		storedMovies.Movies = append(storedMovies.Movies, movie)
	}

	// Store the updated movies list
	storeMovies()

	http.Redirect(w, r, "/main/", http.StatusFound)

}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {

	var data TemplateData

	if len(storedMovies.Movies) > 0 {
		data = TemplateData{
			Movies: storedMovies.Movies[0:7],
		}
	} else {
		fmt.Println("No movies found in the response.")
	}

	data.WinnerMovie = storedMovies.Movies[7]

	tpl.Execute(w, data)

}

func (app *App) SearchMoviesHandlers(w http.ResponseWriter, r *http.Request) {

	var search_data TemplateData

	var url string

	searchQuery := r.URL.Query().Get("search")

	if searchQuery != "" {
		// Use the search movies endpoint
		formattedQuery := strings.ReplaceAll(searchQuery, " ", "-")
		url = fmt.Sprintf("https://api.themoviedb.org/3/search/movie?query=%s&language=en-US&page=1&include_adult=false", formattedQuery)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+app.APIKey)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
			fmt.Println("failed")
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
			fmt.Println("failed")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			fmt.Println("failed")

		}

		var apiResponse MovieAPIResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			http.Error(w, "Failed to parse response", http.StatusInternalServerError)
			fmt.Println("failed")

		}

		baseImageURL := "https://image.tmdb.org/t/p/w500"
		for i := range apiResponse.MovieResults {
			apiResponse.MovieResults[i].PosterPath = baseImageURL + apiResponse.MovieResults[i].PosterPath
		}

		if len(apiResponse.MovieResults) > 0 {
			search_data = TemplateData{
				SearchMovies: apiResponse.MovieResults,
			}
		} else {
			fmt.Println("No movies found in the response.")
		}

		tpl.Execute(w, search_data)

	}
}

func (app *App) hostHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/main/", http.StatusFound)
}

func (app *App) AboutHandlers(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		log.Println(r.Method)
	}

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the movie ID value
	movieID := r.PostFormValue("movie_ids")
	if movieID == "" {
		log.Println("Movie ID is missing")
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	// Get Trailer information Section
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/videos?language=en-US", movieID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var trailerAPIresponse struct {
		TrailerResults []Trailer `json:"results"`
	}

	err = json.Unmarshal(body, &trailerAPIresponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		log.Println("Failed to parse response:", err)
		return
	}

	// Filter trailers to include only those with "type": "Trailer"
	var filteredTrailers []Trailer
	for _, trailer := range trailerAPIresponse.TrailerResults {
		if trailer.Type == "Trailer" {
			filteredTrailers = append(filteredTrailers, trailer)
		}
	}

	// Prepare template data
	var trailer_data TemplateData
	if len(filteredTrailers) > 0 {
		trailer_data.Trailer = filteredTrailers
	} else {
		fmt.Println("No trailers found with type 'Trailer'.")
	}

	// About Movie DATA Section
	url = fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?append_to_response=SE&language=en-US", movieID)

	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ = http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ = io.ReadAll(res.Body)

	var movieAPIresponse Movie
	err = json.Unmarshal(body, &movieAPIresponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		log.Println("Failed to parse response:", err)
		return
	}

	if len(movieAPIresponse.Title) > 0 {
		trailer_data.AboutMovie = movieAPIresponse

		baseImageURL := "https://image.tmdb.org/t/p/w500"
		trailer_data.AboutMovie.PosterPath = baseImageURL + trailer_data.AboutMovie.PosterPath

	} else {
		fmt.Println("No movies found in the response.")
	}

	// Render the template with the filtered trailer data
	tpl.ExecuteTemplate(w, "movie.html", trailer_data)
}

func (app *App) WinnerHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Working")

	if r.Method != "POST" {
		log.Println(r.Method)
	}
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the movie ID value
	movieID := r.PostFormValue("winner_id")
	if movieID == "" {
		log.Println("Movie ID is missing")
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	for i, movie := range storedMovies.Movies {
		if strconv.Itoa(movie.Id) == movieID {
			storedMovies.Movies[7] = storedMovies.Movies[i]
			break
		}
	}

	storeMovies()

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
	http.HandleFunc("/add-movie/", app.getMovie)
	http.HandleFunc("/about/", app.AboutHandlers)
	http.HandleFunc("/search/", app.SearchMoviesHandlers)
	http.HandleFunc("/winner/", app.WinnerHandler)
	http.HandleFunc("/main/", app.indexHandler)
	http.HandleFunc("/", app.hostHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
