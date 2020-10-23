package discogs

import "github.com/gomodule/oauth1/oauth"

type Option func(interface{})

func WithCredentials(creds *oauth.Credentials) Option {
	return func(c interface{}) {
		switch t := c.(type) {
		case *collectionService:
			t.creds = creds
		case *userService:
			t.creds = creds
		}
	}
}

func WithClient(client *oauth.Client) Option {
	return func(c interface{}) {
		switch t := c.(type) {
		case *collectionService:
			t.oauthClient = client
		case *userService:
			t.oauthClient = client
		}
	}
}
