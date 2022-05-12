package unsplash

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

func jsonGet(u *url.URL, v interface{}) (int, error) {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, err
	}
	return jsonDo(req, v)
}
func jsonDo(req *http.Request, response interface{}) (int, error) {
	resp, err := clientDo(http.DefaultClient, req)
	if err != nil {
		if e, ok := err.(*url.Error); ok {
			if e.Timeout() {
				return http.StatusGatewayTimeout, err
			}
		}
		return http.StatusServiceUnavailable, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Errorf("closing body")
		}
	}()

	if resp.StatusCode != 200 {
		return resp.StatusCode, errors.New(fmt.Sprintf("%d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		logrus.WithError(err).Errorf("decoding json")
	}
	return resp.StatusCode, err
}
func clientDo(client *http.Client, req *http.Request) (_ *http.Response, err error) {
	start := time.Now()
	defer func() {
		entry := logrus.WithField("url", req.URL.String())
		if err != nil {
			entry.WithError(err)
		}
		entry.Debugf("[clientDo] clientDo in %s", time.Since(start))
	}()
	return client.Do(req)
}

func Search(queryTerms string) (models.SearchResult, error) {

	url, _ := url.Parse(fmt.Sprintf("https://api.unsplash.com/search/photos"))

	q := url.Query()
	q.Add("client_id", "523369f28be80c2e16d5412edf9ee9357cb75bd92628886327387fcb833ab31b")
	q.Add("query", queryTerms)
	url.RawQuery = q.Encode()

	u, _ := url.Parse(url.String())
	var ret models.SearchResult
	_, err := jsonGet(u, &ret)
	return ret, err
}
