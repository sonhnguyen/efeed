package efeed

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var data map[string]string

const (
	BASE_URL         = "https://www.fanatics.com"
	NFL_URL          = "https://www.fanatics.com/nfl/o-3572+z-953036859-1253393850"
	PERCENT_CRAWLING = 0.1
	SORT_OPTION      = "TopSellers"
	PAGE_SIZE        = "96"
)

func getRequest(url string, params FanaticAPIParams) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")

	if err != nil {
		log.Print(err)
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

	resp, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("error when getRequest %s", url)
	}
	return resp, nil
}

func crawlProductDetailPageJSON(url string) (Product, error) {
	var p Product
	resp, err := getRequest(url, FanaticAPIParams{})
	if err != nil {
		return Product{}, fmt.Errorf("error when crawling: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return Product{}, fmt.Errorf("error when goquery: %s", err)
	}

	doc.Find(".json-ld-pdp").Each(func(i int, s *goquery.Selection) {
		jsonString := s.Text()
		if err := json.Unmarshal([]byte(jsonString), &p); err != nil {
			log.Fatal(err)
		}
	})
	return p, nil
}

func crawlProductsInListingPage(url string) ([]string, error) {
	resp, err := getRequest(url, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: "1"})
	if err != nil {
		return []string{}, fmt.Errorf("error when getRequest crawlProductsInListingPage: %s", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []string{}, fmt.Errorf("error when goquery crawlProductsInListingPage: %s", err)
	}

	totalProducts, err := strconv.Atoi(doc.Find(".page-count-quantity").Text())
	if err != nil {
		return []string{}, fmt.Errorf("error when strconv crawlProductsInListingPage: %s", err)
	}

	numberTotalCrawl := PERCENT_CRAWLING * float64(totalProducts)

	productsURL := []string{}
	currentPageNum := 1

	for {
		fmt.Println("crawling next page")
		doc.Find(".product-image-container").Find("a").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")

			productLink := BASE_URL + link
			productsURL = append(productsURL, productLink)
		})
		fmt.Printf("done crawling: %s, page: %d\n", url, currentPageNum)
		if float64(len(productsURL)) > numberTotalCrawl {
			break
		} else {
			currentPageNum++

			resp, err = getRequest(url, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: strconv.Itoa(currentPageNum)})
			if err != nil {
				return []string{}, fmt.Errorf("error when getRequest crawlProductsInListingPage: %s, currentPageNum: %d", err, currentPageNum)
			}

			doc, err = goquery.NewDocumentFromResponse(resp)
			if err != nil {
				return []string{}, fmt.Errorf("error when goquery crawlProductsInListingPage: %s", err)
			}
		}
	}

	return productsURL, nil
}

func crawlAllProductLinksOfTeam(targetURL string) ([]string, error) {
	var productLink []string
	genderAgeGroups := []string{}
	resp, err := getRequest(targetURL, FanaticAPIParams{})
	if err != nil {
		return []string{}, fmt.Errorf("error when getRequest crawlAllProductLinksOfTeam: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []string{}, fmt.Errorf("error when goquery crawlAllProductLinksOfTeam: %s", err)
	}

	if doc != nil {
		doc.Find(".genderAgeGroups").Find("a").Each(func(i int, s *goquery.Selection) {
			res, _ := s.Attr("href")
			genderAgeGroups = append(genderAgeGroups, res)
		})
	}

	for _, link := range genderAgeGroups {
		productsListLink := BASE_URL + link
		productsLinks, err := crawlProductsInListingPage(productsListLink)
		if err != nil {
			return []string{}, fmt.Errorf("error when productsLinks crawlProductsInListingPage: %s", err)
		}
		fmt.Printf("crawled %d products from link: %s \n", len(productsLinks), link)
		productLink = append(productLink, productsLinks...)
	}

	return productLink, nil
}

func crawlMainPageAndSave(targetURL string) error {
	resp, err := getRequest(targetURL, FanaticAPIParams{})
	if err != nil {
		return fmt.Errorf("error when getRequest crawlMainPage: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return fmt.Errorf("error when goquery crawlMainPage: %s", err)
	}

	teamPages := extractTeamLinks(doc)
	for _, teamPage := range teamPages {
		var productsURLs []string
		var teamProducts []Product

		teamPage = BASE_URL + teamPage
		productsLinks, err := crawlAllProductLinksOfTeam(teamPage)
		if err != nil {
			return fmt.Errorf("error when productsLinks crawlMainPage: %s", err)
		}
		productsURLs = append(productsURLs, productsLinks...)
		fmt.Printf("beginning crawling product details of team %s, number of products: %d \n", teamPage, len(productsURLs))

		for _, url := range productsURLs {
			product, err := crawlProductDetailPageJSON(url)
			if err != nil {
				fmt.Println("error when product crawlMainPage: ", err)
				continue
			}
			teamProducts = append(teamProducts, product)
		}

		fmt.Printf("crawled done team  %s, number of products: %d, saving. \n", teamPage, len(teamProducts))
		DB.Create(&teamProducts)
	}

	return nil
}

func extractTeamLinks(doc *goquery.Document) []string {
	foundLinks := []string{}
	doc.Find(".team-list-column").Find("a").Each(func(i int, s *goquery.Selection) {
		res, _ := s.Attr("href")
		foundLinks = append(foundLinks, res)
	})
	return foundLinks
}

// RunCrawlerFanatics RunCrawlerFanatics
func RunCrawlerFanatics() error {
	fmt.Println("RunCrawlerFanatics")
	err := crawlMainPageAndSave(NFL_URL)
	if err != nil {
		return err
	}

	return nil
}
