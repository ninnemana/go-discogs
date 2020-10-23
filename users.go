package discogs

import (
	"context"
	"strings"

	"github.com/gomodule/oauth1/oauth"
)

type UserService interface {
	CollectionService
}

type CollectionService interface {
	GetFolders(ctx context.Context, creds oauth.Credentials, username string) (*CollectionResponse, error)
}

type collectionService struct {
	url string
}

const (
	collectionsURI = "/users/{username}/collection/folders"
)

func newCollectionService(url string) CollectionService {
	return &collectionService{
		url: url,
	}
}

type CollectionResponse struct {
	Folders []Folder `json:"folders"`
}

type Folder struct {
	ID          int    `json:"id"`
	Count       int    `json:"count"`
	Name        string `json:"name"`
	ResourceURL string `json:"resource_url"`
}

func (c *collectionService) GetFolders(ctx context.Context, creds oauth.Credentials, username string) (*CollectionResponse, error) {
	var collection CollectionResponse

	if err := requestWithCreds(
		c.url+strings.Replace(collectionsURI, "{username}", username, 1),
		creds,
		nil,
		&collection,
	); err != nil {
		return nil, err
	}

	return &collection, nil
}
