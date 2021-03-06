package efeed

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// ProductSearch ProductSearch
type ProductSearch struct {
	Name     string
	Tags     []string
	Category string
	Brand    string
	Site     string
}

// Product struct
type Product struct {
	gorm.Model
	Brand        string         `json:"brand"`
	Category     string         `json:"category"`
	Description  string         `json:"description"`
	Details      pq.StringArray `gorm:"type:varchar(1024)[]"`
	Images       pq.StringArray `gorm:"type:varchar(1024)[]" json:"image"`
	HostedImages pq.StringArray `gorm:"type:varchar(1024)[]"`
	Sizes        pq.StringArray `gorm:"type:varchar(128)[]"`
	Colors       pq.StringArray `gorm:"type:varchar(128)[]"`
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Price        float64        `json:"price"`
	Type         string         `json:"@type"`
	ProductID    string         `json:"productID"`
	Tags         pq.StringArray `gorm:"type:varchar(1024)[]"`
	Ranking      int
	Site         string
}

// FanaticAPIParams FanaticAPIParams
type FanaticAPIParams struct {
	PageSize   string `json:"pageSize"`
	PageNumber string `json:"pageNumber"`
	SortOption string `json:"sortOption"`
}

// Config config
type Config struct {
	EnableProxy bool
	ProxyURL    string
	DoSpaceURL  string
	EnableReuploadImage bool
}
