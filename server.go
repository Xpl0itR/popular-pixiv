package main

import (
	"errors"
	"flag"
	"github.com/Xpl0itR/popular-pixiv/pixiv"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	address := flag.String("address", ":80", "The address the webserver will listen on, in the form \"ip:port\"")
	refreshToken := flag.String("refresh_token", "", "The refresh token of the pixiv account used to access the API")
	flag.Parse()

	client, err := pixiv.NewClient(*refreshToken)
	if err != nil {
		log.Fatalln(err.Error())
	}

	searchTemplate, err := template.ParseFiles("html/search.html")
	if err != nil {
		log.Fatalln(err.Error())
	}

	server := http.NewServeMux()
	server.HandleFunc("/search", SearchHandler(client, searchTemplate))
	server.HandleFunc("/autocomplete", AutocompleteHandler(client))
	server.HandleFunc("/", RootHandler)
	server.HandleFunc("/redirect", RedirectHandler)

	log.Printf("Listening on %s\n", *address)
	if err = http.ListenAndServe(*address, server); err != nil {
		log.Fatalln(err.Error())
	}
}

func SearchHandler(client *pixiv.Client, pageTemplate *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		word := query.Get("word")
		match := query.Get("search_target")
		sortOp := query.Get("sort")
		reSort := query.Get("resort")
		duration := query.Get("duration")
		startOp := query.Get("start_date")
		end := query.Get("end_date")
		filter := query.Get("filter")
		numStr := query.Get("num")

		if word == "" {
			http.Redirect(writer, request, "/?"+request.URL.RawQuery, http.StatusSeeOther)
			return
		}

		if duration == "all_time" {
			duration = ""
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

		if err := searchParameters.Validate(); err != nil {
			http.Error(writer, "400 bad request - "+err.Error(), http.StatusBadRequest)
			return
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
			Result      []pixiv.Illust
			NumResults  int
			TimeElapsed string
		}{
			result,
			len(result),
			time.Since(start).String(),
		}

		if err := pageTemplate.Execute(writer, model); err != nil {
			if !errors.Is(err, syscall.WSAECONNABORTED) {
				log.Println(err.Error())
			}
		}
	}
}

func AutocompleteHandler(client *pixiv.Client) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		word := request.URL.Query().Get("word")
		if word == "" {
			http.Error(writer, "400 bad request - word parameter is mandatory", http.StatusBadRequest)
			return
		}

		response, err := client.GetAutocompleteResponse(word)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		writer.WriteHeader(response.StatusCode)
		if _, err = io.Copy(writer, response.Body); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func RootHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/":
		http.ServeFile(writer, request, "html/index.html")
	case "/script.js":
		http.ServeFile(writer, request, "html/script.js")
	case "/stylesheet.css":
		http.ServeFile(writer, request, "html/stylesheet.css")
	default:
		http.NotFound(writer, request)
	}
}

// RedirectHandler should only be used as a fallback or for testing, it is more efficient to use a cloudflare worker
func RedirectHandler(writer http.ResponseWriter, request *http.Request) {
	destination := request.URL.Query().Get("destination")

	if destination == "" || !strings.HasPrefix(destination, "https://i.pximg.net") {
		http.Error(writer, "400 - bad request", http.StatusBadRequest)
		return
	}

	newRequest, _ := http.NewRequest("GET", destination, nil)
	newRequest.Header.Set("Host", "i.pximg.net")
	newRequest.Header.Set("Referer", "https://www.pixiv.net/")

	response, err := http.DefaultClient.Do(newRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if _, err = io.Copy(writer, response.Body); err != nil {
		if !errors.Is(err, syscall.WSAECONNABORTED) {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}
