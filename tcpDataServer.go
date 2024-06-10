package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func handleDataConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		splitMessage := strings.Split(message, " ")

		if len(splitMessage) != 2 {
			fmt.Fprintf(conn, "Invalid request format. Please send 'path index'.\n")
			continue
		}

		path := splitMessage[0]
		index, err := strconv.Atoi(splitMessage[1])
		if err != nil {
			fmt.Fprintf(conn, "Invalid index: %s\n", splitMessage[1])
			continue
		}

		content, ok := contentMap[path]
		if !ok || index < 0 || index >= len(content) {
			fmt.Fprintf(conn, "Invalid path or index: %s %d\n", path, index)
			continue
		}

		rowData, _ := json.Marshal(content[index])
		fmt.Fprintf(conn, "%s\n", rowData)
	}
}

func startTcpDataServer() {
	ln, err := net.Listen("tcp", ":8923")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleDataConnection(conn)
	}
}
