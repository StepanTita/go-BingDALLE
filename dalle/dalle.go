package dalle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/StepanTita/go-BingDALLE/config"
	"github.com/StepanTita/go-BingDALLE/services/connector"
)

var linksReg = regexp.MustCompile(`src="([^"]+)"`)

type Bot interface {
	CreateImages(ctx context.Context, prompt string) (<-chan PollStatus, error)
}

type bot struct {
	log *logrus.Entry

	cfg config.Config

	conn connector.Connector
}

func New(cfg config.Config) Bot {
	return &bot{
		log: cfg.Logging().WithField("service", "[DALLE-BOT]"),
		cfg: cfg,

		conn: connector.New(cfg),
	}
}

func (b bot) CreateImages(ctx context.Context, prompt string) (<-chan PollStatus, error) {
	b.log.WithField("prompt", prompt).Debug("Creating images with prompt")

	createUrl, err := url.Parse(fmt.Sprintf("%s/images/create", b.cfg.ApiUrl()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api endpoint url")
	}

	requestParams := connector.RequestParams{
		Url:    createUrl,
		Method: http.MethodPost,
		Params: url.Values{
			"q":    []string{prompt},
			"rt":   []string{"3"},
			"FORM": []string{"GENCRE"},
		},
		Payload: []byte(fmt.Sprintf("q=%s&qs=ds", prompt)),
	}

	resp, statusCode, err := b.conn.Request(ctx, requestParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	if statusCode != http.StatusFound {
		// if rt=4 failed, retry with rt=3
		b.log.Warn("rt=4 failed, retrying with rt=3")
		resp, err = b.retry(ctx, requestParams)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create request with rt=3")
		}
	}

	rawRedirectUrl := strings.ReplaceAll(resp.Header.Get("Location"), "&nfy=1", "")
	redirectUrl, err := url.ParseRequestURI(fmt.Sprintf("%s%s", b.cfg.ApiUrl(), rawRedirectUrl))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse redirect url: %s", rawRedirectUrl)
	}

	if _, _, err = b.conn.Request(ctx, connector.RequestParams{Url: redirectUrl}); err != nil {
		return nil, errors.Wrapf(err, "failed to follow redirect: %s", rawRedirectUrl)
	}

	redirectParts := strings.Split(rawRedirectUrl, "id=")
	requestId := redirectParts[len(redirectParts)-1]

	imagesLinks, err := b.pollImages(ctx, requestId, prompt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to poll images")
	}

	return imagesLinks, nil
}

func (b bot) retry(ctx context.Context, params connector.RequestParams) (*http.Response, error) {
	params.Params.Set("rt", "3")
	resp, statusCode, err := b.conn.Request(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retry")
	}

	if statusCode != http.StatusFound {
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, errors.Wrapf(err, "failed to decode body with status code: %d", statusCode)
		}
		b.log.WithFields(logrus.Fields{
			"body":   body,
			"status": statusCode,
		}).Error("create image request failed")

		if _, ok := body["text"]; ok {
			switch strings.ToLower(body["text"].(string)) {
			case responsePromptReview:
				return nil, ErrResponsePromptReview
			case responsePromptBlocked:
				return nil, ErrResponsePromptBlocked
			case responseUnsupportedLang:
				return nil, ErrResponseUnsupportedLang
			}
		}

		return nil, errors.New(fmt.Sprintf("failed to follow redirect with status: %d", statusCode))
	}

	return resp, nil
}

func (b bot) pollImages(ctx context.Context, requestId, prompt string) (<-chan PollStatus, error) {
	pollingUrl, err := url.Parse(fmt.Sprintf("%s/images/create/async/results/%s", b.cfg.ApiUrl(), requestId))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse poll url")
	}
	requestParams := connector.RequestParams{
		Url:    pollingUrl,
		Method: http.MethodGet,
		Params: url.Values{
			"q": []string{prompt},
		},
	}

	pollChannel := make(chan PollStatus)

	respText := ""

	go func() {
		defer close(pollChannel)

		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(200*time.Second))
		defer cancel()

		var resp *http.Response
		var statusCode int
		err = nil
	out:
		for {
			select {
			case <-deadlineCtx.Done():
				return
			case <-time.Tick(1 * time.Second):
				resp, statusCode, err = b.conn.Request(deadlineCtx, requestParams)
				if err != nil {
					b.log.WithError(err).Error("failed to create poll images request")
					break out
				}

				var body []byte
				body, err = io.ReadAll(resp.Body)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						b.log.WithError(err).Error(fmt.Sprintf("failed to decode body with status code: %d", statusCode))
						break out
					}
				}

				if statusCode != http.StatusOK {
					b.log.WithFields(logrus.Fields{
						"status": statusCode,
						"body":   body,
					}).Error("poll images request failed")
					b.log.WithError(err).Error("failed to poll images")
					break out
				}

				if len(body) == 0 || strings.Contains(string(body), "errorMessage") {
					time.Sleep(1 * time.Second)
					continue
				} else {
					respText = string(body)
					break out
				}
			}
		}

		if err != nil {
			pollChannel <- PollStatus{
				Err: err,
			}
			return
		}

		imageLinks := linksReg.FindAllStringSubmatch(respText, -1)
		//  Remove size limit
		normalImageLinksSet := make(map[string]bool)
		for _, link := range imageLinks {
			normalImageLinksSet[strings.Split(link[1], "?w=")[0]] = true
		}

		// Remove duplicates
		normalImageLinks := make([]string, 0, len(normalImageLinksSet))
		for k := range normalImageLinksSet {
			normalImageLinks = append(normalImageLinks, k)
		}

		pollChannel <- PollStatus{
			Links: normalImageLinks,
		}
	}()

	return pollChannel, nil
}
