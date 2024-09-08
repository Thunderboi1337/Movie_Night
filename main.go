package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type App struct { // API KEY
	APIKey string
}

type TemplateData struct { // Storing Template information for CLIENT
	Movies       []Movie
	SearchMovies []Movie
	Trailer      []Trailer
	StreamInfo   []StreamInformation
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

type StreamAPIResponse struct { // Response from TMDB API
	StreamInfo []StreamInformation `json:"results"`
}

type Country struct {
	CountrySE StreamInformation `json:"SE"`
	CountrySV StreamInformation `json:"SV"`
}

type StreamInformation struct {
	StreamLink string                `json:"link"`
	StreamRent StreamSiteInformation `json:"flatrate"`
	StreamFlat StreamSiteInformation `json:"rent"`
	StreamBuy  StreamSiteInformation `json:"buy"`
	StreamFree StreamSiteInformation `json:"free"`
}
type StreamSiteInformation struct {
	LogoPath     string `json:"logo_path"`
	ProviderId   string `json:"provider_id"`
	ProviderName string `json:"provider_name"`
	DisplayPrio  string `json:"display_priority"`
}

var tpl *template.Template

var storedMovies TemplateData
var trailer_data TemplateData

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
	movieID := r.PostFormValue("movie_id")
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

}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {

	var data TemplateData

	if len(storedMovies.Movies) > 0 {
		data = TemplateData{
			Movies: storedMovies.Movies[0:8],
		}
	} else {
		fmt.Println("No movies found in the response.")
	}

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
		log.Println("Not fuckin buzzzin buzzin")
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

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/videos?language=en-US", movieID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ := http.DefaultClient.Do(req)

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()

	var trailerAPIresponse TrailerAPIResponse
	err = json.Unmarshal(body, &trailerAPIresponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		log.Println("Failed to parse response:", err)
	}

	if len(trailerAPIresponse.TrailerResults) > 0 {
		trailer_data = TemplateData{
			Trailer: trailerAPIresponse.TrailerResults,
		}
	} else {
		fmt.Println("No movies found in the response.")
	}
	//FOR LATER IMPLEMATION____________________________
	url = fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/watch/providers", movieID)

	req, _ = http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ = http.DefaultClient.Do(req)

	body, _ = io.ReadAll(res.Body)
	res.Body.Close()

	fmt.Println(string(body))
	//ALSO FOR LATER IMPLEMENTATION _________________________-
	/* 	url = fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?append_to_response=SE&language=en-US", movieID)

	   	req, _ = http.NewRequest("GET", url, nil)

	   	req.Header.Add("accept", "application/json")
	   	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	   	res, _ = http.DefaultClient.Do(req)

	   	defer res.Body.Close()
	   	body, _ = io.ReadAll(res.Body)

	   	fmt.Println(string(body)) */

	tpl.ExecuteTemplate(w, "movie.html", trailer_data)

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
	http.HandleFunc("/main/", app.indexHandler)
	http.HandleFunc("/", app.hostHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
