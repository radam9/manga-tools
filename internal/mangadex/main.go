package mangadex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/radam9/manga-tools/internal/model"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"
	"time"
)

const baseURL = "https://mangadex.org"
const pagingLimit = 500

type Client struct {
	title    string
	mangaID  uuid.UUID
	language string
	// rateLimiter rate limiter for the '/at-home' endpoint which has a rate limit of 40 calls per minute,
	// if we exceed this limit we get a 429, and the consequent chapters fail. This may eventually lead to an IP ban.
	rateLimiter <-chan time.Time
}

func NewClient(mangaID uuid.UUID, lang string) *Client {
	// we set the rate limit at 39 calls per minute instead of 40 to make sure the rate limit is under the threshold,
	// otherwise we occasionally get hit by the rate limiter.
	return &Client{mangaID: mangaID, language: lang, rateLimiter: time.Tick(time.Minute / 39)}
}

func (c *Client) FetchTitle() (string, error) {
	if c.title != "" {
		return c.title, nil
	}

	u := fmt.Sprintf("https://api.mangadex.org/manga/%s", c.mangaID.String())
	rBody, err := request(http.MethodGet, u, baseURL)
	if err != nil {
		return "", err
	}
	defer rBody.Close()

	// decode json response
	body := mangaResponse{}
	if err = json.NewDecoder(rBody).Decode(&body); err != nil {
		return "", err
	}

	if c.language != "" {
		trans := body.Data.Attributes.AltTitles.GetTitleByLang(c.language)

		if trans != "" {
			c.title = trans
			return c.title, nil
		}
	}

	// fallback to english
	c.title = body.Data.Attributes.Title["en"]
	return c.title, nil
}

func (c Client) FetchChapterList() ([]model.Chapter, []error) {
	var chapters []model.Chapter
	var errs []error
	offset := 0

	for {
		uri := fmt.Sprintf("https://api.mangadex.org/manga/%s/feed", c.mangaID.String())
		params := url.Values{}
		params.Add("limit", fmt.Sprint(pagingLimit))
		params.Add("order[volume]", "asc")
		params.Add("order[chapter]", "asc")
		params.Add("offset", fmt.Sprint(offset))
		if c.language != "" {
			params.Add("translatedLanguage[]", c.language)
		}
		uri = fmt.Sprintf("%s?%s", uri, params.Encode())

		rBody, err := request(http.MethodGet, uri, "")
		if err != nil {
			errs = append(errs, err)
			return chapters, errs
		}
		defer rBody.Close()

		body := feedReponse{}
		if err = json.NewDecoder(rBody).Decode(&body); err != nil {
			errs = append(errs, err)
			return chapters, errs
		}

		for _, c := range body.Data {
			num, _ := strconv.ParseFloat(c.Attributes.Chapter, 64)
			vol, _ := strconv.Atoi(c.Attributes.Volume)
			chapters = append(chapters, model.Chapter{
				ID:         c.Id,
				Title:      c.Attributes.Title,
				Number:     num,
				Volume:     vol,
				PagesCount: c.Attributes.Pages,
				Language:   c.Attributes.TranslatedLanguage,
			})
		}

		if len(body.Data) == 0 {
			break
		}
		offset += pagingLimit
	}
	return chapters, errs
}

func (c Client) FetchChapterInfo(chapter *model.Chapter) error {
	<-c.rateLimiter

	u := fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", chapter.ID)
	rBody, err := request(http.MethodGet, u, "")
	if err != nil {
		return err
	}

	body := pagesFeedResponse{}
	if err = json.NewDecoder(rBody).Decode(&body); err != nil {
		return err
	}

	chapter.PagesCount = len(body.Chapter.Data)

	for i, p := range body.Chapter.Data {
		chapter.Pages = append(chapter.Pages, model.Page{
			Number: i + 1,
			URL:    body.BaseUrl + path.Join("/data", body.Chapter.Hash, p),
		})
	}

	return nil
}

func (c Client) FetchChapterPages(chapterNumber float64, chapterID string, pages []model.Page, maxPagesConcurrency int) ([]model.Page, error) {
	wg := sync.WaitGroup{}
	var result []model.Page

	slog.Info("downloading pages", "chapterNumber", chapterNumber, "chapterID", chapterID)
	guard := make(chan struct{}, maxPagesConcurrency)

	for _, page := range pages {
		guard <- struct{}{}
		wg.Add(1)
		go func(page model.Page) {
			defer wg.Done()

			data, err := c.FetchFile(page.URL, baseURL)
			if err != nil {
				slog.Error("downloading page", "pageNumber", page.Number, "url", page.URL, "chapterNumber", chapterNumber, "chapterID", chapterID, "error", err)
				return
			}

			result = append(result, model.Page{Number: page.Number, URL: page.URL, Data: data})

			<-guard
		}(page)
	}
	wg.Wait()
	close(guard)

	model.SortPagesByNumber(result)
	return result, nil
}

func (c Client) FetchFile(uri string, referer string) (io.Reader, error) {
	body, err := request(http.MethodGet, uri, referer)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var data bytes.Buffer
	_, err = io.Copy(&data, body)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

type mangaResponse struct {
	Id   string
	Data struct {
		Attributes struct {
			Title     map[string]string
			AltTitles altTitles
		}
	}
}

// altTitles is a slice of maps with the language as key and the title as value
type altTitles []map[string]string

// GetTitleByLang returns the title in the given language (or empty if string is not found)
func (a altTitles) GetTitleByLang(lang string) string {
	for _, t := range a {
		val, ok := t[lang]
		if ok {
			return val
		}
	}
	return ""
}

type feedReponse struct {
	Data []struct {
		Id         string
		Attributes struct {
			Volume             string
			Chapter            string
			Title              string
			TranslatedLanguage string
			Pages              int
		}
	}
}

type pagesFeedResponse struct {
	BaseUrl string
	Chapter struct {
		Hash      string
		Data      []string
		DataSaver []string
	}
}
