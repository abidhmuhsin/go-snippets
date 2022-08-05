package main

import (
	"fmt"
	"net/http"

	"github.com/xuri/excelize/v2"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/download", downloadExcel)
	fmt.Println("Open http://localhost:8090/")
	http.ListenAndServe(":8090", nil)
}

func downloadExcel(w http.ResponseWriter, r *http.Request) {
	// Get the Excel file with the user input data
	file := PrepareAndReturnExcel()

	// Set the headers necessary to get browsers to interpret the downloadable file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=userInputData.xlsx")
	// above two are required headers --
	//https://stackoverflow.com/questions/24116147/how-to-download-file-in-browser-from-go-server
	//You might also want to copy the Content-Length header of the response to the client, to show proper progress.

	w.Header().Set("File-Name", "userInputData.xlsx")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	file.Write(w)
}

func PrepareAndReturnExcel() *excelize.File {
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "Username")
	f.SetCellValue("Sheet1", "A2", "Abidh Muhsin")
	f.SetCellValue("Sheet1", "B1", "Location")
	f.SetCellValue("Sheet1", "B2", "India")
	f.SetCellValue("Sheet1", "C1", "Occupation")
	f.SetCellValue("Sheet1", "C2", "Fullstack Developer")

	return f
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	const template = `
	<html>
		<head>
			<title>Download Excel</title>
		</head>
		<body>
			<a href="download">Download Excel</a>
		</body>
	</html>
	`

	fmt.Fprintf(w, template)
}

// func streamer() {
// 	file := excelize.NewFile()
// 	streamWriter, err := file.NewStreamWriter("Sheet1")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	styleID, err := file.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "#777777"}})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	// test write single row cells
// 	if err := streamWriter.SetRow("A1", []interface{}{
// 		excelize.Cell{Value: 1},
// 		excelize.Cell{Value: 2},
// 		excelize.Cell{Formula: "SUM(A1,B1)"}},
// 		//-- set some row options
// 		excelize.RowOpts{StyleID: styleID, Height: 20, Hidden: false}); err != nil {
// 		fmt.Println(err)
// 	}

// 	//------
// 	// set data heading
// 	if err := streamWriter.SetRow("A4", []interface{}{
// 		excelize.Cell{StyleID: styleID, Value: "Data"}}); err != nil {
// 		fmt.Println(err)
// 	}
// 	streamWriter.MergeCell("A4", "E4")
// 	// write data
// 	for rowID := 5; rowID <= 1000; rowID++ {
// 		row := make([]interface{}, 50)
// 		for colID := 0; colID < 50; colID++ {
// 			row[colID] = rand.Intn(640000)
// 		}
// 		cell, _ := excelize.CoordinatesToCellName(1, rowID)
// 		fmt.Println("Streaming:: Writing row ", rowID, cell)
// 		if err := streamWriter.SetRow(cell, row); err != nil {
// 			fmt.Println(err)
// 		}
// 	}
// 	if err := streamWriter.Flush(); err != nil {
// 		fmt.Println(err)
// 	}
// 	if err := file.SaveAs("StreamedBook.xlsx"); err != nil {
// 		fmt.Println(err)
// 	}
// }
