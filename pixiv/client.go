package pixiv

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"
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
    tokenExpires time.Time
}

func NewClient(refreshToken string) (*Client, error) {
    client := Client{ refreshToken: refreshToken }

    if err := client.RefreshTokens(); err != nil {
        return nil, err
    }

    return &client, nil
}

func (client *Client) RefreshIfExpired() error {
    if time.Now().After(client.tokenExpires) {
        return client.RefreshTokens()
    }

    return nil
}

func (client *Client) RefreshTokens() error {
    if client.refreshToken == "" {
        return errors.New("the client doesn't have a refresh token")
    }

    response, err := http.PostForm(oAuthURL, url.Values {
        "client_id":     { clientID },
        "client_secret": { clientSecret },
        "refresh_token": { client.refreshToken },
        "grant_type":    { "refresh_token" },
    })

    if err != nil { return err }

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil { return err }

    refreshResponse := RefreshResponse{}
    if err = json.Unmarshal(responseBody, &refreshResponse); err != nil { return err }

    client.accessToken  = refreshResponse.AccessToken
    client.refreshToken = refreshResponse.RefreshToken
    client.tokenExpires = time.Now().Add(time.Duration(refreshResponse.ExpiresIn))

    return nil
}