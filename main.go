package main

import (
	// "io/ioutil"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"m/v2/storage"
	// "m/v2/models"

)

// Struct Repository
type Repository struct {
	DB      *gorm.DB
	CartMap map[uint]int
}

// Struct Message
type Message struct {
	Message string `json:"message"`
}

// Struct Register & Log_In
type (
	Account struct {
		Fullname         string `json:"fullname"`
		Email            string `json:"email"`
		Username         string `json:"username"`
		Password         string `json:"password"`
		Confirm_Password string `json:"confirm_password"`
	}

	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)
type CartItem struct {
	gorm.Model
	ProductID uint
	Quantity  int
}

// Struct UpdateAccountRequest
type UpdateAccountRequest struct {
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Struct UpdateUserRequest (by Admin)
type UpdateUserRequest struct {
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}

// Struct Change password
type UpdatePasswordRequest struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Struct Product
type Product struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	ImageData   []byte  `json:"image_data"`
}

type CartItems struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	Product  uint `json:"product"`
	Quantity int  `json:"quantity"`
}

// Struct GetUserDataResponse
type GetUserDataResponse struct {
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}

// Struct Order
type Order struct {
	Fullname   string `json:"fullname"`
	Mobile     string `json:"mobile"`
	Address    string `json:"address"`
	ItemTitle  string `json:"itemTitle"`
	Quantity   int    `json:"quantity"`
	PurchaseID uint   `json:"-"`
}

// HASH
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Create Account
func (r *Repository) CreateAccount(context *fiber.Ctx) error {
	account := Account{}
	err := context.BodyParser(&account)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid request"})
		return err
	}

	if account.Password != account.Confirm_Password {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "passwords do not match"})
		return nil
	}

	//if the username or email already exists
	var existingAccount Account
	err = r.DB.Table("account").Where("username = ? OR email = ?", account.Username, account.Email).First(&existingAccount).Error
	if err == nil {
		context.Status(http.StatusConflict).JSON(
			&fiber.Map{"message": "username or email already exists"})
		return nil
	}

	// Hash the password
	hashedPassword, err := hashPassword(account.Password)
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "error hashing password"})
		return err
	}

	// Create the new account
	newAccount := Account{
		Fullname: account.Fullname,
		Email:    account.Email,
		Username: account.Username,
		Password: hashedPassword,
	}

	err = r.DB.Table("account").Create(&newAccount).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Successfully Registered!!!"})
	return nil
}

// Add Product with Image Upload
func (r *Repository) AddProduct(context *fiber.Ctx) error {
	product := Product{}
	err := context.BodyParser(&product)
	if err != nil {
		// Handle parsing error
		return err
	}

	// Handle image upload
	file, err := context.FormFile("image")
	if err != nil {
		// Handle image upload error
		return err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		// Handle file open error
		return err
	}
	defer src.Close()

	// Read the file data into a byte slice
	imageData, err := ioutil.ReadAll(src)
	if err != nil {
		// Handle read error
		return err
	}

	// Store the image data in the product object
	product.ImageData = imageData

	// Insert the product (including image data) into the database
	if err := r.DB.Table("product").Create(&product).Error; err != nil {
		// Handle database insert error
		return err
	}

	// Return a success response
	return context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Product added successfully"})
}

// Handle purchase submission
func (r *Repository) SubmitPurchase(context *fiber.Ctx) error {
	purchase := Order{}
	err := context.BodyParser(&purchase)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Store the purchase in the database
	err = r.DB.Table("orders").Create(&purchase).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not create purchase"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Purchase saved successfully"})
	return nil
}

// log in
func (r *Repository) Login(context *fiber.Ctx) error {
	loginRequest := LoginRequest{}
	Clientrespones := Account{}

	err := context.BodyParser(&loginRequest)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid request"})
		return err
	}

	err = r.DB.Table("account").Where("username = ?", loginRequest.Username).First(&Clientrespones).Error
	if err != nil {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "Invalid Username or Password"})
		return nil
	}

	// Check if the provided password matches the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(Clientrespones.Password), []byte(loginRequest.Password))
	if err != nil {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "Invalid Username or Password"})
		return nil
	}

	textMessage := Message{}
	textMessage.Message = "Welcome! " + loginRequest.Username
	return context.JSON(textMessage)
}

