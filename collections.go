package discogs

import (
	"context"
	"strconv"
	"strings"

	"github.com/gomodule/oauth1/oauth"
	"go.opencensus.io/trace"
)

type CollectionService interface {
	GetFolders(ctx context.Context, username string, options ...Option) (*CollectionResponse, error)
	GetFolder(ctx context.Context, args GetFolderArgs, options ...Option) (*Folder, error)
}

type collectionService struct {
	url         string
	oauthClient *oauth.Client
	creds       *oauth.Credentials
}

const (
	foldersURI        = "/users/{username}/collection/folders"
	folderURI         = "/users/{username}/collection/folders/{id}"
	folderReleasesURI = "/users/{username}/collection/folders/{id}/releases"
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

	path := strings.Replace(foldersURI, "{username}", username, 1)

	span.AddAttributes(
		trace.StringAttribute("username", username),
		trace.StringAttribute("path", path),
	)

	var collection CollectionResponse

	if err := requestWithCreds(
		ctx,
		c.url+path,
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

type GetFolderArgs struct {
	ID       int
	Username string
}

func (g GetFolderArgs) TraceAttributes() []trace.Attribute {
	return []trace.Attribute{
		trace.StringAttribute("username", g.Username),
		trace.Int64Attribute("id", int64(g.ID)),
	}
}

// GetFolder retrieves metadata associated with a particular folder.
// /users/{username}/collection/folders/{id}
func (c *collectionService) GetFolder(ctx context.Context, args GetFolderArgs, options ...Option) (*Folder, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs.GetFolder")
	defer span.End()

	for _, opts := range options {
		opts(c)
	}

	path := strings.Replace(folderURI, "{username}", args.Username, 1)
	path = strings.Replace(path, "{id}", strconv.Itoa(args.ID), 1)

	span.AddAttributes(args.TraceAttributes()...)
	span.AddAttributes(trace.StringAttribute("path", path))

	var folder Folder

	if err := requestWithCreds(
		ctx,
		c.url+path,
		c.oauthClient,
		c.creds,
		nil,
		&folder,
	); err != nil {
		span.SetStatus(trace.Status{
			Code: trace.StatusCodeInternal,
		})
		span.AddAttributes(trace.StringAttribute("err", err.Error()))

		return nil, err
	}

	return &folder, nil
}

type FolderReleasesResponse struct {
	Releases []Release `json:"releases"`
}

// GetFolder retrieves metadata associated with a particular folder.
// /users/{username}/collection/folders/{id}/releases
func (c *collectionService) GetFolderReleases(ctx context.Context, args GetFolderArgs, options ...Option) (*FolderReleasesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs.GetFolderReleases")
	defer span.End()

	for _, opts := range options {
		opts(c)
	}

	path := strings.Replace(folderReleasesURI, "{username}", args.Username, 1)
	path = strings.Replace(path, "{id}", strconv.Itoa(args.ID), 1)

	span.AddAttributes(args.TraceAttributes()...)
	span.AddAttributes(trace.StringAttribute("path", path))

	var releases FolderReleasesResponse

	if err := requestWithCreds(
		ctx,
		c.url+path,
		c.oauthClient,
		c.creds,
		nil,
		&releases,
	); err != nil {
		span.SetStatus(trace.Status{
			Code: trace.StatusCodeInternal,
		})
		span.AddAttributes(trace.StringAttribute("err", err.Error()))

		return nil, err
	}

	return &releases, nil
}
