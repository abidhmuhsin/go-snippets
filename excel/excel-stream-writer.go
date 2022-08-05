/*
https://xuri.me/excelize/en/stream.html#NewStreamWriter
NewStreamWriter return stream writer struct by given worksheet name to generate a new worksheet
with large amounts of data. Note that after set rows, you must call the Flush method to end the
streaming writing process and ensure that the order of line numbers is ascending, the common API
and stream API can't be work mixed to writing data on the worksheets. For example, set data for
the worksheet of size 102400 rows x 50 columns with numbers and style:

Each sheet needs to be in memory.. Data may be flushed only per sheet. No streaming within a sheet.
Actual use case refer to ../file-download/dynamic-excel-zipped-download.go
*/

package main

import (
	"fmt"
	"math/rand"

	"github.com/xuri/excelize/v2"
)

func main() {

	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		fmt.Println(err)
	}
	styleID, err := file.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "#777777"}})
	if err != nil {
		fmt.Println(err)
	}
	// test write single row cells
	if err := streamWriter.SetRow("A1", []interface{}{
		excelize.Cell{Value: 1},
		excelize.Cell{Value: 2},
		excelize.Cell{Formula: "SUM(A1,B1)"}},
		//-- set some row options
		excelize.RowOpts{StyleID: styleID, Height: 20, Hidden: false}); err != nil {
		fmt.Println(err)
	}

	//------
	// set data heading
	if err := streamWriter.SetRow("A4", []interface{}{
		excelize.Cell{StyleID: styleID, Value: "Data"}}); err != nil {
		fmt.Println(err)
	}
	streamWriter.MergeCell("A4", "E4")
	// write data
	for rowID := 5; rowID <= 100000; rowID++ {
		row := make([]interface{}, 50)
		for colID := 0; colID < 50; colID++ {
			row[colID] = rand.Intn(640000)
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowID)
		fmt.Println("Streaming:: Writing row ", rowID, cell)
		if err := streamWriter.SetRow(cell, row); err != nil {
			fmt.Println(err)
		}
	}
	if err := streamWriter.Flush(); err != nil {
		fmt.Println(err)
	}
	if err := file.SaveAs("StreamedBook.xlsx"); err != nil {
		fmt.Println(err)
	}

}