// Update user account
func (r *Repository) UpdateAccount(context *fiber.Ctx) error {
	var updateRequest UpdateAccountRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Update the user's account details in the database based on the username
	err := r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		Updates(&Account{
			Email: updateRequest.Email,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Account updated successfully"})
	return nil
}

// Update user account by Admin
func (r *Repository) UpdateUser(context *fiber.Ctx) error {
	var updateRequest UpdateUserRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Update the user's account details in the database based on the username
	err := r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		Updates(&Account{
			Fullname: updateRequest.Fullname,
			Email:    updateRequest.Email,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update user"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "User updated successfully"})
	return nil
}

// Update Product by Admin
func (r *Repository) UpdateProductByTitle(context *fiber.Ctx) error {
	title := context.Query("title")

	// Check if the product exists
	var existingProduct Product
	err := r.DB.Table("product").
		Where("title = ?", title).
		First(&existingProduct).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "Product not found"})
		return err
	}

	// Parse the updated product data from the request body
	var updatedProduct Product
	if err := context.BodyParser(&updatedProduct); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	// Update the product in the database
	err = r.DB.Table("product").
		Where("title = ?", title).
		Updates(&Product{
			Title:       updatedProduct.Title,
			Description: updatedProduct.Description,
			Price:       updatedProduct.Price,
			Quantity:    updatedProduct.Quantity,
			ImageData:   updatedProduct.ImageData,
		}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update product"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Product updated successfully"})
	return nil
}

// Change password
func (r *Repository) UpdatePassword(context *fiber.Ctx) error {
	var updateRequest UpdatePasswordRequest
	if err := context.BodyParser(&updateRequest); err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "Invalid request"})
		return err
	}

	var existingAccount Account
	err := r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		First(&existingAccount).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingAccount.Password), []byte(updateRequest.CurrentPassword))
	if err != nil {
		context.Status(http.StatusUnauthorized).JSON(
			&fiber.Map{"message": "Invalid current password"})
		return nil
	}

	// Hash the new password
	hashedPassword, err := hashPassword(updateRequest.NewPassword)
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Error Hasing new password"})
		return err
	}

	// Update the user's password in the database
	err = r.DB.Table("account").
		Where("username = ?", updateRequest.Username).
		Update("password", hashedPassword).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to update password"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Password updated successfully"})
	return nil
}

// Get fullname & email by username
func (r *Repository) GetUserData(context *fiber.Ctx) error {
	username := context.Query("username")

	var userData struct {
		Fullname string `json:"full_name"`
		Email    string `json:"email"`
		Address  string `json:"address"`
	}

	err := r.DB.Table("account").
		Select("fullname, email, address").
		Where("username = ?", username).
		First(&userData).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	return context.JSON(userData)
}

// GetUserData by username
func (r *Repository) GetUserData2(context *fiber.Ctx) error {
	username := context.Query("username")

	var userData struct {
		Fullname string `json:"fullname"`
		Age      int    `json:"age"`
		Address  string `json:"address"`
		Email    string `json:"email"`
	}

	err := r.DB.Table("account").
		Select("fullname, age, address, email").
		Where("username = ?", username).
		First(&userData).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	return context.JSON(userData)
}

// Get all user accounts
func (r *Repository) GetAllAccounts(context *fiber.Ctx) error {
	var accounts []Account

	// Retrieve all user accounts from the database
	err := r.DB.Table("account").Find(&accounts).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve user accounts"})
		return err
	}

	return context.JSON(accounts)
}

// Get all usernames
func (r *Repository) GetAllUsernames(context *fiber.Ctx) error {
	var usernames []string

	// Retrieve all usernames from the database
	err := r.DB.Table("account").Pluck("username", &usernames).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve usernames"})
		return err
	}

	return context.JSON(usernames)
}

// Get all products
func (r *Repository) GetAllProducts(context *fiber.Ctx) error {
	var products []Product

	// Retrieve all products from the database
	err := r.DB.Table("product").Find(&products).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve products"})
		return err
	}

	return context.JSON(products)
}

// Get all Products Titles
func (r *Repository) GetAllProductTitles(context *fiber.Ctx) error {
	var productTitles []string

	// Retrieve all product titles from the database
	err := r.DB.Table("product").Pluck("title", &productTitles).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to retrieve product titles"})
		return err
	}

	return context.JSON(productTitles)
}

// Delete user account by Admin
func (r *Repository) DeleteAccount(context *fiber.Ctx) error {
	username := context.Query("username")

	var existingAccount Account
	err := r.DB.Table("account").
		Where("username = ?", username).
		First(&existingAccount).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "User not found"})
		return err
	}

	err = r.DB.Table("account").
		Where("username = ?", username).
		Delete(&Account{}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to delete user account"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "User account deleted successfully"})
	return nil
}

// Deletes a product by Admin
func (r *Repository) DeleteProduct(context *fiber.Ctx) error {
	// Get the product title from the query parameters
	title := context.Query("title")

	// Check if the product exists
	var existingProduct Product
	err := r.DB.Table("product").
		Where("title = ?", title).
		First(&existingProduct).Error

	if err != nil {
		context.Status(http.StatusNotFound).JSON(
			&fiber.Map{"message": "Product not found"})
		return err
	}

	// Delete the product from the database
	err = r.DB.Table("product").
		Where("title = ?", title).
		Delete(&Product{}).Error

	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "Failed to delete product"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Product deleted successfully"})
	return nil
}

