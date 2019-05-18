package efeed

import (
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	REVZILLA_BASE_URL = "https://www.revzilla.com"
)

// RunCrawlerRevzilla RunCrawlerRevzilla
func RunCrawlerRevzilla(config Config, svc *s3.S3) error {
	resp, err := getRequest(config, REVZILLA_BASE_URL, FanaticAPIParams{})
	if err != nil {
		return fmt.Errorf("error when getRequest crawlMainPage: %s", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return fmt.Errorf("error when goquery crawlMainPage: %s", err)
	}

	productPages := ExtractProductsPage(doc)
	for header, e := range productPages {
		println(header)
		println(e)
	}

	for _, pageURL := range productPages {
		var productsURLs []Product
		productsLinks, err := crawlProductLinks(config, pageURL)
		if err != nil {
			return fmt.Errorf("error when productsLinks crawlMainPage: %s", err)
		}
		productsURLs = append(productsURLs, productsLinks...)
		//fmt.Printf("beginning crawling product details of team %s, number of products: %d \n", teamPageURL, len(productsURLs))

	}

	println("Job Done!")
	return nil
}

//ExtractHeaderLinks ExtractHeaderLinks
func ExtractProductsPage(doc *goquery.Document) map[string]string {
	foundLinks := map[string]string{}
	doc.Find(".site-navigation__top-link").Each(func(i int, s *goquery.Selection) {
		res, _ := s.Attr("href")
		header := s.Text()
		foundLinks[header] = REVZILLA_BASE_URL + res + "?view_all=true"
	})
	return foundLinks
}

func crawlProductLinks(config Config, targetURL string) ([]Product, error) {
	var productLinks []Product
	resp, err := getRequest(config, targetURL, FanaticAPIParams{SortOption: SORT_OPTION, PageSize: PAGE_SIZE, PageNumber: "1"})
	if err != nil {
		return []Product{}, fmt.Errorf("error when getRequest crawlProductLinks: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []Product{}, fmt.Errorf("error when goquery crawlProductLinks: %s", err)
	}
	//totalProducts, err = strconv.Atoi(doc.Find())
	/*if doc != nil {
		doc.Find("product-index-results__product-tile-wrapper").Find("a").Each(func(i int, s *goquery.Selection) {
			res, _ := s.Atrr("href")

		}
	}*/
	var totalProducts int
	doc.Find(".browse-header__product-count.product-faceted-browse-index__product-count").Find("span").Each(func(i int, s *goquery.Selection) {
		res, _ := s.Attr("data-product-count")
		if res != "" {
			totalProducts, err = strconv.Atoi(res)
		}
	})

	fmt.Printf("Total products at %s is %d\n", targetURL, totalProducts)
	numberTotalCrawl := PERCENT_CRAWLING * float64(totalProducts)
	println(numberTotalCrawl)
	productsURL := []Product{}
	rank := 1
	currentPageNum := 1
	if totalProducts > 0 {
		for {
			doc.Find(".product-index-results__product-tile-wrapper").Find("a").Each(func(i int, s *goquery.Selection) {
				link, _ := s.Attr("href")
				println(link)
				productLink := Product{URL: REVZILLA_BASE_URL + link, Ranking: rank, Site: REVZILLA_BASE_URL}
				//productLink.Tags = AppendIfMissing(productLink.Tags, gender)
				productsURL = append(productsURL, productLink)
				rank++
			})
			fmt.Printf("done crawling: %s, page: %d\n", targetURL, currentPageNum)
			if float64(len(productsURL)) > numberTotalCrawl {
				break
			} else {
				currentPageNum++
				newURL := targetURL + "#page=" + strconv.Itoa(currentPageNum)
				println(newURL)
				resp, err = getRequest(config, newURL, FanaticAPIParams{})
				if err != nil {
					return []Product{}, fmt.Errorf("error when getRequest crawlProductsPage: %s, currentPageNum: %d", err, currentPageNum)
				}

				doc, err = goquery.NewDocumentFromResponse(resp)
				if err != nil {
					return []Product{}, fmt.Errorf("error when goquery crawlProductsPage: %s", err)
				}
			}
		}
	}
	return productLinks, nil
}
