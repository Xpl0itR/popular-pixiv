package pixiv

import (
    "encoding/json"
    "errors"
    "io"
    "io/ioutil"
    "net/http"
    "regexp"
    "strconv"
)

const (
    apiURL          = "https://app-api.pixiv.net"
    searchURL       = apiURL + "/v1/search/illust"
    autocompleteURL = apiURL + "/v2/search/autocomplete"
)

type ErrorRes struct {
    UserMessage        string      `json:"user_message"`
    Message            string      `json:"message"`
    Reason             string      `json:"reason"`
    UserMessageDetails interface{} `json:"user_message_details"`
}

type RefreshResponse struct {
    AccessToken  string `json:"access_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
    Scope        string `json:"scope"`
    RefreshToken string `json:"refresh_token"`
    //User interface{}    `json:"user"`
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
    TotalView      int        `json:"total_view"`
    TotalBookmarks int        `json:"total_bookmarks"`
}

// struct contains only what is used
type ImageURLs struct {
    SquareMedium string `json:"square_medium"`
}

type AutocompleteResponse struct {
    Tags []AutocompleteSuggestions `json:"tags"`
}

type AutocompleteSuggestions struct {
    Name           string `json:"name"`
    TranslatedName string `json:"translated_name"`
}

func (client *Client) Search(params *SearchParameters) (*SearchResult, error) {
    searchResult := &SearchResult{}

    paramString, err := params.toString()
    if err != nil { return searchResult, err }

    if err := client.RefreshIfExpired(); err != nil {
        return searchResult, err
    }

    request, err := http.NewRequest("GET", searchURL, nil)
    if err != nil { return searchResult, err }

    request.URL.RawQuery = paramString
    request.Header.Set("Authorization", "Bearer " + client.accessToken)

    response, err := http.DefaultClient.Do(request)
    if err != nil { return searchResult, err }

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil { return searchResult, err }

    if err = json.Unmarshal(responseBody, searchResult); err != nil { return searchResult, err }

    if searchResult.Error != nil {
        return searchResult, errors.New(searchResult.Error.Message)
    }

    return searchResult, nil
}

func (client *Client) SearchBatch(numResults int, params *SearchParameters) (illusts []Illust, err error) {
    for numResults > 0 {
        results, err := client.Search(params)
        if err != nil {
            return illusts, err // Todo: aggregate errors rather than returning
        }

        illusts     = append(illusts, results.Illusts...)
        numResults -= len(results.Illusts)

        if results.NextURL == "" { break }

        params.Offset, err = GetSearchOffsetFromURL(results.NextURL)
        if err != nil { return illusts, err } // Todo: aggregate errors rather than returning
    }

    return
}

func GetSearchOffsetFromURL(URL string) (int, error) {
    regex, err := regexp.Compile("(?i)\\?offset=\\d+")
    if err != nil { return 0, err }

    offsetString := regex.FindString(URL)
    if offsetString == "" {
        return 0, nil
    }

    offset, err := strconv.Atoi(offsetString[8:])
    if err != nil { return 0, err }

    return offset, nil
}

func (client *Client) GetAutocompleteStream(word string) (io.ReadCloser, error) {
    if word == "" {
        return nil, parameterError(wordParam, word)
    }

    request, err := http.NewRequest("GET", autocompleteURL, nil)
    if err != nil { return nil, err }

    request.URL.RawQuery = "word=" + word
    request.Header.Set("Authorization", "Bearer " + client.accessToken)
    request.Header.Set("Accept-Language", "en-US")

    response, err := http.DefaultClient.Do(request)
    if err != nil { return nil, err }

    return response.Body, nil
}

func (client *Client) GetAutocompleteSuggestions(word string) (*[]AutocompleteSuggestions, error) {
    stream, err := client.GetAutocompleteStream(word)
    if err != nil { return nil, err }

    responseBody, err := ioutil.ReadAll(stream)
    if err != nil { return nil, err }

    autocompleteResponse := AutocompleteResponse{}
    if err = json.Unmarshal(responseBody, &autocompleteResponse); err != nil { return nil, err }

    return &autocompleteResponse.Tags, nil
}