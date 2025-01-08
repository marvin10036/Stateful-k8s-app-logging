package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	// "time"
)

func checkFileNotExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}

// Read and forward data from the client to the server, and log communication
func logAndForwardCommunication(clientConn net.Conn, appConn net.Conn, fd *os.File) {
	requestBuf := make([]byte, 65535)
	responseBuf := make([]byte, 65535)

	writer := bufio.NewWriter(fd)

	// Criando canais
	clientChan := make(chan []byte)
	appChan := make(chan []byte)
	done := make(chan bool)

	// Ler os dados do cliente e coloca no canal
	go func() {
		for {
			n, err := clientConn.Read(requestBuf)
			if err != nil {
				if err.Error() == "EOF" {
					done <- true // Client connection closed
				}
				writer.Write([]byte("goroutine clientConn finalizada"))
				writer.Flush()
				return
			}
			clientChan <- requestBuf[:n]
		}
	}()

	// Ler os dados do servidor e coloca no canal
	go func() {
		for {
			n, err := appConn.Read(responseBuf)
			if err != nil {
				if err.Error() == "EOF" {
					done <- true // App server connection closed
				}
				writer.Write([]byte("goroutine appConn finalizada"))
				writer.Flush()
				return
			}
			appChan <- responseBuf[:n]
		}
	}()

	// Main loop to handle communication
	for {
		select {
		case reqData := <-clientChan:
			// Log the request data
			writer.Write(reqData)
			writer.Write([]byte("\n"))
			writer.Flush()

			// Forward request data to the application server
			_, err := appConn.Write(reqData)
			if err != nil {
				fmt.Println("Error forwarding to app:", err)
				return
			}

		case resData := <-appChan:
			// Forward the response data to the client
			_, err := clientConn.Write(resData)
			if err != nil {
				fmt.Println("Error writing to client:", err)
				return
			}

		case <-done:
			fmt.Println("Connection closed, shutting down proxy")
			return
		default:
			continue
		}
	}
}

// Handles incoming client connections and forwards data to the application server
func handleIncomingConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Opens the log file
	fd, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer fd.Close()

	// Conects to redis-server
	appConn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to redis:", err)
		return
	}
	defer appConn.Close()

	/*
	// Opens the log file
	fd, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer fd.Close()
	*/

	// Handle the communication in a non-blocking manner
	logAndForwardCommunication(clientConn, appConn, fd)
}

// Listen for incoming connections and spawn handlers for each
func listenIncomingConnections() {
	// Listen on port 6380 for incoming connections
	ln, err := net.Listen("tcp", ":6380")
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return
	}
	defer ln.Close()

	for {
		// Accept incoming connections
		clientConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each connection in a new goroutine
		go handleIncomingConnection(clientConn)
	}
}

// Main entry point for the proxy server
func main() {
	// Create the log file if it doesn't exist
	if checkFileNotExists("log.txt") {
		fd, err := os.Create("log.txt")
		if err != nil {
			fmt.Println("Error creating log file:", err)
			return
		}
		fd.Close()
	}

	// Start listening for incoming connections
	listenIncomingConnections()
}
