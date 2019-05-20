package efeed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/s3"
)

var productCategoryURLs = map[string]string{
	"Shoes":     "https://www.soccerpro.com/product-category/soccer-shoes/",
	"Apparel":   "https://www.soccerpro.com/product-category/soccer-apparel/",
	"Equipment": "https://www.soccerpro.com/product-category/soccer-equipment/",
	"Discount":  "https://www.soccerpro.com/product-category/discount-and-clearance-soccer-gear/",
}

var productList []Product
var productPageURLs []string

const (
	BASE_URL_SOCCER = "https://www.soccerpro.com"
	OPTION          = "?orderby=popularity"
	PAGE            = "page/"
)

func crawlProductDetails(config Config, p Product) (Product, error) {

	resp, err := getRequest(config, p.URL, FanaticAPIParams{})
	if err != nil {
		return Product{}, fmt.Errorf("error when crawling: %s", err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	//Price
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		op, _ := s.Attr("property")
		con, _ := s.Attr("content")
		if op == "product:price:amount" {
			p.Price, _ = strconv.ParseFloat(con, 64)
		} else if op == "og:image" {
			p.Images = append(p.Images, con)
		} else if op == "og:type" {
			p.Type = con
		} else if op == "og:description" {
			p.Description = con
		} else if op == "og:title" {
			p.Name = con
		}

	})
	doc.Find(".value").Find("label").Find(".square").Each(func(i int, s *goquery.Selection) {
		p.Sizes = append(p.Sizes, s.Text())
	})
	doc.Find(".woocommerce-breadcrumb .container span a").Each(func(i int, s *goquery.Selection) {
		p.Tags = append(p.Tags, strings.TrimSpace(strings.ToLower(s.Text())))
	})
	doc.Find("p.sku span").Each(func(i int, s *goquery.Selection) {
		op, _ := s.Attr("itemprop")
		if op == "brand" {
			p.Brand = s.Text()
		}
	})
	if err != nil {
		return Product{}, fmt.Errorf("error when goquery: %s", err)
	}

	return p, nil
}

func crawlProductsPage(config Config, category, url, option string) ([]Product, error) {

	resp, err := getRequest(config, url+option, FanaticAPIParams{})
	if err != nil {
		return []Product{}, fmt.Errorf("error when getRequest crawlProductsInListingPage: %s", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []Product{}, fmt.Errorf("error when goquery crawlProductsInListingPage: %s", err)
	}

	max := 1
	productsURL := []Product{}
	doc.Find(".woocommerce-pagination").Find("a").Each(func(i int, s *goquery.Selection) {
		result, _ := strconv.Atoi(s.Text())
		if result > max {
			max = result
		}

	})
	rank := 1
	for i := 1; i <= max; i++ {
		target := url + "page" + strconv.Itoa(i) + "/" + OPTION
		resp, err := getRequest(config, target, FanaticAPIParams{})

		if err != nil {
			return []Product{}, fmt.Errorf("error when get page %s", err)
		}

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return []Product{}, fmt.Errorf("error when goquery %s", err)
		}
		doc.Find(".products.row").Find("a").Each(func(i int, s *goquery.Selection) {

			link, _ := s.Attr("href")
			productLink := Product{Site: "https://www.soccerpro.com", URL: link, Ranking: rank, Category: category, Tags: []string{strings.TrimSpace(strings.ToLower(category))}}
			productsURL = append(productsURL, productLink)
			rank++
		})
	}

	return productsURL, nil
}

// RunCrawlerSoccerPro RunCrawlerSoccerPro
func RunCrawlerSoccerPro(config Config, svc *s3.S3) error {
	fmt.Println("RunCrawlerSoccerPro:")

	for category, element := range productCategoryURLs {
		foundURLs, _ := crawlProductsPage(config, category, element, OPTION)
		productList = append(productList, foundURLs...)
	}

	for _, product := range productList {
		var p Product
		if DB.Where(&Product{URL: product.URL}).First(&p).RecordNotFound() {
			product, err := crawlProductDetails(config, product)
			if err != nil {
				continue
			}
			if config.EnableReuploadImage {
				for _, link := range product.Images {
					hostedImage, err := UploadToDO(config, "soccerpro", link, svc)
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
						hostedImage, err := UploadToDO(config, "soccerpro", link, svc)
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
	return nil
}
