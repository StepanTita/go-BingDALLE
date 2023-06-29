package connector

import (
	"net/url"
)

type RequestParams struct {
	Url    *url.URL
	Params url.Values

	Method  string
	Payload []byte
}
