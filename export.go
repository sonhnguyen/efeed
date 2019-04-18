package efeed

import (
	"strings"
)

// QueryProducts QueryProducts
func QueryProducts(search ProductSearch, limit int) []Product {
	var result []string
	DB.Model(&Product{}).Pluck("DISTINCT name", &result)
	println(result)

	var products []Product
	DB.Limit(limit).Where("name ILIKE ? AND brand ILIKE ? AND category ILIKE ? AND tags @> ?", "%"+search.Name+"%", "%"+search.Brand+"%", "%"+search.Category+"%", value(search.Tags)).Find(&products)
	return products
}

func value(tags []string) string {
	return "{" + strings.Join(tags, ",") + "}"
}

func GetDistinctValue(name string) []string {
	var target = "DISTINCT " + name
	println(target)
	var result []string
	DB.Model(&Product{}).Pluck(target, &result)
	println(result[0])
	return result
}