// add product to cart
func (r *Repository) AddToCart(ctx *fiber.Ctx) error {

	item := CartItem{}

	if err := ctx.BodyParser(&item); err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Request failed",
		})
		return err
	}

	r.CartMap[item.ProductID] += item.Quantity

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Product added to cart successfully",
		"data":    r.CartMap,
	})
	return nil
}

// remove product from the cart
func (r *Repository) RemoveFromCart(ctx *fiber.Ctx) error {
	productIDStr := ctx.Params("product_id")

	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Invalid product ID",
		})
		return err
	}

	delete(r.CartMap, uint(productID))

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Product removed from cart successfully",
		"data":    r.CartMap,
	})
	return nil
}

// Routes
func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Log In
	api.Post("/login", r.Login)
	// Create & Add
	api.Post("/create_account", r.CreateAccount)
	api.Post("/add_product", r.AddProduct)
	api.Post("/submit_purchase", r.SubmitPurchase)

	// Update
	api.Put("/update_account", r.UpdateAccount)
	api.Put("/update_password", r.UpdatePassword)
	api.Put("/update_user", r.UpdateUser)
	api.Put("/update_product_by_title", r.UpdateProductByTitle)
	// Get
	api.Get("/get_user_data", r.GetUserData)
	api.Get("/get_userdata", r.GetUserData2)
	api.Get("/get_all_accounts", r.GetAllAccounts)
	api.Get("/get_all_usernames", r.GetAllUsernames)
	api.Get("/get_all_products", r.GetAllProducts)
	api.Get("/get_all_product_titles", r.GetAllProductTitles)

	//Delete
	api.Delete("/delete_account", r.DeleteAccount)
	api.Delete("/delete_product", r.DeleteProduct)

	api.Post("/add_to_cart", r.AddToCart)
	api.Post("/remove_from_cart/product_id", r.RemoveFromCart)
}

// .env
func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SLLMODE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("Could not load the database")
	}
	// Auto-migrate your database tables here
	db.AutoMigrate(
		&Account{},
		&Product{},
		&Order{},
		&CartItem{},
	)

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	r.SetupRoutes(app)
	app.Listen(":8080")
}

// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/gorilla/mux"
// 	"github.com/jinzhu/gorm"
// 	_ "github.com/jinzhu/gorm/dialects/postgres"
// 	"github.com/joho/godotenv"
// 	"golang.org/x/crypto/bcrypt"
// )

// var db *gorm.DB

// type User struct {
// 	gorm.Model
// 	Username string `gorm:"unique;not null"`
// 	Password string `gorm:"not null"`
// }

// func main() {
// 	// Load environment variables from .env file
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatalf("Error loading .env file: %v", err)
// 	}

// 	// Initialize the database connection
// 	dbHost := os.Getenv("DB_HOST")
// 	dbPort := os.Getenv("DB_PORT")
// 	dbUser := os.Getenv("DB_USER")
// 	dbPassword := os.Getenv("DB_PASSWORD")
// 	dbName := os.Getenv("DB_NAME")

// 	dbURI := fmt.Sprintf("localhost=%s 5432=%s postgres=%s postgres=%s postgres=%s sslmode=disable",
// 		dbHost, dbPort, dbUser, dbName, dbPassword)

// 	var err error
// 	db, err = gorm.Open("postgres", dbURI)
// 	if err != nil {
// 		log.Fatalf("Error connecting to the database: %v", err)
// 	}
// 	defer db.Close()

// 	// AutoMigrate creates the 'users' table in the database
// 	db.AutoMigrate(&User{})

// 	// Create a new router
// 	router := mux.NewRouter()

// 	// Define API routes
// 	router.HandleFunc("/register", RegisterHandler).Methods("POST")
// 	router.HandleFunc("/login", LoginHandler).Methods("POST")

// 	// Start the server
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080" // Default port
// 	}
// 	log.Printf("Server is running on port %s", port)
// 	log.Fatal(http.ListenAndServe(":"+port, router))
// }

// // RegisterHandler handles user registration
// func RegisterHandler(w http.ResponseWriter, r *http.Request) {
// 	var user User

// 	// Parse request body
// 	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Create a new user
// 	if err := db.Create(&user).Error; err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Respond with a success message
// 	w.WriteHeader(http.StatusCreated)
// 	fmt.Fprintln(w, "User registered successfully")
// }

// // LoginHandler handles user login
// func LoginHandler(w http.ResponseWriter, r *http.Request) {
// 	var input struct {
// 		Username string `json:"username"`
// 		Password string `json:"password"`
// 	}

// 	// Parse request body
// 	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	var user User

// 	// Find the user by username
// 	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
// 		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
// 		return
// 	}

// 	// Verify the password
// 	if !CheckPasswordHash(input.Password, user.Password) {
// 		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
// 		return
// 	}

// 	// Respond with a success message or JWT token if desired
// 	fmt.Fprintln(w, "Login successful")
// }

// // CheckPasswordHash verifies a password against a hashed password
// func CheckPasswordHash(password, hash string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
// 	return err == nil
// }
