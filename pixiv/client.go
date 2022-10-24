package pixiv

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	clientID     = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	clientSecret = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
	oAuthURL     = "https://oauth.secure.pixiv.net/auth/token"
)

type Client struct {
	accessToken  string
	refreshToken string
	refreshMutex sync.Mutex
	refreshWait  sync.WaitGroup
	tokenExpires time.Time
	httpClient   http.Client
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	//User interface{}    `json:"user"`
}

func NewClient(refreshToken string) (*Client, error) {
	client := Client{
		refreshToken: refreshToken,
		httpClient:   http.Client{},
	}

	if err := client.refreshTokens(); err != nil {
		return nil, err
	}

	return &client, nil
}

func (client *Client) RefreshIfExpired() error {
	// If the token is not expired we do nothing
	if time.Now().Before(client.tokenExpires) {
		return nil
	}

	// The first goroutine will acquire the lock
	if client.refreshMutex.TryLock() {
		// and increment the wait group
		client.refreshWait.Add(1)
	} else {
		// The rest of the goroutines will block until the token is refreshed
		client.refreshWait.Wait()
		return nil
	}
	defer client.refreshWait.Done()
	defer client.refreshMutex.Unlock()

	// The first goroutine refreshes the token
	return client.refreshTokens()
}

func (client *Client) refreshTokens() error {
	if client.refreshToken == "" {
		return errors.New("the client doesn't have a refresh token")
	}

	response, err := client.httpClient.PostForm(oAuthURL, url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {client.refreshToken},
		"grant_type":    {"refresh_token"},
	})
	if err != nil {
		return err
	}

	refreshResponse := RefreshResponse{}
	if err = unmarshalJSONFromResponse(response, &refreshResponse); err != nil {
		return err
	}

	client.accessToken = refreshResponse.AccessToken
	client.refreshToken = refreshResponse.RefreshToken
	client.tokenExpires = time.Now().Add(time.Duration(refreshResponse.ExpiresIn) * time.Second)

	return nil
}

func unmarshalJSONFromResponse(response *http.Response, v any) error {
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(responseBody, v)
}
