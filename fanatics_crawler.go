package efeed

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/s3"
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

func crawlProductDetailPageJSON(config Config, p Product) (Product, error) {
	var sizes []string
	var colors []string
	var details []string
	resp, err := getRequest(config, p.URL, FanaticAPIParams{})
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
	p.Tags = AppendIfMissing(p.Tags, p.Category)
	p.Tags = AppendIfMissing(p.Tags, p.Brand)
	doc.Find(".breadcrumbs-container li").Each(func(i int, s *goquery.Selection) {
		breadcrumb := s.Text()
		p.Tags = AppendIfMissing(p.Tags, breadcrumb)
	})

	doc.Find(".size-selector-list a").Each(func(i int, s *goquery.Selection) {
		size := s.Text()
		sizes = append(sizes, size)
	})

	doc.Find(".description-box-content li").Each(func(i int, s *goquery.Selection) {
		detail := s.Text()
		details = append(details, detail)
	})

	priceStr := doc.Find(".price-tag span").First().Text()
	price, err := removeCharactersExceptNumbers(priceStr)
	if err != nil {
		return Product{}, fmt.Errorf("error when removeCharactersExceptNumbers: %s", err)
	}
	p.Sizes = sizes
	p.Price = price
	p.Details = details
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

func crawlProductsInListingPage(config Config, gender, url string) ([]Product, error) {
	resp, err := getRequest(config, url, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: "1"})
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
		doc.Find(".product-image-container").Find("a").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")

			productLink := Product{URL: BASE_URL + link, Ranking: rank, Site: "https://www.fanatics.com"}
			productLink.Tags = AppendIfMissing(productLink.Tags, gender)
			productsURL = append(productsURL, productLink)
			rank++
		})
		fmt.Printf("done crawling: %s, page: %d\n", url, currentPageNum)
		if float64(len(productsURL)) > numberTotalCrawl {
			break
		} else {
			currentPageNum++

			resp, err = getRequest(config, url, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: strconv.Itoa(currentPageNum)})
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

func crawlAllProductLinksOfTeam(config Config, targetURL string) ([]Product, error) {
	var productLink []Product
	genderAgeGroups := map[string]string{}
	resp, err := getRequest(config, targetURL, FanaticAPIParams{})
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
		productsLinks, err := crawlProductsInListingPage(config, gender, productsListLink)
		if err != nil {
			return []Product{}, fmt.Errorf("error when productsLinks crawlProductsInListingPage: %s", err)
		}
		fmt.Printf("crawled %d products from link: %s \n", len(productsLinks), link)
		productLink = append(productLink, productsLinks...)
	}

	return productLink, nil
}

func crawlMainPageAndSave(category, targetURL string, svc *s3.S3, config Config) error {
	resp, err := getRequest(config, targetURL, FanaticAPIParams{})
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
		productsLinks, err := crawlAllProductLinksOfTeam(config, teamPageURL)
		if err != nil {
			return fmt.Errorf("error when productsLinks crawlMainPage: %s", err)
		}
		productsURLs = append(productsURLs, productsLinks...)
		fmt.Printf("beginning crawling product details of team %s, number of products: %d \n", teamPageURL, len(productsURLs))

		for _, product := range productsURLs {
			var p Product
			if DB.Where(&Product{URL: product.URL}).First(&p).RecordNotFound() {
				product, err := crawlProductDetailPageJSON(config, product)
				if err != nil {
					fmt.Println("error when product crawlMainPage: ", err)
					continue
				}
				product.Tags = AppendIfMissing(product.Tags, category)
				product.Tags = AppendIfMissing(product.Tags, team)
				if config.EnableReuploadImage {
					for _, link := range product.Images {
						hostedImage, err := UploadToDO(config, "fanatics", link, svc)
						if err != nil {
							fmt.Println("error when product hostedImage: ", err)
							continue
						}
						product.HostedImages = append(product.HostedImages, hostedImage)
					}
				}
				DB.Create(&product)
			} else {
				if config.EnableReuploadImage {
					if len(p.HostedImages) != len(p.Images) {
						var images []string
						for _, link := range p.Images {
							hostedImage, err := UploadToDO(config, "fanatics", link, svc)
							if err != nil {
								fmt.Println("error when product hostedImage: ", err)
								continue
							}
							images = append(images, hostedImage)
						}
						p.HostedImages = images
						DB.Save(&p)
					}
				}
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
func RunCrawlerFanatics(config Config, svc *s3.S3) error {
	fmt.Println("RunCrawlerFanatics")
	for category, url := range categoryURLs {
		err := crawlMainPageAndSave(category, url, svc, config)
		if err != nil {
			fmt.Printf("error at crawling category: %s\n, error: %s", category, err)
		}
	}
	return nil
}

// AppendIfMissing AppendIfMissing
func AppendIfMissing(slice []string, str string) []string {
	str = strings.TrimSpace(strings.ToLower(str))
	for _, ele := range slice {
		if ele == str {
			return slice
		}
	}
	return append(slice, str)
}
