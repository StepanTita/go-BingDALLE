package connector

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/StepanTita/go-BingDALLE/common"
	"github.com/StepanTita/go-BingDALLE/config"
)

type Connector interface {
	Request(ctx context.Context, r RequestParams) (*http.Response, int, error)
}

type connector struct {
	log *logrus.Entry

	cfg config.Config

	client http.Client
}

func New(cfg config.Config) Connector {
	return &connector{
		log: cfg.Logging().WithField("service", "[CONN]"),

		cfg: cfg,
		client: http.Client{
			Timeout: 200 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Request TODO: add retry policy
func (c connector) Request(ctx context.Context, r RequestParams) (*http.Response, int, error) {
	c.log.Debugf("Requesting, %s...", r.Url.String())

	c.client.Transport = &http.Transport{Proxy: http.ProxyURL(c.cfg.Proxy())}

	req, err := http.NewRequestWithContext(ctx, r.Method, fmt.Sprintf("%s?%s", r.Url.String(), r.Params.Encode()), bytes.NewReader(r.Payload))
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to create new request")
	}

	req.Header = common.HEADERS
	req.AddCookie(&http.Cookie{
		Name:    "_U",
		Value:   c.cfg.UCookie(),
		Expires: time.Now().Add(365 * 24 * time.Hour), // TODO: make sure this will not be an issue
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to do client request")
	}
	c.log.WithField("status_code", resp.StatusCode).Debug("request completed with the status code")

	return resp, resp.StatusCode, nil
}
