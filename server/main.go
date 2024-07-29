package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const (
	ApiURL          = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	ApiTimeOut      = 200 * time.Millisecond
	DatabaseTimeOut = 10 * time.Millisecond
	DBFile          = "cotacao.sqlite"

	CreateTable   = `CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY AUTOINCREMENT, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)`
	InsertCotacao = `INSERT INTO cotacoes (bid) VALUES (?)`
)

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}
type Cotacao struct {
	Usdbrl `json:"USDBRL"`
}

func main() {

	db, err := sql.Open("sqlite3", DBFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if _, err := db.Exec(CreateTable); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		contextApi, cancel := context.WithTimeout(context.Background(), ApiTimeOut)
		defer cancel()

		request, err := http.NewRequestWithContext(contextApi, http.MethodGet, ApiURL, nil)
		if err != nil {
			log.Fatal("Failed to create response ", err)
			http.Error(w, "Failed to create response ", http.StatusInternalServerError)
			return
		}

		response, err := http.DefaultClient.Do(request)
		defer response.Body.Close()
		if err != nil {
			log.Fatal("Failed to get response ", err)
			http.Error(w, "Failed to get response ", http.StatusInternalServerError)
			return
		}

		var cotacao Cotacao
		if err := json.NewDecoder(response.Body).Decode(&cotacao); err != nil {
			log.Println("Failed to decode response:", err)
			http.Error(w, "Failed to decode response", http.StatusInternalServerError)
			return
		}

		smt, err := db.Prepare(InsertCotacao)
		if err != nil {
			log.Fatal("Failed to create prepare statement ", err)
			http.Error(w, "Failed to create prepare statement", http.StatusInternalServerError)
			return
		}
		defer smt.Close()

		contextDatabase, cancelDB := context.WithTimeout(context.Background(), DatabaseTimeOut)
		defer cancelDB()
		_, err = smt.ExecContext(contextDatabase, cotacao.Bid)
		if err != nil {
			log.Fatal("Failed to save in db ", err)
			http.Error(w, "Failed to save in db", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cotacao.Bid)
	})

	log.Println("Server running on port", 8080)
	http.ListenAndServe(":8080", nil)

}
