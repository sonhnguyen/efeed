package main

import (
	"efeed"
	"html/template"
	"net/http"
	"path"
	"strings"
)

func (a *App) ProductSearchHandler() http.Handler {
	fp := path.Join("views", "list.html")
	tmpl := template.Must(template.New("list.html").Funcs(template.FuncMap{"StringsJoin": strings.Join}).ParseFiles(fp))

	fn := func(w http.ResponseWriter, r *http.Request) {

		productSearch := efeed.ProductSearch{}
		queryValues := r.URL.Query()
		if value := queryValues.Get("tags"); value != "" {
			tags := strings.Split(value, ",")
			for _, tag := range tags {
				productSearch.Tags = append(productSearch.Tags, strings.ToLower(strings.TrimSpace(tag)))
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
		tmpl.Execute(w, results)
	}
	return http.HandlerFunc(fn)

}
