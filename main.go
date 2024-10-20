package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
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

// TODO: cli arguments, build

func main() {
	domain := "localhost"
	pageLink := fmt.Sprintf("https://%s", domain)

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

func handleLib(w http.ResponseWriter, req *http.Request) {
	file, err := os.ReadFile("./html/lib/qr-scanner.umd.min.js")
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Add("Content-Type", "application/javascript")
	w.Write(file)
}

func handleLibWorker(w http.ResponseWriter, req *http.Request) {
	file, err := os.ReadFile("./html/lib/qr-scanner-worker.min.js")
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Add("Content-Type", "application/javascript")
	w.Write(file)
}

func handleDashboard(w http.ResponseWriter, req *http.Request) {
	file, err := os.ReadFile("./html/dashboard.html")
	if err != nil {
		log.Fatal(err)
	}

	hydrated := strings.ReplaceAll(string(file), "{{START_TIME}}", startTime.Format(time.RFC3339))

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
