package httpwithretry

import (
	"net/http"
)

func Get(url string, retry int) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil && retry > 0 {
		return Get(url, retry-1)
	}

	return resp, err
}
