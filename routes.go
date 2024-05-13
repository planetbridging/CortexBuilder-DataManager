package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/createfile", func(c *fiber.Ctx) error {
		m := new(Message)
		if err := c.BodyParser(m); err != nil {
			return err
		}

		envPath := os.Getenv("FILE_PATH")
		if !strings.HasPrefix(m.Path, envPath) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid path: "+m.Path)
		}

		err := ioutil.WriteFile(m.Path, []byte(m.Data), 0644)
		if err != nil {
			return err
		}

		return c.SendString("File created successfully")
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				break
			}

			switch string(msg) {
			case "case1":
				fmt.Println("Handling case 1")
				// Handle case 1
			case "case2":
				fmt.Println("Handling case 2")
				// Handle case 2
			case "case3":
				fmt.Println("Handling case 3")
				// Handle case 3
			default:
				fmt.Println("Unknown case:", string(msg))
			}
		}
	}))
}
