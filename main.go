package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"iibb-imports/infra"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/fatih/color"
	go_iibb "github.com/fleni/go-iibb"
	"github.com/fleni/go-iibb/mssql"
	"github.com/gosuri/uiprogress"
	"github.com/subosito/gotenv"
)

var templates = template.Must(template.ParseGlob("frontend/*"))

var rowsTABLE = make([]*RowsData, 0)

var (
	listen          = ":" + os.Getenv("API_PORT")
	path            = os.Getenv("PATH_UPLOADS")
	disableColor    = flag.Bool("no-color", false, "Disable color output.")
	disableProgress = flag.Bool("no-progress", false, "Disable progress bars.")
)

var bars *uiprogress.Progress

func init() {
	rowsTABLE := make([]RowsData, 0)
	fmt.Println(rowsTABLE)
	_ = gotenv.Load(".env")
	listen = ":" + os.Getenv("API_PORT")
	path = os.Getenv("PATH_UPLOADS")
}

func main() {
	flag.Parse()
	if *disableColor {
		color.NoColor = true
	}

	// Setup progress bars
	if !*disableProgress {
		bars = uiprogress.New()
		bars.Start()
	}

	infra.SqlConf = &infra.DBData{
		DB_DRIVER:   os.Getenv("DB_DRIVER"),
		DB_USER:     os.Getenv("DB_USERNAME"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_INSTANCE: os.Getenv("DB_INSTANCE"),
		DB_DATABASE: os.Getenv("DB_DATABASE"),
		DB_ENCRYPT:  os.Getenv("DB_ENCRYPT"),
	}
	infra.DbPayment = infra.ConnectDB()
	defer infra.DbPayment.Close()

	// Setup Server
	mux := http.NewServeMux()

	// Serve static files with rice for portability
	staticFiles := rice.MustFindBox("frontend").HTTPBox()
	mux.Handle("/", http.FileServer(staticFiles))
	mux.HandleFunc("/healthcheck", infra.Healthcheck)
	mux.HandleFunc("/inicio", Inicio)
	mux.HandleFunc("/verImports", verImports)
	// mux.HandleFunc("/importTxt", importTxt)
	mux.HandleFunc("/send", func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		fileName := r.Header.Get("X-File-Name")
		if fileName == "" {
			log.Printf(color.RedString("File name not provided"))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		fileExtension := filepath.Ext(fileName)
		if fileExtension != ".txt" && fileExtension != ".TXT" {
			log.Printf(color.RedString("no es un archivo txt"))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		outFile, err := os.Create(fileName)
		defer outFile.Close()
		if err != nil {
			log.Printf(color.RedString("Failed to create file: %v"), err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		bar := bars.AddBar(int(r.ContentLength))
		progWriter := &ProgressWriter{
			Length:       r.ContentLength,
			FileName:     fileName,
			BytesWritten: 0,
			Bar:          bar,
			Writer:       outFile,
		}
		bar.AppendFunc(progWriter.Append())
		bar.PrependFunc(progWriter.Prepend())

		written, err := io.Copy(progWriter, r.Body)
		if err != nil {
			log.Printf(color.RedString("Failed to copy file: %v"), err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			cargarDatos(fileName)
			return
		}

		if *disableProgress {
			log.Printf("Wrote "+color.CyanString("%d")+" bytes to file "+color.CyanString("%s"), written, fileName)
		}
	})

	s := &http.Server{
		Addr:    listen,
		Handler: mux,
	}

	err := os.Chdir(path)
	if err != nil {
		log.Printf(color.RedString("Failed to change to %s, %v"), path, err)
		return
	}

	log.Printf("Starting server on %s", color.CyanString(listen))
	log.Printf("Saving files to %s", color.CyanString(path))

	log.Fatal(s.ListenAndServe())
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func cargarDatos(fileName string) {
	infra.SqlConf = &infra.DBData{
		DB_DRIVER:   os.Getenv("DB_DRIVER"),
		DB_USER:     os.Getenv("DB_USERNAME"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_INSTANCE: os.Getenv("DB_INSTANCE"),
		DB_DATABASE: os.Getenv("DB_DATABASE"),
		DB_ENCRYPT:  os.Getenv("DB_ENCRYPT"),
	}
	dbsetup := mssql.DBSetup{}
	dbsetup.SetDbSetup(os.Getenv("DB_HOST_DV"), infra.SqlConf.DB_USER, infra.SqlConf.DB_PASSWORD, 56, infra.SqlConf.DB_DATABASE, 1500)
	if strings.Contains(fileName, "PadronGral") {
		ch := make(chan go_iibb.CaBa)
		go go_iibb.ProcFileCaba("PadronGralAgosto2022.txt", ch)
		mssql.BulkIibbCaba(ch, 200000)
	} else if strings.Contains(fileName, "PadronRGSPer") {
		chbs := make(chan go_iibb.BsAs)
		go go_iibb.ProcFileBsas("PadronRGSPer082022.TXT", chbs)
		mssql.BulkIibbBsas(chbs, 200000)
	}
}

type ProgressWriter struct {
	Length       int64
	FileName     string
	BytesWritten int64
	Bar          *uiprogress.Bar
	io.Writer
}

type RowsData struct {
	FechaPubPadronBsAs string `json:"FechaPubPadron"`
	CantRegistrosBsAs  string `json:"CantRegistros"`
	FechaPubPadronCABA string `json:"FechaPubPadron"`
	CantRegistrosCABA  string `json:"CantRegistros"`
}

func (writer *ProgressWriter) Write(bytes []byte) (int, error) {
	bars.RefreshInterval = time.Millisecond * 300

	n, err := writer.Writer.Write(bytes)
	writer.BytesWritten += int64(n)
	writer.Bar.Set(int(writer.BytesWritten))

	if err == io.EOF {
		// Slow down, end of a progress bar.
		bars.RefreshInterval = time.Second * 10
	}

	return n, err
}

func (writer *ProgressWriter) Prepend() func(*uiprogress.Bar) string {
	return func(bar *uiprogress.Bar) string {
		return writer.FileName
	}
}

func (writer *ProgressWriter) Append() func(*uiprogress.Bar) string {
	total := byteUnitStr(writer.Length)

	return func(bar *uiprogress.Bar) string {
		completed := byteUnitStr(writer.BytesWritten)
		return bar.CompletedPercentString() + " " + completed + "/" + total
	}
}

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

func byteUnitStr(n int64) string {
	var unit string
	size := float64(n)
	for i := 1; i < len(byteUnits); i++ {
		if size < 1000 {
			unit = byteUnits[i-1]
			break
		}

		size = size / 1000
	}

	return fmt.Sprintf("%.3g %s", size, unit)
}

func Inicio(w http.ResponseWriter, r *http.Request) {
	if len(rowsTABLE) == 0 {
		GetInfoRows(w, r)
	}
	templates.ExecuteTemplate(w, "inicio", nil)
}

func verImports(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "tablas", rowsTABLE)
}

func GetInfoRows(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	rowsCABA, err := infra.DbPayment.Query("SELECT FechaPubPadron, COUNT(FechaPubPadron) as CantReg FROM [FacthosHist].[dbo].[IIBBPadronRSCF] group by FechaPubPadron")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rowsCABA.Close()

	for rowsCABA.Next() {
		bk := new(RowsData)
		err := rowsCABA.Scan(&bk.FechaPubPadronCABA, &bk.CantRegistrosCABA)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		rowsTABLE = append(rowsTABLE, bk)
	}
	if err = rowsCABA.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	rows, err := infra.DbPayment.Query("SELECT FechaPubPadron, COUNT(FechaPubPadron) as CantReg FROM [FacthosHist].[dbo].[IIBBPadronBsAs] group by FechaPubPadron")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	for rows.Next() {
		bk := new(RowsData)
		err := rows.Scan(&bk.FechaPubPadronBsAs, &bk.CantRegistrosBsAs)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		rowsTABLE = append(rowsTABLE, bk)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
}
