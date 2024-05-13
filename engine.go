package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

type Message struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	envPath := os.Getenv("FILE_PATH")
	if envPath == "" {
		envPath = "./host"
		os.Setenv("FILE_PATH", envPath)
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		os.Mkdir(envPath, 0755)
	}

	createTestData(envPath)

	app := fiber.New()

	app.Use(cors.New())

	setupRoutes(app)

	err = app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
	}
}
