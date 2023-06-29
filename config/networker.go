package config

import (
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/StepanTita/go-BingDALLE/common/convert"
)

type Networker interface {
	ApiUrl() string
	Proxy() *url.URL
}

type networker struct {
	apiUrl string
	proxy  string
}

func NewNetworker(apiUrl, proxy string) Networker {
	return &networker{
		apiUrl: apiUrl,
		proxy:  proxy,
	}
}

func (n networker) ApiUrl() string {
	return n.apiUrl
}

func getEnvFirstNotEmptyOrNil(names ...string) *string {
	for _, name := range names {
		if os.Getenv(name) != "" {
			return &name
		}
	}
	return nil
}

func (n networker) Proxy() *url.URL {
	if n.proxy == "" {
		proxy := getEnvFirstNotEmptyOrNil("all_proxy", "ALL_PROXY", "https_proxy", "HTTPS_PROXY")
		if proxy == nil {
			return nil
		}
		n.proxy = convert.FromPtr(proxy)
	}

	u, err := url.Parse(n.proxy)
	if err != nil {
		logrus.WithError(errors.Wrapf(err, "failed to parse proxy url: %s", n.proxy)).Error()
		logrus.Warn("Running without proxy...")
		return nil
	}

	// TODO: remove when http.Client would support socks5h
	if u.Scheme == "socks5h" {
		u.Scheme = "socks5"
	}
	return u
}
