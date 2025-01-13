package main

import (
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
	// DEBUGGING
	/*
	fdD, errD := os.Create("apiDebug.txt")
	if errD != nil {
		return
	}
	defer fdD.Close()
	*/

	// Awnsering request
	resWriter.Header().Set("Content=Type", "application/json")
	json.NewEncoder(resWriter).Encode([]STATUS{{ Status: "OK", }})

	fd, err := os.OpenFile("log.txt", os.O_RDONLY, 0644)
	if err != nil {
		// Error opening log file
		return
	}
	defer fd.Close()

	serverConn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		// Error connecting to the server
		return
	}
	defer serverConn.Close()

	// bufio.Scanner to read line by line
	scanner := bufio.NewScanner(fd)
	// Setting the split function as ScanLines, essentially setting the scan mode
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		currentLine := []byte(scanner.Text())
		// TODO Remover coupling to redis protocol later
		CR := []byte("\r\n")

		_, err = serverConn.Write(append(currentLine, CR...))
		if err != nil {
			// Error writing to the server
			return
		}
	}
}
