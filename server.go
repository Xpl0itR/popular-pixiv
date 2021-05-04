package main

import (
	"github.com/Xpl0itR/popular-pixiv/pixiv"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {
	refreshToken := os.Getenv("REFRESH_TOKEN")

	client, err := pixiv.NewClient(refreshToken)
	if err != nil {
		log.Fatal(err.Error())
	}

	http.Handle("/",               FileHandler("html/index.html"))
	http.Handle("/stylesheet.css", FileHandler("html/stylesheet.css"))
	http.Handle("/script.js",      FileHandler("html/script.js"))
	http.Handle("/search",         SearchHandler(client))
	http.Handle("/autocomplete",   AutocompleteHandler(client))
	http.Handle("/redirect",       RedirectHandler())

	if err = http.ListenAndServe(":80", nil); err != nil {
		println(err.Error())
	}
}

func FileHandler(filePath string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, filePath)
	})
}

func SearchHandler(client *pixiv.Client) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		query    := request.URL.Query()
		word     := query.Get("word")
		match    := query.Get("search_target")
		sortOp   := query.Get("sort")
		reSort   := query.Get("resort")
		duration := query.Get("duration")
		startOp  := query.Get("start_date")
		end      := query.Get("end_date")
		filter   := query.Get("filter")
		numStr   := query.Get("num")

		if word == "" {
			http.Redirect(writer, request, "/?"+request.URL.RawQuery, http.StatusSeeOther)
			return
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			num = 30
		}

		searchParameters := pixiv.SearchParameters{
			Word:      word,
			Match:     match,
			Sort:      sortOp,
			Duration:  duration,
			StartDate: startOp,
			EndDate:   end,
			Filter:    filter,
		}

		start := time.Now()

		result, err := client.SearchBatch(num, &searchParameters)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if reSort == "views" {
			sort.Slice(result, func(i, j int) bool {
				return result[i].TotalView > result[j].TotalView
			})
		} else {
			sort.Slice(result, func(i, j int) bool {
				return result[i].TotalBookmarks > result[j].TotalBookmarks
			})
		}

		model := struct {
			Result      *[]pixiv.Illust
			NumResults  int
			TimeElapsed string
		}{
			&result,
			len(result),
			time.Since(start).String(),
		}

		tmpl, err := template.ParseFiles("html/search.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(writer, model); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func AutocompleteHandler(client *pixiv.Client) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		word := request.URL.Query().Get("word")

		if word == "" {
			http.Error(writer, "400 bad request - word parameter is mandatory", http.StatusBadRequest)
			return
		}

		stream, err := client.GetAutocompleteStream(word)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = io.Copy(writer, stream); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func RedirectHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		destination := request.URL.Query().Get("destination")

		if destination == "" {
			http.Error(writer, "400 bad request - destination parameter is mandatory", http.StatusBadRequest)
			return
		}

		request, err := http.NewRequest("GET", destination, nil)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if request.Host != "i.pximg.net" {
			http.Error(writer, "400 bad request - destination host must be i.pximg.net", http.StatusBadRequest)
			return
		}

		request.Header.Set("Referer", "https://i.pximg.net")

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = io.Copy(writer, response.Body); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
