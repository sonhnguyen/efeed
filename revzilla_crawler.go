package efeed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
)

const (
	REVZILLA_BASE_URL = "https://www.revzilla.com"
)

type RevzillaData struct {
	ProductID   string `json:"productID"`
	Type        string `json:"@type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Color       string `json:"color"`

	Image struct {
		ContentURL string `json:"contentUrl"`
	} `json:"image"`
	Offers struct {
		Price string `json:"price"`
	} `json:"offers"`
	Brand struct {
		BrandName string `json:"name"`
	} `json:"brand"`
}

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
	var productsURLs []Product
	for _, pageURL := range productPages {

		productsLinks, err := crawlProductLinks(config, pageURL)

		if err != nil {
			return fmt.Errorf("error when productsLinks crawlMainPage: %s", err)
		}
		productsURLs = append(productsURLs, productsLinks...)
		fmt.Printf("beginning crawling product details, number of products: %d \n", len(productsURLs))
	}

	for _, product := range productsURLs {
		var p Product
		if DB.Where(&Product{URL: product.URL}).First(&p).RecordNotFound() {
			_, err := crawlRevzillaProductDetails(config, product)
			if err != nil {
				fmt.Println("error when product crawlMainPage: ", err)
				continue
			}
			//product.Tags = AppendIfMissing(product.Tags, category)
			//product.Tags = AppendIfMissing(product.Tags, team)
			/*for _, link := range product.Images {
				hostedImage, err := UploadToDO(config, "fanatics", link, svc)
				if err != nil {
					fmt.Println("error when product hostedImage: ", err)
					continue
				}
				product.HostedImages = append(product.HostedImages, hostedImage)
			}*/
			//DB.Create(&product)
		} else {
			/*if len(p.HostedImages) != len(p.Images) {
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
			}*/
			fmt.Println("Product already existed")
		}

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

	//fmt.Printf("Total products at %s is %d\n", targetURL, totalProducts)
	numberTotalCrawl := PERCENT_CRAWLING * float64(totalProducts)
	println(numberTotalCrawl)
	productsURL := []Product{}
	rank := 1
	currentPageNum := 1
	if totalProducts > 0 {
		for {
			doc.Find(".product-index-results__product-tile-wrapper").Find("a").Each(func(i int, s *goquery.Selection) {
				link, _ := s.Attr("href")
				//println(link)
				productLink := Product{URL: REVZILLA_BASE_URL + link, Ranking: rank, Site: REVZILLA_BASE_URL}
				//productLink.Tags = AppendIfMissing(productLink.Tags, gender)
				productsURL = append(productsURL, productLink)
				rank++
			})
			fmt.Printf("done crawling: %s, page: %d, productsURL: %d \n", targetURL, currentPageNum, len(productsURL))
			if float64(len(productsURL)) > numberTotalCrawl {
				break
			} else {
				currentPageNum++
				newURL := targetURL + "&page=" + strconv.Itoa(currentPageNum)
				println(newURL)
				resp, err = http.Get(newURL)
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
	return productsURL, nil
}

func crawlRevzillaProductDetails(config Config, p Product) (Product, error) {
	/*var sizes []string
	var colors []string
	var details []string*/
	resp, err := getRequest(config, p.URL, FanaticAPIParams{})
	if err != nil {
		return Product{}, fmt.Errorf("error when crawling: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return Product{}, fmt.Errorf("error when goquery: %s", err)
	}
	productDetails := make([]RevzillaData, 0)
	doc.Find("script[type='application/ld+json']").Last().Each(func(i int, s *goquery.Selection) {
		json.Unmarshal([]byte(s.Text()), &productDetails)
	})
	fmt.Println("productDetails:", productDetails, len(productDetails))

	if len(productDetails) != 0 {
		p.Name = productDetails[0].Name
		p.Price, err = strconv.ParseFloat(productDetails[0].Offers.Price, 64)
		p.Description = productDetails[0].Description
		p.Details = append(p.Details, p.Description)
		p.ProductID = productDetails[0].ProductID
		categoryString := strings.Split(productDetails[0].Category, " > ")
		p.Category = strings.Join(categoryString, ", ")
		p.Brand = productDetails[0].Brand.BrandName
		// colorSet := make(map[string]bool)
		// imageSet := make(map[string]bool)
		// for _, e := range productDetails {
		// 	if !colorSet[e.Color] {
		// 		colorSet[e.Color] = true
		// 		p.Colors = append(p.Colors, e.Color)
		// 	}

		// 	if !imageSet[e.Image.ContentURL] {
		// 		imageSet[e.Image.ContentURL] = true
		// 		p.Images = append(p.Images, e.Image.ContentURL)
		// 	}
		// }
		fmt.Println(p.Price)
	}
	doc.Find("label.option-type__swatch").Each(func(i int, s *goquery.Selection) {
		dataLabel, _ := s.Attr("data-label")
		p.Colors = append(p.Colors, dataLabel)
	})
	doc.Find(".product-show-media-image__thumbnail meta").Each(func(i int, s *goquery.Selection) {
		itemprop, _ := s.Attr("itemprop")
		if itemprop == "contentUrl" {
			content, _ := s.Attr("content")
			p.Images = append(p.Images, content)
		}
	})

	spew.Dump(p)
	return p, nil
}
