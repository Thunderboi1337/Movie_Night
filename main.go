package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) *Page {
	filename := title + ".txt"
	body, _ := os.ReadFile(filename)
	return &Page{Title: title, Body: body}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	url := "https://api.themoviedb.org/3/authentication"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiI3YTcwZWY0YTRiODM0MzYyYmRjNzNkNDc2YmJiYzdmMyIsIm5iZiI6MTcyNDkyMzQ1My4yMDc1NjEsInN1YiI6IjY2ZDAzYWY2Yjg4YzIxOTMyYjY2ZmZhYiIsInNjb3BlcyI6WyJhcGlfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.kXlGmpCLk_v1qttPHp5XFWmi4bQ3PpAI3XaS-9792wc")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	bbody, _ := io.ReadAll(res.Body)

	fmt.Println(string(bbody))

	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
	p1.save()
	p2 := loadPage("TestPage")
	fmt.Println(string(p2.Body))

}
