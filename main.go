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

type MovieAPIResponse struct {
	MovieResults []Movie `json:"results"`
}

type TrailerAPIResponse struct {
	TrailerResults []Trailer `json:"results"`
}
type TemplateData struct {
	Movies       []Movie
	SearchMovies []Movie
	Trailer      []Trailer
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

type Trailer struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Offical bool   `json:"official"`
	Key     string `json:"key"`
	Site    string `json:"site"`
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

	fmt.Println("Successfully stored movies to m.json")
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {

	var data TemplateData
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

		t, err := template.ParseFiles("index.html")
		if err != nil {
			log.Fatal(err)
		}

		t.Execute(w, search_data)

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

func (app *App) getMovie(w http.ResponseWriter, r *http.Request) {

	log.Print("HTMX request received")
	log.Print(r.Header.Get("HX-Request"))

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

func (app *App) getTrailer(w http.ResponseWriter, r *http.Request) {

	log.Print("HTMX request received")
	log.Print(r.Header.Get("HX-Request"))

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		return
	}

	// Retrieve the movie ID value
	movieID := r.PostFormValue("movie_id")
	log.Printf(" Movie ID: %s", movieID)

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/videos?language=en-US", movieID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+app.APIKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var trailerAPIresponse TrailerAPIResponse
	err = json.Unmarshal(body, &trailerAPIresponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		log.Println("Failed to parse response:", err)
		return
	}

	// Find the first official trailer of type "Trailer"
	var officialTrailer Trailer
	for _, trailer := range trailerAPIresponse.TrailerResults {
		if trailer.Offical && trailer.Type == "Trailer" {
			officialTrailer = trailer
			break
		}
	}

	if officialTrailer.Key != "" {
		// Construct the YouTube embed URL
		officialTrailer.Key = "https://www.youtube.com/embed/" + officialTrailer.Key
	} else {
		log.Println("No official trailer found in the response.")
	}

	t, err := template.New("t").ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	t.Execute(w, officialTrailer)
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
	http.HandleFunc("/about/", app.getTrailer)
	http.HandleFunc("/", app.indexHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
