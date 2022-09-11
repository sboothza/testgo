// imdb_test.go
package main

import (
	"container/list"
	"testing"
)

func getPlotMock(imdbId string) string {
	return "plot"
}

func TestGetPlot(t *testing.T) {
	actual := GetPlot("tt0000001")
	expected := "Performing on what looks like a small wooden stage, wearing a dress with a hoop skirt and white high-heeled pumps, Carmencita does a dance with kicks and twirls, a smile always on her face."
	if actual != expected {
		t.Error("Failed")
	}
}

func TestProcessLineTitle(t *testing.T) {
	sampleLine := "tt0000001	short	Carmencita	Carmencita	0	1894	\\N	1	Documentary,Short"
	var criteria Criteria
	criteria.PrimaryTitle = "Carmencita"
	criteria.TitleType = "notset"
	criteria.EndYear = -1
	criteria.Genre = "notset"
	criteria.OriginalTitle = "notset"
	criteria.PlotFilter = "notset"
	criteria.RuntimeMinutes = -1
	criteria.StartYear = -1
	var movies *list.List = list.New()

	ProcessLine(sampleLine, criteria, 1, getPlotMock, movies)
	if movies.Len() != 1 {
		t.Error("Failed - nothing matched")
	}
	movie := Movie(movies.Front().Value.(Movie))
	if movie.ImdbId != "tt0000001" {
		t.Error("Failed - id wrong " + movie.ImdbId)
	}
}
