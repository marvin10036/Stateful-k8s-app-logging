package main

import (
	"fmt"
	"bufio"
	"os"
	"net"
	"net/http"
	"encoding/json"
	"log"
)

type STATUS struct {
  Status string `json:"status"`
}

func recoverApi() {
  http.HandleFunc("/recover", recoverLogs)
  log.Fatal(http.ListenAndServe(":3000", nil))
}

func recoverLogs(resWriter http.ResponseWriter, request *http.Request) {
	// Awnsering request
	resWriter.Header().Set("Content=Type", "application/json")
	json.NewEncoder(resWriter).Encode([]STATUS{{ Status: "OK", }})

	fd, err := os.OpenFile("log.txt", os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer fd.Close()

	serverConn, err := net.Dial("tcp", "localhost:6378")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer serverConn.Close()

	// bufio.Scanner to read line by line
	scanner := bufio.NewScanner(fd)
	// Setting the split function as ScanLines, essentially setting the scan mode
	scanner.Split(bufio.ScanLines)

	/*
	go func() {
		buf := make([]byte, 65535)
		// Lê uma vez e fica bloqueado no read
		serverConn.Read(buf)
		fmt.Println("asd")
	}()
	*/

	for scanner.Scan() {
		currentLine := []byte(scanner.Text())
		currentLine = currentLine[:len(currentLine)-1]

		fmt.Println(currentLine)
		_, err = serverConn.Write(currentLine)
		if err != nil {
			fmt.Println("Erro ao escrever no server")
			return
		}
	}
}
