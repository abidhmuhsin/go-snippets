package main

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func main() {

	f, err := excelize.OpenFile("read-write-example.xlsx")

	if err != nil {
		log.Fatal(err)
	}

	c1, err := f.GetCellValue("Sheet1", "A1")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(c1)

	c2, err := f.GetCellValue("Sheet1", "A4")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(c2)

	c3, err := f.GetCellValue("Sheet1", "B2")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(c3)

	///-----------------
	// Get all the rows in the Sheet1.
	fmt.Println("-------------------------------------")
	fmt.Println("***Get all the rows in the Sheet1***")
	fmt.Println("-------------------------------------")
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}
	for rowIndex, row := range rows {
		fmt.Println("--_", rowIndex, "-_-")
		for colIndex, colCell := range row {
			fmt.Print("[", colIndex, "] ")
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}
