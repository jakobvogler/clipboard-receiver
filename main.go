package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	qrcode "github.com/skip2/go-qrcode"
)

var (
	startTime                   = time.Now()
	transferTokenPrinted        = false
	registrationCompletePrinted = false
	transferToken               = randomToken()
	port                        = 4999
)

func main() {
	urlPtr := flag.String("url", "localhost", "Your tunnel url.")
	flag.IntVar(&port, "port", 4999, "The port the server should run on, default is 4999.")
	flag.Parse()

	pageLink := fmt.Sprintf("https://%s", *urlPtr)

	pageLinkCode, err := qrcode.New(pageLink, qrcode.High)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("\n", 100))
	fmt.Println(pageLinkCode.ToString(true))
	fmt.Println("Code 1")

	http.HandleFunc("/lib/qr-scanner.umd.min.js", handleLib)
	http.HandleFunc("/lib/qr-scanner-worker.min.js", handleLibWorker)

	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/api/data", handleData)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

//go:embed html/lib/qr-scanner.umd.min.js
var libFile []byte

func handleLib(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/javascript")
	w.Write(libFile)
}

//go:embed html/lib/qr-scanner-worker.min.js
var libWorkerFile []byte

func handleLibWorker(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/javascript")
	w.Write(libWorkerFile)
}

//go:embed html/dashboard.html
var dashboardFile []byte

func handleDashboard(w http.ResponseWriter, req *http.Request) {
	hydrated := strings.ReplaceAll(string(dashboardFile), "{{START_TIME}}", startTime.Format(time.RFC3339))

	w.Write(([]byte)(hydrated))
	printTransferTokenCode()
}

type RegisterRequest struct {
	Token string `json:"token"`
}

func handleRegister(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error registering device")
	}

	w.Write(nil)
	var request RegisterRequest

	json.Unmarshal(body, &request)

	if request.Token != transferToken || registrationCompletePrinted {
		return
	}

	fmt.Println(strings.Repeat("\n", 100))
	fmt.Print("Setup Complete!\nYou can now copy text from your phone to this device.\n\nClose Application: Ctrl + C\n\n")

	registrationCompletePrinted = true
}

type DataRequest struct {
	Token     string `json:"token"`
	Clipboard string `json:"clipboard"`
}

func handleData(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error reading data")
	}

	w.Write(nil)
	var request DataRequest

	json.Unmarshal(body, &request)

	if request.Token != transferToken {
		return
	}

	clipboard.WriteAll(request.Clipboard)

	fmt.Println("Clipboard received")
}

func printTransferTokenCode() {
	if transferTokenPrinted {
		return
	}

	transferTokenCode, err := qrcode.New(transferToken, qrcode.High)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("\n", 100))
	fmt.Println(transferTokenCode.ToString(true))
	fmt.Println("Code 2")

	transferTokenPrinted = true
}

func randomToken() string {
	alphabet := ([]rune)("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-")
	alphabetLen := len(alphabet)
	upperBound := big.NewInt(int64(alphabetLen))

	token := make([]rune, 42, 42)

	for i := range token {
		bigValue, err := rand.Int(rand.Reader, upperBound)
		if err != nil {
			log.Fatal("Error during startup")
		}

		token[i] = alphabet[bigValue.Int64()]
	}

	return string(token)
}
