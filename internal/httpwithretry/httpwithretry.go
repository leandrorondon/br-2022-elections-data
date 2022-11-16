package httpwithretry

import (
	"net/http"
	"time"
)

func Get(url string, retry int) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil && retry > 0 {
		time.Sleep(3 * time.Second)
		return Get(url, retry-1)
	}

	return resp, err
}
