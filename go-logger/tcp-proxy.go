package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
	"sync"
)

func checkFileNotExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}

func checkIsNetErrorTimeout(err error) bool {
	// Checking if it's a network error. err.(net.Error) is type assertion.  
	if networkError, isNetError := err.(net.Error); isNetError {
		// If it's specifically a Timeout error
		if networkError.Timeout() {
			return true
		}
	}
	return false
}

func clientSideListener(clientConn net.Conn, serverConn net.Conn, fdD *os.File, done chan bool, fd *os.File, wg *sync.WaitGroup) {
	// Signaling to the caller that it has finished once it returns
	defer wg.Done()

	// Buffer to save the message
	requestBuf := make([]byte, 65535)
	// io.Writer to interact with the log file
	writer := bufio.NewWriter(fd)

	for {
		select {
		// Exit if the server has signalized the end of the communication
		case <-done:
			fdD.Write([]byte("Server has set done to true, client is returning\n"))
			return
		default:
			// Sets a timeout so that it doesn't block on reads
			serverConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			n, err := clientConn.Read(requestBuf)
			if err != nil {
				// Checking if its just a timeout error from the deadline
				if checkIsNetErrorTimeout(err) {
					continue
				} else {
					done <- true  // Signaling the connection's end to the client routine
					return
				}
			}

			// Logs the message
			writer.Write(requestBuf[:n])
			writer.Write([]byte("\n"))
			writer.Flush()

			// Forwards message to the server
			_, err = serverConn.Write(requestBuf[:n])
			if err != nil {
				fmt.Println("Error forwarding to server:", err)
				fdD.Write([]byte("Error forwarding to server\n"))
				done <- true  // Signaling the connection's end to the server routine
				return
			}
		}
	}
}

func serverSideListener(clientConn net.Conn, serverConn net.Conn, fdD *os.File, done chan bool, wg *sync.WaitGroup) {
	// Signaling to the caller that it has finished once it returns
	defer wg.Done()

	// Buffer to save the message
	responseBuf := make([]byte, 65535)

	for {
		select {
		// Exit if the client has signalized the end of the communication
		case <-done:
			fdD.Write([]byte("Client has set done to true, server is returning\n"))
			return
		default:
			// Sets a timeout so that it doesn't block on reads
			serverConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			// Reads the serverConn message
			n, err := serverConn.Read(responseBuf)
			if err != nil {
				// Checking if its just a timeout error from the deadline
				if checkIsNetErrorTimeout(err) {
					continue
				} else {
					done <- true  // Signaling the connection's end to the client routine
					return
				}
			}

			// Forwards message to the server
			_, err = clientConn.Write(responseBuf[:n])
			if err != nil {
				fmt.Println("Error writing to client:", err)
				fdD.Write([]byte("Error writing to client\n"))
				done <- true  // Signaling the connection's end to the client routine
				return
			}
		}
	}
}

// Read and forward data from the client to the server, and log communication
func logAndForwardCommunication(clientConn net.Conn, serverConn net.Conn, fd *os.File) {
	// DEBUGGING
	fdD, err := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer fdD.Close()

	/*
	Creating channel so that clientConn may sinalize serverConn about
	the end of the communication, or vice-versa
	*/
	done := make(chan bool)

	/*
	Creating sync.WaitGroup so that this function waits for both of the
	go routines to finish
	*/
	var wg sync.WaitGroup
	wg.Add(2)

	// Listen to client, logs the message and forwards it to the server
	go clientSideListener(clientConn, serverConn, fdD, done, fd, &wg)
	// Listen to the server and forwards messages to the client
	go serverSideListener(clientConn, serverConn, fdD, done, &wg)

	// Waits for the two go routines to finish
	fdD.Write([]byte("Esperando as duas go routines finalizarem\n"))
	wg.Wait()
	fdD.Write([]byte("As duas go routines finalizaram\n"))
}

// Handles incoming client connections and forwards data to the application server
func handleIncomingConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Abre o log
	fd, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer fd.Close()

	// Cria a conexão com o redis-server
	serverConn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer serverConn.Close()

	logAndForwardCommunication(clientConn, serverConn, fd)
}

// Waits for new connections in a loop
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

func main() {
	// Cria o log se ele não existe
	if checkFileNotExists("log.txt") {
		fd, err := os.Create("log.txt")
		if err != nil {
			fmt.Println("Error creating log file:", err)
			return
		}
		fd.Close()
	}

	// DEBUGGING
	if checkFileNotExists("debug.txt") {
		fdD, err := os.Create("debug.txt")
		if err != nil {
			fmt.Println("Error creating debug file:", err)
			return
		}
		fdD.Close()
	}

	// Começa a escutar por conexões
	listenIncomingConnections()
}
