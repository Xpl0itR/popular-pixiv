package pixiv

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

const (
	apiURL          = "https://app-api.pixiv.net"
	searchApiURL    = apiURL + "/v1/search/illust"
	autocompleteURL = apiURL + "/v2/search/autocomplete"
)

var offsetRegex = regexp.MustCompile("offset=(\\d+)")

type ErrorRes struct {
	UserMessage string `json:"user_message"`
	Message     string `json:"message"`
	Reason      string `json:"reason"`
	//UserMessageDetails interface{} `json:"user_message_details"`
}

// struct contains only what is used
type SearchResult struct {
	Error   *ErrorRes `json:"error"`
	Illusts []Illust  `json:"illusts"`
	NextURL string    `json:"next_url"`
}

// struct contains only what is used
type Illust struct {
	ID             uint64     `json:"id"`
	Title          string     `json:"title"`
	ImageURLs      *ImageURLs `json:"image_urls"`
	XRestrict      int        `json:"x_restrict"`
	TotalView      int        `json:"total_view"`
	TotalBookmarks int        `json:"total_bookmarks"`
}

// struct contains only what is used
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

func (client *Client) SearchApi(params *SearchParameters) (*SearchResult, error) {
	if err := client.RefreshIfExpired(); err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("GET", searchApiURL, nil)
	request.URL.RawQuery = params.toURLEncodedParams()
	request.Header.Set("Authorization", "Bearer "+client.accessToken)

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	searchResult := &SearchResult{}
	if err = unmarshalJSONFromResponse(response, searchResult); err != nil {
		return searchResult, err
	}

	if searchResult.Error != nil {
		return searchResult, errors.New(searchResult.Error.Message)
	}

	return searchResult, nil
}

func (client *Client) SearchBatch(numResults int, params *SearchParameters) ([]Illust, error) {
	var illusts []Illust

	for numResults > 0 {
		results, err := client.SearchApi(params)
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

func (client *Client) GetAutocompleteResponse(word string) (*http.Response, error) {
	if err := client.RefreshIfExpired(); err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("GET", autocompleteURL, nil)
	request.URL.RawQuery = "word=" + word
	request.Header.Set("Authorization", "Bearer "+client.accessToken)
	request.Header.Set("Accept-Language", "en-US")

	return client.httpClient.Do(request)
}

func (client *Client) GetAutocompleteSuggestions(word string) ([]AutocompleteSuggestions, error) {
	if word == "" {
		return nil, parameterError("word", word)
	}

	response, err := client.GetAutocompleteResponse(word)
	if err != nil {
		return nil, err
	}

	autocompleteResponse := AutocompleteResult{}
	if err = unmarshalJSONFromResponse(response, &autocompleteResponse); err != nil {
		return nil, err
	}

	return autocompleteResponse.Tags, nil
}
