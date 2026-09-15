package openlibrary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
)

type OpenLibrary struct{}

const apiBaseUrl string = "https://openlibrary.org"

// see https://openlibrary.org/dev/docs/api/covers
const CoverBaseUrl string = "https://covers.openlibrary.org"
const resultsPerPage int = 20

func NewOpenLibrary() OpenLibrary {
	return OpenLibrary{}
}

// Make an API request at the provided `path` and the given params `p`.
//
// If the method returns no error, it is guaranteed that `resp` now contains the JSON response.
func (o *OpenLibrary) apiRequest(path string, p map[string]string, resp interface{}) error {
	targetUrl, err := url.Parse(fmt.Sprintf("%s%s", apiBaseUrl, path))
	if err != nil {
		return errors.New("failed to parse api uri")
	}

	// Query params
	params := url.Values{}
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	targetUrl.RawQuery = params.Encode()
	res, err := http.DefaultClient.Get(targetUrl.String())
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return errors.New("failed to load response body")
	}

	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		slog.Error("openlibrary request failed:", "status_code", res.StatusCode)
		return errors.New(string(body))
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		slog.Error("failed to parse response body", "err", err)
		return errors.New("failed to parse response body into JSON struct")
	}

	return nil
}

// Search books.
//
// For reference, see [Search Docs].
//
// [Search Docs]: https://openlibrary.org/dev/docs/api/search
func (o *OpenLibrary) Search(query string, pageNum int) (BookSearchResponse, error) {
	p := map[string]string{
		"q": query,
		// list of fields is documented at https://github.com/internetarchive/openlibrary/blob/b4afa14b0981ae1785c26c71908af99b879fa975/openlibrary/plugins/worksearch/schemes/works.py#L38-L91
		"fields": "key,author_key,author_name,isbn,lending_edition_s,publish_year,publisher,title,ratings_average,id_doi,subject",
		"page":   strconv.Itoa(pageNum),
		"limit":  strconv.Itoa(resultsPerPage),
	}

	resp := new(OpenLibrarySearchResponse)
	if err := o.apiRequest("/search.json", p, resp); err != nil {
		return BookSearchResponse{}, err

	}

	var books []entity.Book
	for _, doc := range resp.Docs {
		olid := strings.Replace(doc.Key, "/works/", "", 1)

		var releaseDate *time.Time
		if len(doc.PublishYear) > 0 {
			year := slices.Min(doc.PublishYear)
			d := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			releaseDate = &d
		}

		books = append(books, entity.Book{
			OLID:  olid,
			ISBN:  strings.Join(doc.Isbn, "|"),
			Title: doc.Title,
			// ratings are on a scale of 1-5, so normalize them to 1-10
			RatingAverage: 2 * doc.RatingsAverage,
			// see https://openlibrary.org/dev/docs/api/covers
			Genres: strings.Join(doc.Subject, "|"),

			AuthorNames: strings.Join(doc.AuthorName, "|"),
			AuthorIDs:   strings.Join(doc.AuthorKey, "|"),

			ReleaseDate: releaseDate,
		})
	}

	return BookSearchResponse{
		NumPages:   resp.NumFound/resultsPerPage + 1,
		NumResults: resp.NumFound,
		Books:      books,
	}, nil
}

func extractDescription(description any) string {
	// see the description field's documentation about this
	switch desc := description.(type) {
	case string:
		return desc
	case map[string]interface{}:
		return desc["value"].(string)
	default:
		// no description available
		return ""
	}
}

func (o *OpenLibrary) workToBook(olid string, work *OpenLibraryWorkDetailsResponse) (entity.Book, error) {
	ratingsResp := new(OpenLibraryRatingsResponse)
	path := fmt.Sprintf("/works/%s/ratings.json", olid)
	if err := o.apiRequest(path, map[string]string{}, ratingsResp); err != nil {
		return entity.Book{}, err
	}

	var releaseDate *time.Time
	if releaseDateParsed, err := time.Parse(time.RFC3339Nano, work.Created.Value); err == nil {
		releaseDate = &releaseDateParsed
	}

	return entity.Book{
		OLID:          olid,
		Title:         work.Title,
		Storyline:     util.MdToHTMLSafe(extractDescription(work.Description)),
		Genres:        strings.Join(work.Subjects, "|"),
		ReleaseDate:   releaseDate,
		RatingAverage: ratingsResp.Summary.Average,
		RatingCount:   ratingsResp.Summary.Count,
	}, nil
}

// Get a book by its Open Library ID (olid).
//
// This is documented at [Works API].
//
// [Works API]: https://openlibrary.org/dev/docs/api/books
func (o *OpenLibrary) GetBookDetails(olid string) (entity.Book, error) {
	detailsResp := new(OpenLibraryWorkDetailsResponse)
	path := fmt.Sprintf("/works/%s.json", olid)
	if err := o.apiRequest(path, map[string]string{}, detailsResp); err != nil {
		return entity.Book{}, err
	}

	book, err := o.workToBook(olid, detailsResp)
	if err != nil {
		return entity.Book{}, err
	}

	// load and append author information to the book
	var authorIDs []string
	var authorNames []string
	for _, authorMetaInfo := range detailsResp.Authors {
		authorId := strings.Replace(authorMetaInfo.Author.Key, "/authors/", "", 1)

		authorInfo, err := o.GetAuthorDetails(authorId)
		if err == nil {
			authorIDs = append(authorIDs, authorId)
			authorNames = append(authorNames, authorInfo.Name)
		}
	}
	book.AuthorIDs = strings.Join(authorIDs, "|")
	book.AuthorNames = strings.Join(authorNames, "|")

	return book, nil
}

