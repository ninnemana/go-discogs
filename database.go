package discogs

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go.opencensus.io/trace"
)

const (
	releasesURI = "/releases/"
	artistsURI  = "/artists/"
	labelsURI   = "/labels/"
	mastersURI  = "/masters/"
)

// DatabaseService is an interface to work with database.
type DatabaseService interface {
	// Artist represents a person in the discogs database.
	Artist(ctx context.Context, artistID int) (*Artist, error)
	// ArtistReleases returns a list of releases and masters associated with the artist.
	ArtistReleases(ctx context.Context, artistID int, pagination *Pagination) (*ArtistReleases, error)
	// Label returns a label.
	Label(ctx context.Context, labelID int) (*Label, error)
	// LabelReleases returns a list of Releases associated with the label.
	LabelReleases(ctx context.Context, labelID int, pagination *Pagination) (*LabelReleases, error)
	// Master returns a master release.
	Master(ctx context.Context, masterID int) (*Master, error)
	// MasterVersions retrieves a list of all Releases that are versions of this master.
	MasterVersions(ctx context.Context, masterID int, pagination *Pagination) (*MasterVersions, error)
	// Release returns release by release's ID.
	Release(ctx context.Context, releaseID int) (*Release, error)
	// ReleaseRating retruns community release rating.
	ReleaseRating(ctx context.Context, releaseID int) (*ReleaseRating, error)
}

type databaseService struct {
	url      string
	currency string
}

func newDatabaseService(url string, currency string) DatabaseService {
	return &databaseService{
		url:      url,
		currency: currency,
	}
}

// Release serves relesase response from discogs.
type Release struct {
	Title             string         `json:"title"`
	ID                int            `json:"id"`
	Artists           []ArtistSource `json:"artists"`
	ArtistsSort       string         `json:"artists_sort"`
	DataQuality       string         `json:"data_quality"`
	Thumb             string         `json:"thumb"`
	Community         Community      `json:"community"`
	Companies         []Company      `json:"companies"`
	Country           string         `json:"country"`
	DateAdded         string         `json:"date_added"`
	DateChanged       string         `json:"date_changed"`
	EstimatedWeight   int            `json:"estimated_weight"`
	ExtraArtists      []ArtistSource `json:"extraartists"`
	FormatQuantity    int            `json:"format_quantity"`
	Formats           []Format       `json:"formats"`
	Genres            []string       `json:"genres"`
	Identifiers       []Identifier   `json:"identifiers"`
	Images            []Image        `json:"images"`
	Labels            []LabelSource  `json:"labels"`
	LowestPrice       float64        `json:"lowest_price"`
	MasterID          int            `json:"master_id"`
	MasterURL         string         `json:"master_url"`
	Notes             string         `json:"notes,omitempty"`
	NumForSale        int            `json:"num_for_sale,omitempty"`
	Released          string         `json:"released"`
	ReleasedFormatted string         `json:"released_formatted"`
	ResourceURL       string         `json:"resource_url"`
	Series            []Series       `json:"series"`
	Status            string         `json:"status"`
	Styles            []string       `json:"styles"`
	Tracklist         []Track        `json:"tracklist"`
	URI               string         `json:"uri"`
	Videos            []Video        `json:"videos"`
	Year              int            `json:"year"`
}

func (s *databaseService) Release(ctx context.Context, releaseID int) (*Release, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.Release")
	defer span.End()

	params := url.Values{}
	params.Set("curr_abbr", s.currency)

	path := s.url + releasesURI + strconv.Itoa(releaseID)
	span.AddAttributes(
		trace.StringAttribute("currency", s.currency),
		trace.StringAttribute("path", path),
	)

	var release *Release
	err := request(ctx, path, params, &release)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch release",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(releaseID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}

	return release, nil
}

// ReleaseRating serves response for community release rating request.
type ReleaseRating struct {
	ID     int    `json:"release_id"`
	Rating Rating `json:"rating"`
}

func (s *databaseService) ReleaseRating(ctx context.Context, releaseID int) (*ReleaseRating, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.ReleaseRating")
	defer span.End()

	path := s.url + releasesURI + strconv.Itoa(releaseID) + "/rating"
	span.AddAttributes(trace.StringAttribute("path", path))

	var rating *ReleaseRating
	err := request(ctx, path, nil, &rating)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch release rating",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(releaseID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch release rating: %w", err)
	}

	return rating, nil
}

// Artist resource represents a person in the Discogs database
// who contributed to a Release in some capacity.
// More information https://www.discogs.com/developers#page:database,header:database-artist
type Artist struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Realname       string   `json:"realname"`
	Members        []Member `json:"members,omitempty"`
	Aliases        []Alias  `json:"aliases,omitempty"`
	Namevariations []string `json:"namevariations"`
	Images         []Image  `json:"images"`
	Profile        string   `json:"profile"`
	ReleasesURL    string   `json:"releases_url"`
	ResourceURL    string   `json:"resource_url"`
	URI            string   `json:"uri"`
	URLs           []string `json:"urls"`
	DataQuality    string   `json:"data_quality"`
}

