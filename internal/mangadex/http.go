package mangadex

import (
	"fmt"
	"io"
	"net/http"
)

func request(method, url, referer string) (io.ReadCloser, error) {
	client := &http.Client{Transport: &http.Transport{DisableCompression: true}}

	req, _ := http.NewRequest(method, url, nil)
	if referer != "" {
		req.Header.Add("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received %d response code", resp.StatusCode)
	}

	return resp.Body, nil
}
