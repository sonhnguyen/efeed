package main

import (
	"efeed"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
)

func (a *App) ProductSearchHandler() http.Handler {
	fp := path.Join("views", "list.html")
	tmpl := template.Must(template.ParseFiles(fp))
	fn := func(w http.ResponseWriter, r *http.Request) {

		//tmpl.Execute(w, nil)

		productSearch := efeed.ProductSearch{}
		queryValues := r.URL.Query()
		if value := queryValues.Get("tags"); value != "" {
			tags := strings.Split(value, ",")
			for _, tag := range tags {
				productSearch.Tags = append(productSearch.Tags, strings.TrimSpace(tag))
			}
		}
		if value := queryValues.Get("site"); value != "" {
			productSearch.Site = value
		}
		if value := queryValues.Get("brand"); value != "" {
			productSearch.Brand = value
		}
		if value := queryValues.Get("name"); value != "" {
			productSearch.Name = value
		}
		if value := queryValues.Get("category"); value != "" {
			productSearch.Category = value
		}

		results := efeed.QueryProducts(productSearch, 1000)
		fmt.Println(results)
		tmpl.Execute(w, results)
	}
	return http.HandlerFunc(fn)

}
