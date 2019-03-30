package efeed

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
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

func crawlProductDetails(p Product) (Product, error) {

	resp, err := getRequest(p.URL, FanaticAPIParams{})
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
		} else if op == "og:brand" {
			p.Brand = con
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
	fmt.Println("gg")
	doc.Find(".value").Find("label").Find(".square").Each(func(i int, s *goquery.Selection) {
		fmt.Println("hehe")
		fmt.Println(s.Text())
		p.Sizes = append(p.Sizes, s.Text())
	})

	//fmt.Println(p.Brand)
	if err != nil {
		return Product{}, fmt.Errorf("error when goquery: %s", err)
	}

	return p, nil
}

func getSoccerRequest(url string) (*http.Response, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return res, nil
}
func crawlProductsPage(url string, option string) ([]Product, error) {

	resp, err := getSoccerRequest(url + option)
	if err != nil {
		return []Product{}, fmt.Errorf("error when getRequest crawlProductsInListingPage: %s", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return []Product{}, fmt.Errorf("error when goquery crawlProductsInListingPage: %s", err)
	}

	/*totalProducts, err := strconv.Atoi(doc.Find(".page-count-quantity").Text())
	if err != nil {
		return []Product{}, fmt.Errorf("error when strconv crawlProductsInListingPage: %s", err)
	}*/

	//numberTotalCrawl := float64(1000)
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
		resp, err := getSoccerRequest(target)
		//fmt.Println(target)
		if err != nil {
			return []Product{}, fmt.Errorf("error when get page %s", err)
		}

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return []Product{}, fmt.Errorf("error when goquery %s", err)
		}
		doc.Find(".products.row").Find("a").Each(func(i int, s *goquery.Selection) {

			link, _ := s.Attr("href")
			fmt.Println(link)
			fmt.Println("here")
			productLink := Product{URL: link, Ranking: rank}
			productsURL = append(productsURL, productLink)
			rank++
		})
	}

	//rank := 1
	//currentPageNum := 1

	return productsURL, nil
}

func RunCrawlerSoccerPro() error {

	for _, element := range productCategoryURLs {
		//fmt.Println(element)
		foundURLs, _ := crawlProductsPage(element, OPTION)
		productList = append(productList, foundURLs...)
	}

	var count = 0
	for _, product := range productList {
		var p Product
		if DB.Where(&Product{URL: product.URL}).First(&p).RecordNotFound() {
			product, err := crawlProductDetails(product)
			if err != nil {
				//fmt.Println("error when product crawlMainPage: ", err)
				continue
			}
			fmt.Println("saving product: ", product.URL)
			fmt.Println(count)

			DB.Create(&product)

		} else {

			fmt.Println("skipping: ", product.URL)

			fmt.Println(count)

		}
		count++

	}
	return nil
}
