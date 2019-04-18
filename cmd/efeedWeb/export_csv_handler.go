package main

import (
	"efeed"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
)

// ExportCSVHandler ExportCSVHandler
func (a *App) ExportCSVHandler() HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		header := w.Header()
		header.Set("Content-Type", "text/csv")
		header.Set("Content-Disposition", "attachment; filename="+"export_products.csv")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")

		productSearch := efeed.ProductSearch{}
		queryValues := req.URL.Query()
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

		results := efeed.QueryProducts(productSearch, -1)

		wr := csv.NewWriter(w)

		if err := wr.Write([]string{"Title", "Description", "Price", "Type", "Option1 Name", "Option1 Value", "Option2 Name", "Option2 Value", "Image Src", "Hosted Images"}); err != nil {
			fmt.Println("error writing record to csv:", err)
		}

		for _, result := range results {
			var record []string
			record = append(record, result.Name)
			record = append(record, result.Description)
			record = append(record, fmt.Sprintf("%f", result.Price))
			record = append(record, result.Type)
			if result.Sizes != nil {
				record = append(record, "Size")
			} else {
				record = append(record, "")
			}
			record = append(record, strings.Join(result.Sizes, ","))

			if result.Colors != nil {
				record = append(record, "Color")
			} else {
				record = append(record, "")
			}
			record = append(record, strings.Join(result.Colors, ","))
			record = append(record, strings.Join(result.Images, ","))
			record = append(record, strings.Join(result.HostedImages, ","))

			if err := wr.Write(record); err != nil {
				fmt.Println("error writing record to csv:", err)
			}
		}
		wr.Flush()

		if err := wr.Error(); err != nil {
			fmt.Println(err)
		}

		return nil
	}
}
