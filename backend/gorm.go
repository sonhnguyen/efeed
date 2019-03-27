package efeed

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// DB is gorm connection
var DB *gorm.DB

// OpenDB new gorm connection
func OpenDB(URI string) (*gorm.DB, error) {
	var err error
	DB, err = gorm.Open("postgres", URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database")
	}
	DB.AutoMigrate(&Product{})
	return DB, nil
}

// CloseDB gorm connection
func CloseDB() error {
	return DB.Close()
}
