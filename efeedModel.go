package efeed

import (
	"github.com/jinzhu/gorm"
)

// User struct
type User struct {
	gorm.Model
	Name string
}
