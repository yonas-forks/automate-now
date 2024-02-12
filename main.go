package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Skyth3r/automate-now/backloggd"
	"github.com/Skyth3r/automate-now/letterboxd"
	"github.com/Skyth3r/automate-now/nomadlist"
	"github.com/Skyth3r/automate-now/serializd"
	emoji "github.com/jayco/go-emoji-flag"
	"github.com/mmcdole/gofeed"
)

func main() {

	// Movies
	latestMovieItems, err := getGoFeedItems(fmt.Sprintf("%s%s/rss/", letterboxd.Url, letterboxdUsername))
	if err != nil {
		log.Fatalf("unable to parse rss url. Error: %v", err)
	}
	itemCount := maxItems(latestMovieItems)
	movies := latestGoFeedItems(latestMovieItems, itemCount)

	// Books
	latestBookItems, err := getGoFeedItems(fmt.Sprintf("%s%s", OkuUrl, okuCollectionID))
	if err != nil {
		log.Fatalf("unable to parse rss url. Error: %v", err)
	}
	itemCount = maxItems(latestBookItems)
	books := latestGoFeedItems(latestBookItems, itemCount)

	// TV Shows
	showTitlesAndUrls, err := serializd.GetShows(fmt.Sprintf("%s%s/diary", serializd.Url, serializdUsername))
	if err != nil {
		log.Fatalf("unable to get shows from Serializd. Error: %v", err)
	}
	itemCount = maxItems(showTitlesAndUrls)
	shows := serializd.LatestShows(showTitlesAndUrls, itemCount)

	// Video games
	games, err := backloggd.GetGames(fmt.Sprintf("%s/u/%s/playing/", backloggd.Url, backloggdUsername))
	if err != nil {
		log.Fatalf("unable to get games from Backloggd. Error: %v", err)
	}

	// Travel
	countries, err := nomadlist.GetTravel(fmt.Sprintf("%s%s.json", nomadlist.Url, nomadlistUsername))
	if err != nil {
		log.Fatalf("unable to get countries from Nomadlist. Error: %v", err)
	}

	// Formatting Travel
	travelHeader := "## 🌏 Travel\n"
	// 2023 travel
	countriesIn2023SubHeader := "### 2023\n"
	tripsIn2023 := nomadlist.TripsInYear(countries, "2023")
	countriesIn2023 := removeDupes(tripsIn2023)
	countriesIn2023Body := formatCountries(countriesIn2023)
	// 2024 travel
	countriesIn2024SubHeader := "### 2024\n"
	tripsIn2024 := nomadlist.TripsInYear(countries, "2024")
	countriesIn2024 := removeDupes(tripsIn2024)
	countriesIn2024Body := formatCountries(countriesIn2024)

	// Formatting Books
	booksHeader := "## 📚 Books\n"
	booksBody := formatMediaItems(books)

	// Formatting Movies and TV Shows
	moviesAndTvShowsHeader := "## 🎬 Movies and TV Shows\n"
	// Formatting Movies
	moviesSubHeader := "### Recently watched movies\n"
	moviesBody := formatMediaItems(movies)

	// Formatting TV Shows
	showsSubHeader := "### Recently watched TV shows\n"
	showsBody := formatMediaItems(shows)

	// Formatting Video games
	gamesHeader := "## 🎮 Video Games\n"
	gamesBody := formatMediaItems(games)

	// Get today's date
	date := time.Now().Format("2 Jan 2006")
	updated := fmt.Sprintf("\nLast updated: %v", date)

	staticContent, err := os.ReadFile("static.md")
	if err != nil {
		log.Fatalf("unable to read from static.md file. Error: %v", err)
	}

	// Create now.md
	file, err := os.Create("now.md")
	if err != nil {
		log.Fatalf("unable to create now.md file. Error: %v", err)
	}
	defer file.Close()

	data := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n---\n%s", travelHeader, countriesIn2024SubHeader, countriesIn2024Body, countriesIn2023SubHeader, countriesIn2023Body, booksHeader, booksBody, moviesAndTvShowsHeader, moviesSubHeader, moviesBody, showsSubHeader, showsBody, gamesHeader, gamesBody, updated)
	data = fmt.Sprintf("%s\n\n%s", staticContent, data)

	_, err = io.WriteString(file, data)
	if err != nil {
		log.Fatalf("unable to write to now.md file. Error: %v", err)
	}
	err = file.Sync()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

}

func getGoFeedItems(input string) ([]gofeed.Item, error) {
	var feedItems []gofeed.Item

	feedParser := gofeed.NewParser()
	feed, err := feedParser.ParseURL(input)

	if err != nil {
		return nil, err
	}

	for _, item := range feed.Items {
		feedItems = append(feedItems, *item)
	}

	return feedItems, nil
}

func latestGoFeedItems(items []gofeed.Item, count int) []map[string]string {
	var itemSlice []map[string]string

	for i := 0; i < count; i++ {
		item := make(map[string]string)
		if strings.HasPrefix(items[i].Link, "https://letterboxd.com") {
			item["title"] = letterboxd.GetMovieTitle(items[i].Title)
			item["url"] = letterboxd.GetMovieUrl(items[i].Link)
		} else {
			item["title"] = items[i].Title
			item["url"] = items[i].Link
		}
		itemSlice = append(itemSlice, item)
	}
	return itemSlice
}

func removeDupes(trips []map[string]string) []map[string]string {
	var countries []map[string]string

	for _, trip := range trips {
		// check if a trip["name"] is present in the slice countries
		if !containsValue(countries, "name", trip["name"]) {
			countries = append(countries, trip)
		}
	}

	return countries
}

func containsValue(slice []map[string]string, key, value string) bool {
	for _, m := range slice {
		if _, ok := m[key]; ok {
			if val, ok := m[key]; ok && val == value {
				return true
			}
		}
	}
	return false
}

func formatMarkdownLink(title string, url string) string {
	return fmt.Sprintf("* [%v](%v)", title, url)
}

func formatMediaItems(mediaItems []map[string]string) string {
	var mediaText string
	for i := range mediaItems {
		itemText := formatMarkdownLink(mediaItems[i]["title"], mediaItems[i]["url"])
		mediaText += fmt.Sprintf("%v\n", itemText)
	}
	return mediaText
}

func formatCountries(countries []map[string]string) string {
	var formattedText string

	slices.Reverse(countries)

	for i := range countries {
		if countries[i]["code"] == "UK" {
			countries[i]["code"] = "GB"
		}
		countryEmoji := emoji.GetFlag(countries[i]["code"])
		countryText := fmt.Sprintf("%s %s\n", countryEmoji, countries[i]["name"])
		formattedText += countryText
	}

	return formattedText
}

func maxItems[T gofeed.Item | map[string]string](items []T) int {
	limit := 3
	if len(items) < limit {
		limit = len(items)
	}
	return limit
}
