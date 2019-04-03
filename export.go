package efeed

import "strings"

// QueryProducts QueryProducts
func QueryProducts(search ProductSearch) []Product {
	var products []Product
	DB.Where("name LIKE ? AND brand LIKE ? AND category LIKE ? AND tags @> ?", "%"+search.Name+"%", "%"+search.Brand+"%", "%"+search.Category+"%", value(search.Tags)).Find(&products)
	return products
}

func value(tags []string) string {
	return "{" + strings.Join(tags, ",") + "}"
}
