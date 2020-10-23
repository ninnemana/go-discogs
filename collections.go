package discogs

import (
	"context"
	"strings"

	"github.com/gomodule/oauth1/oauth"
	"go.opencensus.io/trace"
)

type CollectionService interface {
	GetFolders(ctx context.Context, username string, options ...Option) (*CollectionResponse, error)
}

type collectionService struct {
	url         string
	oauthClient *oauth.Client
	creds       *oauth.Credentials
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

func (c *collectionService) GetFolders(ctx context.Context, username string, options ...Option) (*CollectionResponse, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs.GetFolders")
	defer span.End()

	for _, opts := range options {
		opts(c)
	}

	route := c.url + strings.Replace(collectionsURI, "{username}", username, 1)

	span.AddAttributes(
		trace.StringAttribute("username", username),
		trace.StringAttribute("route", route),
	)

	var collection CollectionResponse

	if err := requestWithCreds(
		ctx,
		route,
		c.oauthClient,
		c.creds,
		nil,
		&collection,
	); err != nil {
		span.SetStatus(trace.Status{
			Code: trace.StatusCodeInternal,
		})
		span.AddAttributes(trace.StringAttribute("err", err.Error()))

		return nil, err
	}

	return &collection, nil
}