func (o *OpenLibrary) GetAuthorDetails(olid string) (domain.Author, error) {
	authorResp := new(OpenLibraryAuthorResponse)
	path := fmt.Sprintf("/authors/%s.json", olid)
	if err := o.apiRequest(path, map[string]string{}, authorResp); err != nil {
		return domain.Author{}, err
	}

	var birthDate *time.Time
	if date, err := time.Parse("02 January 2006", authorResp.BirthDate); err == nil {
		birthDate = &date
	}

	var deathDate *time.Time
	if date, err := time.Parse("02 January 2006", authorResp.DeathDate); err == nil {
		deathDate = &date
	}

	name := authorResp.FullerName
	if strings.TrimSpace(name) == "" {
		name = authorResp.Name
	}

	var homepage *string
	if len(authorResp.Links) > 0 {
		homepage = &authorResp.Links[0].URL
	}

	coverUrl := fmt.Sprintf("%s/a/olid/%s-M.jpg", CoverBaseUrl, olid)
	return domain.Author{
		ID:        olid,
		Name:      name,
		Biography: authorResp.Bio.Value,
		BirthDate: birthDate,
		DeathDate: deathDate,
		Homepage:  homepage,
		Photo:     &coverUrl,
	}, nil
}

func (o *OpenLibrary) GetAuthorCredits(olid string) ([]entity.Book, error) {
	authorWorksResp := new(OpenLibraryAuthorWorksResponse)
	path := fmt.Sprintf("/authors/%s/works.json", olid)
	if err := o.apiRequest(path, map[string]string{}, authorWorksResp); err != nil {
		return []entity.Book{}, err
	}

	var books []entity.Book
	for _, work := range authorWorksResp.Entries {
		var releaseDate *time.Time
		d, err := time.Parse(time.RFC3339Nano, work.Created.Value)
		if err != nil {
			releaseDate = &d
		}

		book := entity.Book{
			OLID:        extractDescription(strings.Replace(work.Key, "/works/", "", 1)),
			Genres:      strings.Join(work.Subjects, "|"),
			Title:       work.Title,
			Storyline:   extractDescription(work.Description),
			ReleaseDate: releaseDate,
		}
		books = append(books, book)
	}

	return books, nil
}

type OpenLibrarySearchResponse struct {
	NumFound         int               `json:"num_found"`
	Start            int               `json:"start"`
	DocumentationURL string            `json:"documentation_url"`
	Q                string            `json:"q"`
	Offset           any               `json:"offset"`
	Docs             []OpenLibraryBook `json:"docs"`
}

type OpenLibraryBook struct {
	Key            string   `json:"key"`
	Title          string   `json:"title"`
	AuthorKey      []string `json:"author_key,omitempty"`
	AuthorName     []string `json:"author_name,omitempty"`
	Isbn           []string `json:"isbn,omitempty"`
	PublishYear    []int    `json:"publish_year,omitempty"`
	Publisher      []string `json:"publisher,omitempty"`
	Subject        []string `json:"subject,omitempty"`
	RatingsAverage float64  `json:"ratings_average,omitempty"`
}

type OpenLibraryWorkDetailsResponse struct {
	Title string `json:"title"`
	// 'description' can be either:
	// - a simple string
	// - a json object like `{type: "/type/text", value: "actual description"}`
	Description any    `json:"description"`
	Key         string `json:"key"`
	Authors     []struct {
		Author struct {
			Key string `json:"key"`
		} `json:"author"`
		Type struct {
			Key string `json:"key"`
		} `json:"type"`
	} `json:"authors"`
	Subjects []string `json:"subjects"`
	Created  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
}

type OpenLibraryRatingsResponse struct {
	Summary struct {
		Average  float64 `json:"average"`
		Count    int     `json:"count"`
		Sortable float64 `json:"sortable"`
	} `json:"summary"`
}

type OpenLibraryAuthorResponse struct {
	Links []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Type  struct {
			Key string `json:"key"`
		} `json:"type"`
	} `json:"links"`
	RemoteIds struct {
		Viaf         string `json:"viaf"`
		Wikidata     string `json:"wikidata"`
		Isni         string `json:"isni"`
		Amazon       string `json:"amazon"`
		Goodreads    string `json:"goodreads"`
		Bookbrainz   string `json:"bookbrainz"`
		Musicbrainz  string `json:"musicbrainz"`
		Imdb         string `json:"imdb"`
		LcNaf        string `json:"lc_naf"`
		Librarything string `json:"librarything"`
		OpacSbn      string `json:"opac_sbn"`
	} `json:"remote_ids"`
	Type struct {
		Key string `json:"key"`
	} `json:"type"`
	Key        string `json:"key"`
	BirthDate  string `json:"birth_date"`
	DeathDate  string `json:"death_date"`
	FullerName string `json:"fuller_name"`
	Bio        struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"bio"`
	AlternateNames []string `json:"alternate_names"`
	Name           string   `json:"name"`
}

type OpenLibraryAuthorWorksResponse struct {
	Size    int                              `json:"size"`
	Entries []OpenLibraryWorkDetailsResponse `json:"entries"`
}

type BookSearchResponse struct {
	NumPages   int           `json:"numPages"`
	NumResults int           `json:"numResults"`
	Books      []entity.Book `json:"books"`
}
