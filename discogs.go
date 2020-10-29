package discogs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gomodule/oauth1/oauth"
)

const (
	discogsAPI = "https://api.discogs.com"
)

// Options is a set of options to use discogs API client
type Options struct {
	// Discogs API endpoint (optional).
	URL string
	// Currency to use (optional, default is USD).
	Currency string
	// UserAgent to to call discogs api with.
	UserAgent string
	// Token provided by discogs (optional).
	Token string
}

// Discogs is an interface for making Discogs API requests.
type Discogs interface {
	DatabaseService
	SearchService
	UserService
	CollectionService
}

type discogs struct {
	DatabaseService
	SearchService
	UserService
	CollectionService
}

var header *http.Header

// New returns a new discogs API client.
func New(o *Options) (Discogs, error) {
	header = &http.Header{}

	if o == nil || o.UserAgent == "" {
		return nil, ErrUserAgentInvalid
	}

	header.Add("User-Agent", o.UserAgent)

	cur, err := currency(o.Currency)
	if err != nil {
		return nil, err
	}

	// set token, it's required for some queries like search
	if o.Token != "" {
		header.Add("Authorization", "Discogs token="+o.Token)
	}

	if o.URL == "" {
		o.URL = discogsAPI
	}

	return discogs{
		newDatabaseService(o.URL, cur),
		newSearchService(o.URL + "/database/search"),
		newUserService(o.URL),
		newCollectionService(o.URL),
	}, nil
}

// currency validates currency for marketplace data.
// Defaults to the authenticated users currency. Must be one of the following:
// USD GBP EUR CAD AUD JPY CHF MXN BRL NZD SEK ZAR
func currency(c string) (string, error) {
	switch c {
	case "USD", "GBP", "EUR", "CAD", "AUD", "JPY", "CHF", "MXN", "BRL", "NZD", "SEK", "ZAR":
		return c, nil
	case "":
		return "USD", nil
	default:
		return "", ErrCurrencyNotSupported
	}
}

func request(ctx context.Context, path string, params url.Values, resp interface{}) error {
	r, err := http.NewRequest("GET", path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	r.Header = *header

	client := &http.Client{}
	response, err := client.Do(r.WithContext(ctx))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		switch response.StatusCode {
		case http.StatusUnauthorized:
			return ErrUnauthorized
		default:
			return fmt.Errorf("unknown error: %s", response.Status)
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &resp)
}

func requestWithCreds(ctx context.Context, path string, client *oauth.Client, creds *oauth.Credentials, params url.Values, resp interface{}) error {
	response, err := client.GetContext(ctx, creds, path, params)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		switch response.StatusCode {
		case http.StatusUnauthorized:
			return ErrUnauthorized
		default:
			return fmt.Errorf("unknown error: %s", response.Status)
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &resp)
}
