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

var envPath string

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	envPath = os.Getenv("FILE_PATH")
	if envPath == "" {
		envPath = "./host"
		//os.Setenv("FILE_PATH", envPath)
	}

	envPort := os.Getenv("PORT")

	if envPort == "" {
		envPort = "4123"
	}

	createTestData(envPath)

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	setupRoutes(app)

	err = app.Listen(":" + envPort)
	if err != nil {
		fmt.Println(err)
	}
}