func (s *databaseService) Artist(ctx context.Context, artistID int) (*Artist, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.Artist")
	defer span.End()

	path := s.url + artistsURI + strconv.Itoa(artistID)
	span.AddAttributes(trace.StringAttribute("path", path))

	var artist *Artist
	err := request(ctx, path, nil, &artist)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch artist",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(artistID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch artist: %w", err)
	}

	return artist, nil
}

// ArtistReleases ...
type ArtistReleases struct {
	Pagination Page            `json:"pagination"`
	Releases   []ReleaseSource `json:"releases"`
}

func (s *databaseService) ArtistReleases(ctx context.Context, artistID int, pagination *Pagination) (*ArtistReleases, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.ArtistReleases")
	defer span.End()

	path := s.url + artistsURI + strconv.Itoa(artistID) + "/releases"
	span.AddAttributes(trace.StringAttribute("path", path))

	var releases *ArtistReleases
	err := request(ctx, path, pagination.params(), &releases)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch artist releases",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(artistID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return releases, nil
}

// Label resource represents a label, company, recording studio, location,
// or other entity involved with artists and releases.
type Label struct {
	Profile     string     `json:"profile"`
	ReleasesURL string     `json:"releases_url"`
	Name        string     `json:"name"`
	ContactInfo string     `json:"contact_info"`
	URI         string     `json:"uri"`
	Sublabels   []Sublable `json:"sublabels"`
	URLs        []string   `json:"urls"`
	Images      []Image    `json:"images"`
	ResourceURL string     `json:"resource_url"`
	ID          int        `json:"id"`
	DataQuality string     `json:"data_quality"`
}

func (s *databaseService) Label(ctx context.Context, labelID int) (*Label, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.Label")
	defer span.End()

	path := s.url + labelsURI + strconv.Itoa(labelID)
	span.AddAttributes(trace.StringAttribute("path", path))

	var label *Label
	err := request(ctx, path, nil, &label)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch artist releases",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(labelID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return label, nil
}

// LabelReleases is a list of Releases associated with the label.
type LabelReleases struct {
	Pagination Page            `json:"pagination"`
	Releases   []ReleaseSource `json:"releases"`
}

func (s *databaseService) LabelReleases(ctx context.Context, labelID int, pagination *Pagination) (*LabelReleases, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.LabelReleases")
	defer span.End()

	path := s.url + labelsURI + strconv.Itoa(labelID) + "/releases"
	span.AddAttributes(trace.StringAttribute("path", path))

	var releases *LabelReleases
	err := request(ctx, path, pagination.params(), &releases)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch artist releases",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(labelID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return releases, nil
}

// Master resource represents a set of similar releases.
// Masters (also known as `master releases`) have a `main release` which is often the chronologically earliest.
// More information https://www.discogs.com/developers#page:database,header:database-master-release
type Master struct {
	ID                   int            `json:"id"`
	Styles               []string       `json:"styles"`
	Genres               []string       `json:"genres"`
	Title                string         `json:"title"`
	Year                 int            `json:"year"`
	Tracklist            []Track        `json:"tracklist"`
	Notes                string         `json:"notes"`
	Artists              []ArtistSource `json:"artists"`
	Images               []Image        `json:"images"`
	Videos               []Video        `json:"videos"`
	NumForSale           int            `json:"num_for_sale"`
	LowestPrice          float64        `json:"lowest_price"`
	URI                  string         `json:"uri"`
	MainRelease          int            `json:"main_release"`
	MainReleaseURL       string         `json:"main_release_url"`
	MostRecentRelease    int            `json:"most_recent_release"`
	MostRecentReleaseURL string         `json:"most_recent_release_url"`
	VersionsURL          string         `json:"versions_url"`
	ResourceURL          string         `json:"resource_url"`
	DataQuality          string         `json:"data_quality"`
}

func (s *databaseService) Master(ctx context.Context, masterID int) (*Master, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.Master")
	defer span.End()

	path := s.url + mastersURI + strconv.Itoa(masterID)
	span.AddAttributes(trace.StringAttribute("path", path))

	var master *Master
	err := request(ctx, path, nil, &master)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch artist releases",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(masterID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return master, nil
}

// MasterVersions retrieves a list of all releases that are versions of this master.
type MasterVersions struct {
	Pagination Page      `json:"pagination"`
	Versions   []Version `json:"versions"`
}

func (s *databaseService) MasterVersions(ctx context.Context, masterID int, pagination *Pagination) (*MasterVersions, error) {
	ctx, span := trace.StartSpan(ctx, "ninnemana.discogs/DatabaseService.MasterVersions")
	defer span.End()

	path := s.url + mastersURI + strconv.Itoa(masterID) + "/versions"
	span.AddAttributes(trace.StringAttribute("path", path))

	var versions *MasterVersions
	err := request(ctx, path, pagination.params(), &versions)
	if err != nil {
		RecordError(ctx, ErrorConfig{
			Error:   err,
			Code:    trace.StatusCodeInternal,
			Message: "failed to fetch artist releases",
			Attributes: []trace.Attribute{
				trace.Int64Attribute("id", int64(masterID)),
			},
		})
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return versions, nil
}
