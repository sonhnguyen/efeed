package efeed

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var data map[string]string

var categoryURLs = map[string]string{
	"NFL":     "https://www.fanatics.com/nfl/o-3572+z-953036859-1253393850",
	"COLLEGE": "https://www.fanatics.com/college/o-27+z-9314487535-1329600116",
	"MLB":     "https://www.fanatics.com/mlb/o-8987+z-80725673-162114610",
	"NBA":     "https://www.fanatics.com/nba/o-1370+z-938737729-293541727",
	"NHL":     "https://www.fanatics.com/nhl/o-2428+z-935562038-1765108222",
	"NASCAR":  "https://www.fanatics.com/nascar/o-3580+z-7979470-3715318076",
	"MLS":     "https://www.fanatics.com/mls/o-3500+z-994320316-3715076868",
	"ESPORT":  "https://www.fanatics.com/esports/x-492706+z-90733431-2457216011",
}

const (
	BASE_URL         = "https://www.fanatics.com"
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

func crawlProductDetailPageJSON(p Product) (Product, error) {
	var sizes []string
	var colors []string
	resp, err := getRequest(p.URL, FanaticAPIParams{})
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
	tags := p.Tags
	tags = append(tags, p.Category)
	tags = append(tags, p.Brand)
	doc.Find(".breadcrumbs-container li").Each(func(i int, s *goquery.Selection) {
		breadcrumb := s.Text()
		tags = append(tags, breadcrumb)
	})
	p.Tags = tags

	doc.Find(".size-selector-list a").Each(func(i int, s *goquery.Selection) {
		size := s.Text()
		sizes = append(sizes, size)
	})
	priceStr := doc.Find(".price-tag div").First().Text()
	price, err := removeCharactersExceptNumbers(priceStr)
	if err != nil {
		return Product{}, fmt.Errorf("error when removeCharactersExceptNumbers: %s", err)
	}
	p.Sizes = sizes
	p.Price = price
	doc.Find(".color-selector-list a").Each(func(i int, s *goquery.Selection) {
		colorStr, _ := s.Attr("aria-label")
		color := strings.Replace(colorStr, ", selected", "", -1)
		colors = append(colors, color)
	})
	p.Colors = colors

	return p, nil
}

func removeCharactersExceptNumbers(str string) (float64, error) {
	re := regexp.MustCompile("[^0-9.]")
	result := re.ReplaceAllString(str, "")
	s, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return 0, fmt.Errorf("erroring parsing float: %s", err)
	}
	return s, nil
}

func crawlProductsInListingPage(gender, url string) ([]Product, error) {
	resp, err := getRequest(url, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: "1"})
	if err != nil {
		return []Product{}, fmt.Errorf("error when getRequest crawlProductsInListingPage: %s", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []Product{}, fmt.Errorf("error when goquery crawlProductsInListingPage: %s", err)
	}

	totalProducts, err := strconv.Atoi(doc.Find(".page-count-quantity").Text())
	if err != nil {
		return []Product{}, fmt.Errorf("error when strconv crawlProductsInListingPage: %s", err)
	}

	numberTotalCrawl := PERCENT_CRAWLING * float64(totalProducts)

	productsURL := []Product{}
	rank := 1
	currentPageNum := 1

	for {
		fmt.Println("crawling next page")
		doc.Find(".product-image-container").Find("a").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")

			productLink := Product{URL: BASE_URL + link, Ranking: rank}
			productLink.Tags = append(productLink.Tags, gender)
			productsURL = append(productsURL, productLink)
			rank++
		})
		fmt.Printf("done crawling: %s, page: %d\n", url, currentPageNum)
		if float64(len(productsURL)) > numberTotalCrawl {
			break
		} else {
			currentPageNum++

			resp, err = getRequest(url, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: strconv.Itoa(currentPageNum)})
			if err != nil {
				return []Product{}, fmt.Errorf("error when getRequest crawlProductsInListingPage: %s, currentPageNum: %d", err, currentPageNum)
			}

			doc, err = goquery.NewDocumentFromResponse(resp)
			if err != nil {
				return []Product{}, fmt.Errorf("error when goquery crawlProductsInListingPage: %s", err)
			}
		}
	}

	return productsURL, nil
}

func crawlAllProductLinksOfTeam(targetURL string) ([]Product, error) {
	var productLink []Product
	genderAgeGroups := map[string]string{}
	resp, err := getRequest(targetURL, FanaticAPIParams{})
	if err != nil {
		return []Product{}, fmt.Errorf("error when getRequest crawlAllProductLinksOfTeam: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []Product{}, fmt.Errorf("error when goquery crawlAllProductLinksOfTeam: %s", err)
	}

	if doc != nil {
		doc.Find(".genderAgeGroups").Find("a").Each(func(i int, s *goquery.Selection) {
			res, _ := s.Attr("href")
			genderAgeGroups[s.Text()] = res
		})
	}

	for gender, link := range genderAgeGroups {
		productsListLink := BASE_URL + link
		productsLinks, err := crawlProductsInListingPage(gender, productsListLink)
		if err != nil {
			return []Product{}, fmt.Errorf("error when productsLinks crawlProductsInListingPage: %s", err)
		}
		fmt.Printf("crawled %d products from link: %s \n", len(productsLinks), link)
		productLink = append(productLink, productsLinks...)
	}

	return productLink, nil
}

func crawlMainPageAndSave(category, targetURL string) error {
	resp, err := getRequest(targetURL, FanaticAPIParams{})
	if err != nil {
		return fmt.Errorf("error when getRequest crawlMainPage: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return fmt.Errorf("error when goquery crawlMainPage: %s", err)
	}

	teamPages := extractTeamLinks(doc)
	for team, teamPageURL := range teamPages {
		var productsURLs []Product

		teamPageURL = BASE_URL + teamPageURL
		productsLinks, err := crawlAllProductLinksOfTeam(teamPageURL)
		if err != nil {
			return fmt.Errorf("error when productsLinks crawlMainPage: %s", err)
		}
		productsURLs = append(productsURLs, productsLinks...)
		fmt.Printf("beginning crawling product details of team %s, number of products: %d \n", teamPageURL, len(productsURLs))

		for _, product := range productsURLs {
			var p Product
			if DB.Where(&Product{URL: product.URL}).First(&p).RecordNotFound() {
				product, err := crawlProductDetailPageJSON(product)
				if err != nil {
					fmt.Println("error when product crawlMainPage: ", err)
					continue
				}
				fmt.Println("saving product: ", product.URL)
				tags := product.Tags
				tags = append(tags, category)
				tags = append(tags, team)
				product.Tags = tags
				DB.Create(&product)
			} else {
				fmt.Println("skipping: ", product.URL)
			}

		}
	}

	return nil
}

func extractTeamLinks(doc *goquery.Document) map[string]string {
	foundLinks := map[string]string{}
	doc.Find(".team-list-column").Find("a").Each(func(i int, s *goquery.Selection) {
		res, _ := s.Attr("href")
		team := s.Text()
		foundLinks[team] = res
	})
	return foundLinks
}

// RunCrawlerFanatics RunCrawlerFanatics
func RunCrawlerFanatics() error {
	fmt.Println("RunCrawlerFanatics")
	for category, url := range categoryURLs {
		err := crawlMainPageAndSave(category, url)
		if err != nil {
			fmt.Printf("error at crawling category: %s\n, error: %s", category, err)
		}
	}
	return nil
}
