package main

import (
	"fmt"

	"github.com/blevesearch/bleve"
)

func main() {
	// open a new index
	//mapping := bleve.NewIndexMapping()
	//index, err := bleve.New("example.bleve", mapping)
	index, err := bleve.Open("example.bleve")
	bleve.NewMemOnly()
	if err != nil {
		fmt.Println(err)
		return
	}

	data := struct {
		Name string
	}{
		Name: "text",
	}

	// index some data
	index.Index("2", data)
	index.Index("1", struct {
		Name string
	}{
		Name: "another text",
	})

	// search for some text
	query := bleve.NewMatchQuery("text")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(searchResults)
}