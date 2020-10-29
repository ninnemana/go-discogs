package discogs

import (
	"context"

	"github.com/gomodule/oauth1/oauth"
	"go.opencensus.io/trace"
)

type UserService interface {
	OAuthIdentity(ctx context.Context, options ...Option) (*Identity, error)
}

type userService struct {
	url         string
	oauthClient *oauth.Client
	creds       *oauth.Credentials
}

const (
	oauthIdentityURI = "/oauth/identity"
)

func newUserService(url string) UserService {
	return &userService{
		url: url,
	}
}

type Identity struct {
	ConsumerName string `json:"consumer_name"`
	ID           int64  `json:"id"`
	ResourceURL  string `json:"resource_url"`
	Username     string `json:"username"`
}

func (u *userService) OAuthIdentity(ctx context.Context, options ...Option) (*Identity, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discog/Users.OAuthIdentity")
	defer span.End()

	for _, opts := range options {
		opts(u)
	}

	route := u.url + oauthIdentityURI

	span.AddAttributes(
		trace.StringAttribute("path", route),
	)

	var id Identity

	if err := requestWithCreds(
		ctx,
		route,
		u.oauthClient,
		u.creds,
		nil,
		&id,
	); err != nil {
		span.SetStatus(trace.Status{
			Code: trace.StatusCodeInternal,
		})
		span.AddAttributes(trace.StringAttribute("err", err.Error()))

		return nil, err
	}

	return &id, nil
}
