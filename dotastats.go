package dotastats

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

// Errors
var ErrNoRows = errors.New("db: no rows in result set")
var ErrDuplicateRow = errors.New("db: duplicate row found for unique constraint")

func TimeNow() time.Time {
	return time.Now().UTC()
}

// VPGameGet is a communicate with vpgame GET apis
func VPGameGet(url string, params VPGameAPIParams) (http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := req.URL.Query()

	s := reflect.ValueOf(&params).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {

		f := s.Field(i)
		if f.String() != "" {
			q.Add(strings.ToLower(typeOfT.Field(i).Name), f.String())
		}
	}

	req.URL.RawQuery = q.Encode()
	fmt.Println(req.URL.String())
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return *resp, nil
}
