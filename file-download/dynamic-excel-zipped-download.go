package main

import (
	"archive/zip"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/xuri/excelize/v2"
)

var s3Client *minio.Client

func main() {
	setupRoutes()
}

func setupRoutes() {
	http.HandleFunc("/", downloadFileHandler)
	http.HandleFunc("/zip", downloadZipFileHandler)
	err := http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}
	// go func() {
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()
}

func prepareAndReturnExcel() (file *excelize.File) {
	// This method causes very high memory usage for large files.
	// [when working with multiple sheets- all sheets maintained in memory]
	// Approx 1gb for 65mb file.(10sheetx 10000rows x 100cols)

	f := excelize.NewFile()

	for i := 1; i < 10; i++ {
		sheetId := fmt.Sprintf("Sheet-%d", i)
		// Create a new sheet.
		if i == 1 {
			f.SetSheetName("Sheet1", sheetId)
		} else {
			f.NewSheet(sheetId)
		}
		fmt.Println(sheetId)

		for j := 1; j < 1000; j++ {
			for k := 1; k < 100; k++ {

				cellId, _ := excelize.CoordinatesToCellName(k, j)
				f.SetCellValue(
					sheetId,
					cellId,
					fmt.Sprintf("My cell!(%d,%d)", j+1, k))
			}
		}

		runtime.GC()
		PrintMemUsage()
	}
	return f
}

func prepareWithStreamAndReturnExcel() (file *excelize.File) {
	// Low memory usage [when working with multiple sheets -- only current sheet in memory,
	// others in temp folder once flushed]
	// Approx 170mb for 65mb file
	//(10sheetx 10000rows x 100cols ~16mb per sheet irrespective of rows. All sheets data in temp files)

	file = excelize.NewFile()

	for sheetID := 1; sheetID <= 10; sheetID++ {
		sheetName := fmt.Sprintf("Sheet-%d", sheetID)
		if sheetID == 1 {
			file.SetSheetName("Sheet1", sheetName)
		} else {
			file.NewSheet(sheetName)
		}
		fmt.Println(sheetName)

		streamWriter, err := file.NewStreamWriter(sheetName)
		if err != nil {
			fmt.Println(err)
		}
		styleID, err := file.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "#777777"}})
		if err != nil {
			fmt.Println(err)
		}
		if err := streamWriter.SetRow("A1", []interface{}{
			excelize.Cell{StyleID: styleID, Value: "Data"}}); err != nil {
			fmt.Println(err)
		}
		for rowID := 2; rowID <= 1000; rowID++ {
			row := make([]interface{}, 100)
			for colID := 0; colID < 100; colID++ {
				row[colID] = rand.Intn(640000) //fmt.Sprintf("My cell!(%d,%d)", rowID, colID)
			}
			cell, _ := excelize.CoordinatesToCellName(1, rowID)
			if err := streamWriter.SetRow(cell, row); err != nil {
				fmt.Println(err)
			}
		}
		// Flush the stream after each done with each sheet.
		if err := streamWriter.Flush(); err != nil {
			fmt.Println(err)
		}
		runtime.GC()
		PrintMemUsage()
	}
	// if err := file.SaveAs("Book1.xlsx"); err != nil {
	// 	fmt.Println(err)
	// }
	fmt.Println("Sending download")
	return file
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Hit downloadFileHandler")

	downloadName := time.Now().UTC().Format("data-20060102150405.xlsx")

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+downloadName)
	w.WriteHeader(http.StatusOK)

	// excelFile := prepareAndReturnExcelStream()
	excelFile := prepareAndReturnExcel()
	excelFile.WriteTo(w)
	// prepareAndReturnExcelStream(w)
	return
}

func downloadZipFileHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Hit downloadZipFileHandler")

	downloadName := time.Now().UTC().Format("data-20060102150405")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+downloadName+".zip")
	w.WriteHeader(http.StatusOK)

	excelFile := prepareWithStreamAndReturnExcel()
	// excelFile := prepareAndReturnExcel()
	// excelFile.WriteTo(w)

	zipFile := zip.NewWriter(w)
	defer zipFile.Close()

	file1, _ := zipFile.Create(downloadName + ".xlsx") // Create file#1 in zipFile
	// file2, _ := zipFile.Create("another_file.xlsx") //?? Create file#2 in zipFile
	excelFile.Write(file1)
	return
}

//--------------------------------------------
// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
