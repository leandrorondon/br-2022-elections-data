package httpwithretry

import (
	"context"
	"net/http"
	"time"
)

const defaultSleep = 3 * time.Second

func Get(ctx context.Context, url string, retry int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil && retry > 0 {
		time.Sleep(defaultSleep)

		return Get(ctx, url, retry-1)
	}

	return resp, err
}
