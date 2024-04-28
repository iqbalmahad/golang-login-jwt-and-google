package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User model
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Email    string `gorm:"unique"`
	Password string
}

const (
	DBUsername = "root"
	DBPassword = "admin"
	DBName     = "go_login_jwt"
	DBHost     = "localhost"
	DBPort     = "3306"
)

func main() {
	// Initialize GORM
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUsername, DBPassword, DBHost, DBPort, DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
	// Migrate the schema
	db.AutoMigrate(&User{})

	// Initialize Fiber
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	// Register routes
	setupRoutes(app, db)

	// Start the server
	log.Fatal(app.Listen(":3000"))
}

func setupRoutes(app *fiber.App, db *gorm.DB) {
	// Register handlers
	app.Post("/register", registerHandler(db))
	app.Post("/login", loginHandler(db))
}

func registerHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user User
		if err := c.BodyParser(&user); err != nil {
			return err
		}
		// Check if username or email already exists
		var existingUser User
		if err := db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error; err == nil {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "Username or email already exists"})
		}
		// Hash password (you should use a proper password hashing library)
		user.Password = "hashed_password"
		hashPassword(user.Password)
		// Create the user
		db.Create(&user)

		return c.JSON(user)
	}
}

func loginHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginData struct {
			UsernameOrEmail string `json:"username_or_email"`
			Password        string `json:"password"`
		}
		if err := c.BodyParser(&loginData); err != nil {
			return err
		}

		// Check if username or email exists
		var user User
		if err := db.Where("username = ? OR email = ?", loginData.UsernameOrEmail, loginData.UsernameOrEmail).First(&user).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}

		// Compare hashed password
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
		if err != nil {
			// Password mismatch
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}

		// Password matches, generate JWT token
		token := generateJWT(user.ID)

		return c.JSON(fiber.Map{"token": token})
	}
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func generateJWT(userID uint) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours

	tokenString, _ := token.SignedString([]byte("your_secret_key"))
	return tokenString
}
