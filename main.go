package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Lord-Developer/golang-bookshelf-api/models"
	"github.com/Lord-Developer/golang-bookshelf-api/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type User struct {
	ID     uint    `json:"id"`
	Name   *string `json:"name"`
	Key    *string `json:"key"`
	Secret *string `json:"secret"`
}

type Book struct {
	ID        uint    `json:"id"`
	ISBN      *string `json:"isbn"`
	Title     *string `json:"title"`
	Author    *string `json:"author"`
	Published *int    `json:"published"`
	Pages     *int    `json:"pages"`
	Status    *int    `json:"status,omitempty"`
}

type BookResponse struct {
	Book   Book
	Status *int
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) Authenticate(context *fiber.Ctx) bool {
	key := context.Get("key", os.DevNull)
	sign := context.Get("sign", os.DevNull)
	if key == os.DevNull || sign == os.DevNull {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{
				"isOk":    false,
				"message": "User is not authenticated!"})
		return false
	}

	user := &models.Users{}
	err := r.DB.Where("key = ? AND secret = ?", key, sign).First(user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "Key or Secret is wrong!"})
		return false
	}
	return true
}

func (r *Repository) GetUserInfo(context *fiber.Ctx) error {
	if !r.Authenticate(context) {
		return nil
	}
	id := context.Params("id")
	user := &models.Users{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"isOk":    false,
			"message": "id cannot be empty",
		})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id = ?", id).First(user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "could not get the user"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"data": &fiber.Map{
			"user": &fiber.Map{
				"id":     user.ID,
				"name":   user.Name,
				"key":    user.Key,
				"secret": user.Secret,
			},
		},
		"isOk":    true,
		"message": "ok",
	})

	return nil
}
func (r *Repository) CreateUser(context *fiber.Ctx) error {
	user := User{}

	err := context.BodyParser(&user)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return nil
	}

	err = r.DB.Create(&user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data": &fiber.Map{
			"user": &fiber.Map{
				"id":     user.ID,
				"name":   user.Name,
				"key":    user.Key,
				"secret": user.Secret,
			},
		},
		"isOk":    true,
		"message": "ok",
	})
	return nil
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	if !r.Authenticate(context) {
		return nil
	}
	book := Book{}

	errObj := context.BodyParser(&book)

	if errObj != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return errObj
	}

	errObj = r.DB.Create(&book).Error
	if errObj != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return errObj
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data": &fiber.Map{
			"book": &fiber.Map{
				"id":        book.ID,
				"isbn":      book.ISBN,
				"title":     book.Title,
				"author":    book.Author,
				"published": book.Published,
				"pages":     book.Pages,
			},
			"status": book.Status,
		},
		"isOk":    true,
		"message": "ok",
	})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	if !r.Authenticate(context) {
		return nil
	}
	book := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"isOk":    false,
			"message": "id cannot be empty",
		})
		return nil
	}

	errObj := r.DB.Delete(book, id)

	if errObj.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return errObj.Error
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data":    "Successfully deleted",
		"isOk":    true,
		"message": "ok",
	})
	return nil

}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	if !r.Authenticate(context) {
		return nil
	}
	bookModels := &[]models.Books{}

	errObj := r.DB.Find(bookModels).Error
	if errObj != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{
				"isOk":    false,
				"message": "could not get books"})
		return errObj
	}

	var books []BookResponse
	for _, book := range *bookModels {
		books = append(books, BookResponse{Book: Book{ID: book.ID, Author: book.Author, Title: book.Title, ISBN: book.ISBN, Published: book.Published, Pages: book.Pages}, Status: book.Status})

	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"data":    books,
		"isOk":    true,
		"message": "ok",
	})
	return nil
}

func (r *Repository) UpdateBookStatus(context *fiber.Ctx) error {
	if !r.Authenticate(context) {
		return nil
	}

	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"isOk":    false,
			"message": "id cannot be empty",
		})
		return nil
	}

	bookModel := &models.Books{}
	book := Book{}

	errIns := context.BodyParser(&book)
	if errIns != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return nil
	}

	errIns = r.DB.Model(bookModel).Where("id = ?", id).Updates(book).Error
	if errIns != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not update book",
		})
		return errIns
	}

	bookObject := &models.Books{}

	errObj := r.DB.Where("id = ?", id).First(bookObject).Error
	if errObj != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the book"})
		return errObj
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data": &fiber.Map{
			"book": &fiber.Map{
				"id":        bookObject.ID,
				"isbn":      bookObject.ISBN,
				"title":     bookObject.Title,
				"author":    bookObject.Author,
				"published": bookObject.Published,
				"pages":     bookObject.Pages,
			},
			"status": bookObject.Status,
		},
		"isOk":    true,
		"message": "ok",
	})
	return nil

}

func (r *Repository) GetBookByISBN(context *fiber.Ctx) error {
	if !r.Authenticate(context) {
		return nil
	}
	isbn := context.Params("isbn")
	book := &models.Books{}
	if isbn == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"isOk":    false,
			"message": "isbn cannot be empty",
		})
		return nil
	}

	errObj := r.DB.Where("isbn = ?", isbn).First(book).Error
	if errObj != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "could not get the book"})
		return nil
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data": &fiber.Map{
			"book": &fiber.Map{
				"id":        book.ID,
				"isbn":      book.ISBN,
				"title":     book.Title,
				"author":    book.Author,
				"published": book.Published,
				"pages":     book.Pages,
			},
			"status": book.Status,
		},
		"isOk":    true,
		"message": "ok",
	})
	return nil

}

func (r *Repository) GetBookByID(context *fiber.Ctx) error {

	if !r.Authenticate(context) {
		return nil
	}

	id := context.Params("id")
	book := &models.Books{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"isOk":    false,
			"message": "id cannot be empty",
		})
		return nil
	}

	errObj := r.DB.Where("id = ?", id).First(book).Error
	if errObj != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "could not get the book"})
		return errObj
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data": &fiber.Map{
			"book": &fiber.Map{
				"id":        book.ID,
				"isbn":      book.ISBN,
				"title":     book.Title,
				"author":    book.Author,
				"published": book.Published,
				"pages":     book.Pages,
			},
			"status": book.Status,
		},
		"isOk":    true,
		"message": "ok",
	})
	return nil

}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app
	api.Post("/signup", r.CreateUser)
	api.Get("/myself/:id", r.GetUserInfo)
	api.Post("books", r.CreateBook)
	api.Delete("books/:id", r.DeleteBook)
	api.Patch("books/:id", r.UpdateBookStatus)
	api.Get(("/books/:isbn"), r.GetBookByISBN)
	api.Get("/book_by_id/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}

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
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}
	err = models.MigrateBooks(db)
	err = models.MigrateUsers(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
