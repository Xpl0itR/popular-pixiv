package pixiv

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

const (
	apiBaseURL        = "https://app-api.pixiv.net"
	illustURL         = apiBaseURL + "/v1/search/illust"
	popularPreviewURL = apiBaseURL + "/v1/search/popular-preview/illust"
	autocompleteURL   = apiBaseURL + "/v2/search/autocomplete"
)

var offsetRegex = regexp.MustCompile("offset=(\\d+)")

type ErrorRes struct {
	Message string `json:"message"`
	//UserMessage string `json:"user_message"`
	//Reason      string `json:"reason"`
	//UserMessageDetails interface{} `json:"user_message_details"`
}

// SearchIllustResult struct contains only what is used
type SearchIllustResult struct {
	Error   *ErrorRes `json:"error"`
	Illusts []Illust  `json:"illusts"`
	NextURL string    `json:"next_url"`
}

// PopularPreviewResult struct contains only what is used
type PopularPreviewResult struct {
	Error   *ErrorRes `json:"error"`
	Illusts []Illust  `json:"illusts"`
}

// Illust struct contains only what is used
type Illust struct {
	ID             uint64     `json:"id"`
	Title          string     `json:"title"`
	ImageURLs      *ImageURLs `json:"image_urls"`
	XRestrict      int        `json:"x_restrict"`
	TotalView      int        `json:"total_view"`
	TotalBookmarks int        `json:"total_bookmarks"`
}

// ImageURLs struct contains only what is used
type ImageURLs struct {
	SquareMedium string `json:"square_medium"`
}

type AutocompleteResult struct {
	Tags []AutocompleteSuggestions `json:"tags"`
}

type AutocompleteSuggestions struct {
	Name           string `json:"name"`
	TranslatedName string `json:"translated_name"`
}

func (client *Client) SearchIllust(params *SearchParameters) (*SearchIllustResult, error) {
	request, _ := http.NewRequest("GET", illustURL, nil)
	request.URL.RawQuery = params.toURLEncodedParams()

	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}

	searchResult := &SearchIllustResult{}
	if err = unmarshalJSONFromResponse(response, searchResult); err != nil {
		return searchResult, err
	}

	if searchResult.Error != nil {
		return searchResult, errors.New(searchResult.Error.Message)
	}

	return searchResult, nil
}

func (client *Client) SearchIllustBatch(numResults int, params *SearchParameters) ([]Illust, error) {
	var illusts []Illust

	for numResults > 0 {
		results, err := client.SearchIllust(params)
		if err != nil {
			return illusts, err
		}

		illusts = append(illusts, results.Illusts...)
		numResults -= len(results.Illusts)

		if results.NextURL == "" {
			break
		}
		params.Offset = GetSearchOffsetFromURL(results.NextURL)
	}

	return illusts, nil
}

func GetSearchOffsetFromURL(URL string) int {
	offsetString := offsetRegex.FindStringSubmatch(URL)[1]
	offset, _ := strconv.Atoi(offsetString)
	return offset
}

func (client *Client) SearchPopularPreview(params *SearchParameters) (*PopularPreviewResult, error) {
	request, _ := http.NewRequest("GET", popularPreviewURL, nil)
	request.URL.RawQuery = params.toURLEncodedParams()

	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}

	popularPreviewResult := &PopularPreviewResult{}
	if err = unmarshalJSONFromResponse(response, popularPreviewResult); err != nil {
		return popularPreviewResult, err
	}

	if popularPreviewResult.Error != nil {
		return popularPreviewResult, errors.New(popularPreviewResult.Error.Message)
	}

	return popularPreviewResult, nil
}

func (client *Client) SearchAutocompleteResponse(word string) (*http.Response, error) {
	request, _ := http.NewRequest("GET", autocompleteURL, nil)
	request.URL.RawQuery = "word=" + word
	request.Header.Set("Accept-Language", "en-US")

	return client.Send(request)
}

func (client *Client) SearchAutocomplete(word string) ([]AutocompleteSuggestions, error) {
	if word == "" {
		return nil, parameterError("word", word)
	}

	response, err := client.SearchAutocompleteResponse(word)
	if err != nil {
		return nil, err
	}

	autocompleteResponse := AutocompleteResult{}
	if err = unmarshalJSONFromResponse(response, &autocompleteResponse); err != nil {
		return nil, err
	}

	return autocompleteResponse.Tags, nil
}
