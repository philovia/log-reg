package models

import "gorm.io/gorm"

type Account struct {
	Fullname         *string `json:"fullname"`
	Email            *string `json:"email"`
	Username         *string `json:"username"`
	Password         *string `json:"password"`
	Confirm_Password *string `json:"confirm_password"`
}
type LoginRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

type Cart struct {
	gorm.Model
	UserID uint
	Items  []CartItem
	Total  float64 // You can add a total price for the cart if needed.
}
type Products struct {
	gorm.Model
	Name     string
	Price    float64
	CartItem CartItem // Add this field to establish a relationship with cart items.
}
type CartItem struct {
	gorm.Model
	ProductID uint
	Quantity  int
}

// func MigrateAccount(db *gorm.DB) error {
// 	// AutoMigrate the Product, Cart, and CartItem models.
// 	if err := db.AutoMigrate(&Products{}, &Cart{}, &CartItem{}, &Account{}); err != nil {
// 		return err
// 	}
// 	return nil
// }
