package main

import (
	"fmt"
	"os"
	"path/filepath"
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

	setupEmptyHost()

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


func setupEmptyHost(){
	// Check if directory exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// Create directory
		errDir := os.MkdirAll(envPath, 0755)
		if errDir != nil {
			panic(err)
		}
	}

	// Create config.json file in the directory
	filePath := filepath.Join(envPath, "config.json")
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write some example data to the file
	file.WriteString(`{
		"setProjectPath": ""
	}`)
}