package models

import "gorm.io/gorm"

type Books struct {
	ID        uint    `gorm:"primary key;autoIncrement" json:"id"`
	ISBN      *string `json:"isbn"`
	Title     *string `json:"title"`
	Author    *string `json:"author"`
	Published *int    `json:"published"`
	Pages     *int    `json:"pages"`
	Status    *int    `json:"status"`
}

func MigrateBooks(db *gorm.DB) error {
	err := db.AutoMigrate(&Books{})
	return err
}
