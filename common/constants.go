package common

import (
	"fmt"
	"net/http"
)

var FORWARDED_IP = fmt.Sprintf("13.%d.%d.%d", RandFromRange(104, 107), RandFromRange(0, 255), RandFromRange(0, 255))

var HEADERS = http.Header{
	"accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	"accept-language": []string{"en-US,en;q=0.9"},
	"cache-control":   []string{"max-age=0"},
	"content-type":    []string{"application/x-www-form-urlencoded"},
	"referrer":        []string{"https://www.bing.com/images/create/"},
	"origin":          []string{"https://www.bing.com"},
	"user-agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63"},
	"x-forwarded-for": []string{FORWARDED_IP},
}
