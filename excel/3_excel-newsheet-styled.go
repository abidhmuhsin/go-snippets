package main

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func main() {

	f := excelize.NewFile()

	f.SetCellValue("Sheet1", "A1", 50)

	idx := f.NewSheet("Sheet2")

	fmt.Println(idx)

	f.SetCellValue("Sheet2", "A1", "2x50")

	f.SetActiveSheet(idx)

	//----styling a cell
	//https://xuri.me/excelize/en/style.html

	f.SetCellValue("Sheet2", "A2", "an old falcon")
	f.SetColWidth("Sheet2", "A", "A", 20)

	style, _ := f.NewStyle(`{"alignment":{"horizontal":"center"}, 
        "font":{"bold":true,"italic":true}}`)

	f.SetCellStyle("Sheet2", "A2", "A2", style)

	//--------merge cells------

	f.SetCellValue("Sheet2", "A4", "Sunny Day")
	f.MergeCell("Sheet2", "A4", "C9")

	style2, _ := f.NewStyle(`
	{
		"alignment":{"horizontal":"center","vertical":"center"},	
		"font":{"bold":true,"italic":true,"family":"Times New Roman","size":36,"color":"#5FD432"},
		"border":[{"type":"left","color":"0000FF","style":2},
				{"type":"top","color":"00FF00","style":3},
				{"type":"bottom","color":"FFFF00","style":4},
				{"type":"right","color":"FF0000","style":5},
				{"type":"diagonalDown","color":"A020F0","style":6},
				{"type":"diagonalUp","color":"A020F0","style":7}
				],
	"fill":{"type":"gradient","color":["#FFFFFF","#E0EBF5"],"shading":1}
	}`)

	f.SetCellStyle("Sheet2", "A4", "C9", style2)

	//---------------

	// Test set border and solid style pattern fill for a single cell.
	style3, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "0000FF", Style: 3},
			{Type: "top", Color: "00FF00", Style: 4},
			{Type: "bottom", Color: "FFFF00", Style: 5},
			{Type: "right", Color: "FF0000", Style: 6},
			// {Type: "diagonalDown", Color: "A020F0", Style: 7},
			// {Type: "diagonalUp", Color: "A020F0", Style: 8},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#45d41g"},
			Pattern: 7,
		},
	})
	f.SetCellStyle("Sheet2", "A1", "A1", style3)

	//--------------------

	if err := f.SaveAs("new_sheet_styled.xlsx"); err != nil {
		log.Fatal(err)
	}
}
