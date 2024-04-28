package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRegisterHandler(t *testing.T) {
	app := fiber.New()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUsername, DBPassword, DBHost, DBPort, DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{})
	setupRoutes(app, db)

	payload := []byte(`{"username": "testuser", "email": "test@example.com", "password": "testpassword"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var responseUser User
	json.NewDecoder(resp.Body).Decode(&responseUser)
	assert.Equal(t, "testuser", responseUser.Username)
	assert.Equal(t, "test@example.com", responseUser.Email)
	// You can add more assertions based on your requirements
}

func TestLoginHandler(t *testing.T) {
	app := fiber.New()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUsername, DBPassword, DBHost, DBPort, DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{})
	setupRoutes(app, db)

	// Hash password for test user
	hashedPassword, _ := hashPassword("testpassword")

	// Register a test user
	db.Create(&User{Username: "admin", Email: "admin@example.com", Password: hashedPassword})

	payload := []byte(`{"username_or_email": "admin", "password": "testpassword"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var responseBody map[string]string
	json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NotNil(t, responseBody["token"])
	// You can add more assertions based on your requirements
}
