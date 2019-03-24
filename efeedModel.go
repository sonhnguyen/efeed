package efeed

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// Product struct
type Product struct {
	gorm.Model
	ProductURL  string         `json:"@id"`
	Brand       string         `json:"brand"`
	Category    string         `json:"category"`
	Description string         `json:"description"`
	Images      pq.StringArray `gorm:"type:varchar(64)[]" json:"image"`
	Name        string         `json:"name"`
	URL         string         `json:"url"`
	Price       string         `json:"price"`
	Type        string         `json:"@type"`
	ProductID   string         `json:"productID"`
}

// FanaticAPIParams FanaticAPIParams
type FanaticAPIParams struct {
	PageSize   string `json:"pageSize"`
	PageNumber string `json:"pageNumber"`
	SortOption string `json:"sortOption"`
}
