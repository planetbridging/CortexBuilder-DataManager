package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func createTestData(envPath string) {
	csvFilePath := envPath + "/data.csv"
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		file, err := os.Create(csvFilePath)
		if err != nil {
			fmt.Println("error creating file:", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		headers := []string{"input1", "input2", "input3", "output1", "output2", "output3"}
		writer.Write(headers)

		writeData(writer, 10) // Writes 10 rows of data
	}
}

func writeData(writer *csv.Writer, numRows int) {
	for i := 1; i <= numRows; i++ {
		inputs := []string{strconv.Itoa(i), strconv.Itoa(i + 1), strconv.Itoa(i + 2)}
		outputs := []string{strconv.Itoa(2 * i), strconv.Itoa(2 * (i + 1)), strconv.Itoa(2 * (i + 2))}
		data := append(inputs, outputs...)
		writer.Write(data)
	}
	writer.Flush() // Flush the buffer
}
