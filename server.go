package main

import (
	"bytes"
	"flag"
	"github.com/Xpl0itR/popular-pixiv/pixiv"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

	htmlTemplate := template.Must(template.ParseFiles("html/index.html"))

	var indexHtml bytes.Buffer
	if err := htmlTemplate.Execute(&indexHtml, struct{ IsSearchPage bool }{false}); err != nil {
		log.Fatalln(err.Error())
	}
	indexHtmlReader := bytes.NewReader(indexHtml.Bytes())

	server := http.NewServeMux()
	server.HandleFunc("/search", SearchHandler(client, htmlTemplate))
	server.HandleFunc("/autocomplete", AutocompleteHandler(client))
	server.HandleFunc("/", RootHandler(indexHtmlReader))
	server.HandleFunc("/redirect", RedirectHandler)

	log.Printf("Listening on %s\n", *address)
	if err = http.ListenAndServe(*address, server); err != nil {
		log.Fatalln(err.Error())
	}
}

func SearchHandler(client *pixiv.Client, htmlTemplate *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		word := query.Get("word")
		match := query.Get("search_target")
		sortOp := query.Get("sort")
		reSort := query.Get("resort")
		duration := query.Get("duration")
		start := query.Get("start_date")
		end := query.Get("end_date")
		filter := query.Get("filter")
		excludeAi := query.Get("exclude_ai")
		numStr := query.Get("num")
		redirect := query.Get("redirect")
		blurR18 := query.Get("blur_r18")

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
			StartDate: start,
			EndDate:   end,
			Filter:    filter,
			ExcludeAi: excludeAi == "true",
		}

		if err := searchParameters.Validate(); err != nil {
			http.Error(writer, "400 bad request - "+err.Error(), http.StatusBadRequest)
			return
		}

		startTime := time.Now()

		popular, err := client.SearchPopularPreview(&searchParameters)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		illusts, err := client.SearchIllustBatch(num, &searchParameters)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if reSort == "views" {
			sort.Slice(illusts, func(i, j int) bool {
				return illusts[i].TotalView > illusts[j].TotalView
			})
		} else {
			sort.Slice(illusts, func(i, j int) bool {
				return illusts[i].TotalBookmarks > illusts[j].TotalBookmarks
			})
		}

		model := struct {
			IsSearchPage bool
			Result       [][]pixiv.Illust
			NumResults   int
			TimeElapsed  string
			Redirect     bool
			BlurR18      bool
		}{
			true,
			[][]pixiv.Illust{popular.Illusts, illusts},
			len(illusts),
			time.Since(startTime).String(),
			redirect == "true",
			blurR18 == "true",
		}

		var searchHtml bytes.Buffer
		if err := htmlTemplate.Execute(&searchHtml, model); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		_, _ = searchHtml.WriteTo(writer)
	}
}

func AutocompleteHandler(client *pixiv.Client) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		word := request.URL.Query().Get("word")
		if word == "" {
			http.Error(writer, "400 bad request - word parameter is mandatory", http.StatusBadRequest)
			return
		}

		response, err := client.SearchAutocompleteResponse(word)
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

func RootHandler(indexHtmlReader *bytes.Reader) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/":
			_, _ = indexHtmlReader.Seek(0, io.SeekStart)
			_, _ = indexHtmlReader.WriteTo(writer)
		case "/script.js":
			http.ServeFile(writer, request, "html/script.js")
		default:
			http.NotFound(writer, request)
		}
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

	_, _ = io.Copy(writer, response.Body)
}
