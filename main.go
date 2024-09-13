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

// API KEY
type App struct {
	APIKey string
}

// Storing Template information for CLIENT
type TemplateData struct {
	Movies       []Movie
	WinnerMovie  Movie
	SearchMovies []Movie
	Trailer      []Trailer
	AboutMovie   Movie
}

// Storing response from TMDB API
type MovieAPIResponse struct {
	MovieResults []Movie `json:"results"`
}

// Storing response from TMDB API
type TrailerAPIResponse struct {
	TrailerResults []Trailer `json:"results"`
}

// Movie Template
type Movie struct {
	Title       string  `json:"title"`
	Id          int     `json:"id"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	Genre       string  `json:"Genre"`
}

// Trailer Template
type Trailer struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Offical bool   `json:"official"`
	Key     string `json:"key"`
	Site    string `json:"site"`
}

var tpl *template.Template    // Templete variable
var storedMovies TemplateData // Stored movies from local jsonfile. For main page

// Initzilation for template var and Selects which all html to parse data to.
func init() {

	tpl = template.Must(template.ParseGlob("*.html"))

}

// Get json moviedata from root folder for main pag
func getStoredMovies() {

	// Open the JSON file
	jsonFile, err := os.ReadFile("m.json")
	if err != nil {
		fmt.Println("Error opening JSON file:", err)

	}

	// Unmarshal JSON data into the storedMovies variable
	err = json.Unmarshal(jsonFile, &storedMovies.Movies)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
	}

}

func storeMovies() { // Stores current StoredMovied data into json-file.

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

// Get selected movie Id and stores it in right Category
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

	//Gets movie data from api
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?language=en-US", movieID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	baseImageURL := "https://image.tmdb.org/t/p/w500"

	var movie Movie
	err = json.Unmarshal(body, &movie) //Stores data collected from API into movie struct
	if err != nil {
		fmt.Println("Error:", err)

	}

	movie.Genre = category // Add genre to movie for easy management

	// If find a poster add path to poster.
	if movie.PosterPath != "/static/images/No-Picture-Found.png" {

		movie.PosterPath = baseImageURL + movie.PosterPath
	}

	// if movies find a matching genre it's replaced with current movie data.
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
	// redirects back to main page after selection.
	http.Redirect(w, r, "/main/", http.StatusFound)

}

// Takes the stored movie-data and executes it to the main page
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

// Searches for inputed movie and calls api for inputed movie data
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

		// Calls api for moviedata.
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
			fmt.Println("failed")
		}
		defer res.Body.Close()
		//Checks if status is okay
		if res.StatusCode != http.StatusOK {
			http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
			fmt.Println("failed")
		}
		// Reads Api movie data and stores it in body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			fmt.Println("failed")

		}
		// Stores data in template format
		var apiResponse MovieAPIResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			http.Error(w, "Failed to parse response", http.StatusInternalServerError)
			fmt.Println("failed")

		}
		// Appends url link for posters else loads local alternitive
		baseImageURL := "https://image.tmdb.org/t/p/w500"
		NoImageFound := "/static/images/No-Picture-Found.png"
		for i := range apiResponse.MovieResults {

			if apiResponse.MovieResults[i].PosterPath != "" {

				apiResponse.MovieResults[i].PosterPath = baseImageURL + apiResponse.MovieResults[i].PosterPath
			} else {
				apiResponse.MovieResults[i].PosterPath = NoImageFound
			}
		}
		// Inserts Movie results into Templete
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

// Redirects to main page if attempting to connect to root adress
func (app *App) hostHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/main/", http.StatusFound)
}

// Handles about page data, retrevies trailer data and movie information
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

	// Handles Request to API
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close() // Closes body when function ends

	body, _ := io.ReadAll(res.Body)

	var trailerAPIresponse TrailerAPIResponse
	err = json.Unmarshal(body, &trailerAPIresponse) // Stores data into Trailer Api template
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
	body, _ = io.ReadAll(res.Body) // Stores Request in body

	var movieAPIresponse Movie
	//Stores Body data into Movie template
	err = json.Unmarshal(body, &movieAPIresponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		log.Println("Failed to parse response:", err)
		return
	}

	// Appends url link for posters else loads local alternitive
	NoImageFound := "/static/images/No-Picture-Found.png"
	baseImageURL := "https://image.tmdb.org/t/p/w500"
	if len(movieAPIresponse.Title) > 0 {
		trailer_data.AboutMovie = movieAPIresponse
		if trailer_data.AboutMovie.PosterPath != "" {

			trailer_data.AboutMovie.PosterPath = baseImageURL + trailer_data.AboutMovie.PosterPath
		} else {

			trailer_data.AboutMovie.PosterPath = NoImageFound
		}

	} else {
		fmt.Println("No movies found in the response.")
	}

	// Render the template with the filtered trailer data
	tpl.ExecuteTemplate(w, "movie.html", trailer_data)
}

func (app *App) WinnerHandler(w http.ResponseWriter, r *http.Request) {
	// Checks for Post Method
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

	//Places selcted winner in winner postion.
	for i, movie := range storedMovies.Movies {
		if strconv.Itoa(movie.Id) == movieID {
			storedMovies.Movies[7] = storedMovies.Movies[i]
			break
		}
	}

	storeMovies() // Stores Movies in json

	http.Redirect(w, r, "/main/", http.StatusFound) // Send you back to main page again
}

func main() {

	getStoredMovies() // Gets current movie data from stored in json

	err := godotenv.Load() // Loads ENV files
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("API_KEY") // Get API KEY from ENV format
	if apiKey == "" {
		log.Fatal("API_KEY is not set")
	}

	app := &App{APIKey: apiKey} // Stores API in KEY

	// Finds needed files in static directory
	fs := http.FileServer(http.Dir("static"))
	//Handle functions
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/add-movie/", app.getMovie)
	http.HandleFunc("/about/", app.AboutHandlers)
	http.HandleFunc("/search/", app.SearchMoviesHandlers)
	http.HandleFunc("/winner/", app.WinnerHandler)
	http.HandleFunc("/main/", app.indexHandler)
	http.HandleFunc("/", app.hostHandler)

	// Runs Local Server at port 8080
	log.Fatal(http.ListenAndServe(":8080", nil))

}
