package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	ServerURL = "http://localhost:8080/cotacao"
	TimeOut   = 300 * time.Millisecond
	Filename  = "cotacao.txt"
)

func main() {
	client := http.Client{}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, TimeOut)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ServerURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("Valor: %s\n", string(body))))
}
