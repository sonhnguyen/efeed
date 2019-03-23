package efeed

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

var data map[string]string

func getRequest(url string) (*http.Response, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func extractData(doc *goquery.Document) {

	foundName := []string{}
	foundPrice := []string{}
	if doc != nil {

		fmt.Println("here")
		doc.Find("h4").Find("a").Each(func(i int, s *goquery.Selection) {
			//fmt.Println("here2")
			res, _ := s.Attr("title")
			foundName = append(foundName, res)
			//fmt.Println(res)
		})
		fmt.Println("here2")
		doc.Find(".price-tag").Each(func(i int, s *goquery.Selection) {
			//s.Children()
			/*if strings.Contains(res, "Regular: ") {
				fmt.Println("This has discount")
			}*/
			//class, _ := s.Attr("class")
			//foundPrice = append(foundPrice,c)
			res := s.Text()
			foundPrice = append(foundPrice, res)
		})
	}
	data := make(map[string]string)
	for i := range foundName {
		//fmt.Println(foundName[i])
		//fmt.Println(foundPrice[i])
		data[foundName[i]] = foundPrice[i]
	}

	for index, element := range data {
		fmt.Println(index + " " + element)
	}

}

func crawlPage(targetURL string) {

	//fmt.Println("Requesting: ", targetURL)
	resp, _ := getRequest(targetURL)
	//html, err := ioutil.ReadAll(resp.Body)
	/*if err != nil {
		panic(err)
	}
	f, err := os.Create("doc.txt")
	if err != nil {
		fmt.Println(err)

	}
	//f.Write(html)
	//f.Close()
	//fmt.Printf("%s\n", html)
	*/
	doc, _ := goquery.NewDocumentFromResponse(resp)
	extractData(doc)
}

func crawlMainPage(targetURL string) {
	resp, _ := getRequest(targetURL)
	doc, _ := goquery.NewDocumentFromResponse(resp)
	foundLinks := extractLinks(doc)
	baseURL := "https://www.fanatics.com"
	for _, element := range foundLinks {
		//fmt.Println(element)
		element = baseURL + element
		crawlPage(element)
	}
}

func extractLinks(doc *goquery.Document) []string {
	foundLinks := []string{}
	if doc != nil {
		doc.Find(".team-list-column").Find("a").Each(func(i int, s *goquery.Selection) {
			//fmt.Println("here2")
			res, _ := s.Attr("href")
			fmt.Println(res)
			foundLinks = append(foundLinks, res)
			//fmt.Println(res)
		})

	}
	return foundLinks
}

func RunCrawlerFanatics() (int, error) {
	crawlMainPage("https://www.fanatics.com/nfl/o-3572+z-953036859-1253393850")

	return 1, nil
}
