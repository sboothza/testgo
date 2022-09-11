package main

import (
	"bufio"
	"container/list"

	"encoding/json"
	"flag"
	"fmt"

	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiUrl = "http://www.omdbapi.com/?i=<IMDBID>&apikey=720b8bf9"

type Movie struct {
	ImdbId string
	Title  string
	Plot   string
}

type MoviePlot struct {
	Title string `json:"Title"`
	Plot  string `json:"Plot"`
}

type Criteria struct {
	TitleType      string
	PrimaryTitle   string
	OriginalTitle  string
	Genre          string
	StartYear      int
	EndYear        int
	RuntimeMinutes int
	PlotFilter     string
	MaxLines       int
}

type GetPlotFunc func(imdbId string) string

var mutex = &sync.Mutex{}

func GetPlot(imdbId string) string {
	url := strings.Replace(apiUrl, "<IMDBID>", imdbId, -1)
	response, err := http.Get(url)

	if err != nil {
		fmt.Print(err.Error())
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var moviePlot MoviePlot
	json.Unmarshal(responseData, &moviePlot)
	return moviePlot.Plot
}

func ProcessLine(line string, criteria Criteria, loopCounter int, getPlot func(string) string, moviesList *list.List) {
	// tconst  titleType       primaryTitle    originalTitle   isAdult startYear       endYear runtimeMinutes  genres

	words := strings.Split(line, "\t")
	var movie Movie
	movie.ImdbId = words[0]
	titleType := words[1]
	movie.Title = words[2]
	originalTitle := words[3]
	startYear, _ := strconv.Atoi(words[5])
	endYear, _ := strconv.Atoi(words[6])
	runtimeMinutes, _ := strconv.Atoi(words[7])
	genres := words[8]

	if (criteria.TitleType == "notset" || titleType == criteria.TitleType) &&
		(criteria.PrimaryTitle == "notset" || strings.Contains(movie.Title, criteria.PrimaryTitle)) &&
		(criteria.OriginalTitle == "notset" || strings.Contains(originalTitle, criteria.OriginalTitle)) &&
		(criteria.Genre == "notset" || strings.Contains(genres, criteria.Genre)) &&
		(criteria.StartYear == -1 || startYear == criteria.StartYear) &&
		(criteria.EndYear == -1 || endYear == criteria.EndYear) &&
		(criteria.RuntimeMinutes == -1 || runtimeMinutes == criteria.RuntimeMinutes) {

		movie.Plot = getPlot(movie.ImdbId)
		if criteria.PlotFilter == "notset" {
			defer mutex.Unlock()
			mutex.Lock()
			moviesList.PushFront(movie)
		} else if match, _ := regexp.MatchString(criteria.PlotFilter, movie.Plot); match {
			defer mutex.Unlock()
			mutex.Lock()
			moviesList.PushFront(movie)
		}
	}
}

func importFile(filename string, criteria Criteria, moviesList *list.List) {
	var file, _ = os.Open(filename)
	scanner := bufio.NewScanner(file)
	var i int = 0
	var wg sync.WaitGroup

scannerLoop:
	for scanner.Scan() {
		line := scanner.Text()

		if i > 0 {
			if criteria.MaxLines > 0 && i >= criteria.MaxLines {
				break scannerLoop
			}

			wg.Add(1)
			go func(value string, loop int) {
				defer wg.Done()
				ProcessLine(value, criteria, loop, GetPlot, moviesList)
			}(line, i)
		}
		i++
	}
	wg.Wait()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncateString(value string, length int) string {
	maxLen := min(len(value), length)
	return value[:maxLen]
}

func main() {
	var criteria Criteria

	filePathPtr := flag.String("filePath", "notset", "absolute path to the inflated `title.basics.tsv.gz` file")
	titleTypePtr := flag.String("titleType", "notset", "filter on `titleType` column")
	primaryTitlePtr := flag.String("primaryTitle", "notset", "filter on `primaryTitle` column")
	originalTitlePtr := flag.String("originalTitle", "notset", "filter on `originalTitle` column")
	genrePtr := flag.String("genre", "notset", "filter on `genre` column")
	startYearPtr := flag.Int("startYear", -1, "filter on `startYear` column")
	endYearPtr := flag.Int("endYear", -1, "filter on `endYear` column")
	runtimeMinutesPtr := flag.Int("runtimeMinutes", -1, "filter on `runtimeMinutes` column")
	//duplicate (of sorts)
	// genresPtr := flag.String("genres", "notset", "filter on `genres` column")
	plotFilterPtr := flag.String("plotFilter", "notset", "regex pattern to apply to the plot of a film retrieved from [omdbapi](https://www.omdbapi.com/)")
	maxApiRequestsPtr := flag.Int("maxApiRequests", 0, "maximum number of requests to be made to [omdbapi](https://www.omdbapi.com/)")
	_ = maxApiRequestsPtr
	maxRunTimePtr := flag.String("maxRunTime", "8760h", "maximum run time of the application. Format is a `time.Duration` string see [here](https://godoc.org/time#ParseDuration)")
	_ = maxRunTimePtr
	maxLinesPtr := flag.Int("maxLines", 0, "maximum lines to process")

	//duplicate
	// maxRequests := flag.Int("maxRequests", 0, "maximum number of requests to send to [omdbapi](https://www.omdbapi.com/)")

	flag.Parse()

	maxRunTime, _ := time.ParseDuration(*maxRunTimePtr)
	_ = maxRunTime
	criteria.TitleType = *titleTypePtr
	criteria.PrimaryTitle = *primaryTitlePtr
	criteria.OriginalTitle = *originalTitlePtr
	criteria.Genre = *genrePtr
	criteria.StartYear = *startYearPtr
	criteria.EndYear = *endYearPtr
	criteria.RuntimeMinutes = *runtimeMinutesPtr
	// criteria.Genres = *genresPtr
	criteria.PlotFilter = *plotFilterPtr
	criteria.MaxLines = *maxLinesPtr

	var movies *list.List = list.New()

	importFile(*filePathPtr, criteria, movies)

	fmt.Printf("%-12s|   %-40s|   %-50s\n", "IMDB_ID", "Title", "Plot")
	for moviePtr := movies.Front(); moviePtr != nil; moviePtr = moviePtr.Next() {
		movie := Movie(moviePtr.Value.(Movie))
		fmt.Printf("%-12s|   %-40s|   %-50s\n", movie.ImdbId, truncateString(movie.Title, 40), truncateString(movie.Plot, 50))
	}
}
