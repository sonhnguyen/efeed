package efeed

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/base64"
	"reflect"
	"log"
	"time"
)

func getRequest(config Config, link string, params FanaticAPIParams) (*http.Response, error) {
	var netTransport = &http.Transport{}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Print(err)
	}

	if config.EnableProxy {
		proxyURL, _ := url.Parse(config.ProxyURL)
		netTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(proxyURL.User.String()))
		req.Header.Add("Proxy-Authorization", basicAuth)
	}

	client := &http.Client{
		Transport: netTransport,
		Timeout:   time.Second * 10,
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("error when getRequest %s: %s", link, err)
	}

	q := req.URL.Query()

	s := reflect.ValueOf(&params).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {

		f := s.Field(i)
		if f.String() != "" {
			q.Add(typeOfT.Field(i).Tag.Get("json"), f.String())
		}
	}

	req.URL.RawQuery = q.Encode()

	return resp, nil
}
