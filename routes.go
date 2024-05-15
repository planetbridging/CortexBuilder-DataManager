package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type File struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Type string `json:"type"`
}

type WSMessage struct {
	Action string `json:"action"`
	Path   string `json:"path"`
}

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/createfile", func(c *fiber.Ctx) error {
		m := new(Message)
		if err := c.BodyParser(m); err != nil {
			return err
		}

		envPath = os.Getenv("FILE_PATH")
		if !strings.HasPrefix(m.Path, envPath) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid path: "+m.Path)
		}

		err := ioutil.WriteFile(m.Path, []byte(m.Data), 0644)
		if err != nil {
			return err
		}

		return c.SendString("File created successfully")
	})

	app.Post("/createfolder", func(c *fiber.Ctx) error {
		m := new(Message)
		if err := c.BodyParser(m); err != nil {
			return err
		}

		//envPath := os.Getenv("FILE_PATH")
		if !strings.HasPrefix(m.Path, envPath) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid path: "+m.Path)
		}

		err := os.MkdirAll(m.Path, 0755)
		if err != nil {
			return err
		}

		return c.SendString("Folder created successfully")
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				break
			}

			remoteAddr := c.RemoteAddr().String()
			ip, port, _ := net.SplitHostPort(remoteAddr)

			// Create a map to hold the result.
			result := make(map[string]string)

			result["ip"] = ip
			result["port"] = port

			m := new(WSMessage)
			err = json.Unmarshal(msg, m)
			if err != nil {
				fmt.Println("Invalid JSON:", err)
				break
			}

			switch m.Action {
			case "mount":
				mRes := mountFile(m.Path)

				result["cmd"] = "mount"
				result["result"] = mRes
				result["path"] = m.Path
				jsonData, _ := json.Marshal(result)

				c.WriteMessage(websocket.TextMessage, jsonData)

			case "unmount":
				unmountFile(m.Path)
				result["cmd"] = "unmount"
				result["result"] = m.Path
				result["path"] = m.Path
				jsonData, _ := json.Marshal(result)

				c.WriteMessage(websocket.TextMessage, jsonData)
			case "status":
				mountedFiles, err := getStatus()
				if err != nil {
					fmt.Println("Error getting status:", err)
				} else {
					jsonData, _ := json.Marshal(mountedFiles)
					c.WriteMessage(websocket.TextMessage, jsonData)
				}
			case "sysinfo":
		
				// Get some basic computer specs
				os := runtime.GOOS
				arch := runtime.GOARCH
				numCPU := runtime.NumCPU()

	
				result["cachePath"] = envPath
				result["os"] = os
				result["arch"] = arch
				result["numCPU"] = strconv.Itoa(numCPU)
	
				jsonData, _ := json.Marshal(result)
	
				c.WriteMessage(websocket.TextMessage, jsonData)
			default:
				fmt.Println("Unknown action:", m.Action)
				result["cmd"] = "unknown"
				jsonData, _ := json.Marshal(result)
				c.WriteMessage(websocket.TextMessage, jsonData)
			}
		}
	}))

	app.Get("/path/*", func(c *fiber.Ctx) error {
		//envPath := os.Getenv("FILE_PATH")
		path := c.Params("*")
		if path == "" {
			path = envPath // Default to envPath if no path is provided
		} else {
			path = filepath.Join(envPath, path) // Join envPath and requested path
		}

		// Check if the path is within the envPath
		if !strings.HasPrefix(path, envPath) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid path: "+path)
		}

		// Add your security checks here

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		var fileList []File
		for _, f := range files {
			fileType := "file"
			if f.IsDir() {
				fileType = "directory"
			}
			fileList = append(fileList, File{
				Name: f.Name(),
				Size: fmt.Sprintf("%.2f MB", float64(f.Size())/1024.0/1024.0),
				Type: fileType,
			})
		}

		return c.JSON(fileList)
	})

	app.Get("/row", func(c *fiber.Ctx) error {
		path := c.Query("path")
		indexStr := c.Query("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid index: "+indexStr)
		}
		//fmt.Println(contentMap)
		content, ok := contentMap[path]
		if !ok || index < 0 || index >= len(content) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid path or index: "+path+"/"+indexStr)
		}

		return c.JSON(content[index])
	})

}
