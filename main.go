package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Lord-Developer/bookshelf-api/models"
	"github.com/Lord-Developer/bookshelf-api/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	ID        uint    `json:"id"`
	ISBN      *string `json:"isbn"`
	Title     *string `json:"title"`
	Author    *string `json:"author"`
	Published *int    `json:"published"`
	Pages     *int    `json:"pages"`
	Status    *int    `json:"status,omitempty"`
}

type User struct {
	ID     uint    `json:"id"`
	Name   *string `json:"name"`
	Key    *string `json:"key"`
	Secret *string `json:"secret"`
}

type BookResponse struct {
	Book   Book
	Status *int
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) GetUserInfo(context *fiber.Ctx) error {

	id := context.Params("id")
	user := &models.Users{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id = ?", id).First(user).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the user"})
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
		return err
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
	book := Book{}

	err := context.BodyParser(&book)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"isOk": false,
				"message": "request failed"})
		return err
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
	book := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Delete(book, id)

	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err.Error
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{

		"data":    "Successfully deleted",
		"isOk":    true,
		"message": "ok",
	})
	return nil

}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get books"})
		return err
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
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	bookModel := &models.Books{}
	book := Book{}

	err := context.BodyParser(&book)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Model(bookModel).Where("id = ?", id).Updates(book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not update book",
		})
		return err
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

	isbn := context.Params("isbn")
	book := &models.Books{}
	if isbn == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "isbn cannot be empty",
		})
		return nil
	}

	fmt.Println("the ISBN is", isbn)

	err := r.DB.Where("isbn = ?", isbn).First(book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the book"})
		return err
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

	id := context.Params("id")
	bookModel := &models.Books{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the book"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book id fetched successfully",
		"data":    bookModel,
	})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/signup", r.CreateUser)
	api.Post("/create_books", r.CreateBook)
	api.Delete("delete_book/:id", r.DeleteBook)
	api.Put("update_book/:id", r.UpdateBookStatus)
	api.Get(("/get_book_by_isbn/:isbn"), r.GetBookByISBN)
	api.Get("/get_books/:id", r.GetBookByID)
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
