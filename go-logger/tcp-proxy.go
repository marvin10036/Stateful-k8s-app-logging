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

	// DEBUGGING
	fdD, err := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer fdD.Close()

	writer := bufio.NewWriter(fd)

	// Criando canais
	// clientChan := make(chan []byte)
	// appChan := make(chan []byte)
	done := make(chan bool)

	// Vai ficar escutando a clientConn e mandando os dados para clientChan
	go func() {
		for {
			fdD.Write([]byte("Debug primário\n"))
			n, err := clientConn.Read(requestBuf)
			if err != nil {
				if err.Error() == "EOF" {
					done <- true // Client connection closed
				}
				// writer.Write([]byte("goroutine clientConn finalizada"))
				// writer.Flush()
				fdD.Write([]byte("goroutine clientConn finalizada\n"))
				return
			}
			// clientChan <- requestBuf[:n]
			fdD.Write([]byte("Debug secundário\n"))
			writer.Write(requestBuf[:n])
			writer.Write([]byte("\n"))
			writer.Flush()

			fdD.Write([]byte("Debug terciário\n"))
			// Forwarding da mensagem para o server
			_, err = appConn.Write(requestBuf[:n])
			if err != nil {
				fmt.Println("Error forwarding to app:", err)
				fdD.Write([]byte("Error forwarding to app\n"))
				return
			}
			fdD.Write([]byte("Debug quarternário\n"))
		}
	}()

	// Vai ficar escutando a appConn e mandando os dados para appChan
	go func() {
		for {
			// fdD.Write([]byte("Debug primário\n"))
			n, err := appConn.Read(responseBuf)
			if err != nil {
				if err.Error() == "EOF" {
					done <- true // App server connection closed
				}
				fdD.Write([]byte("goroutine appConn finalizada\n"))
				// writer.Write([]byte("goroutine appConn finalizada"))
				// writer.Flush()
				return
			}
			// appChan <- responseBuf[:n]
			// fdD.Write([]byte("Debug secundário\n"))
			_, err = clientConn.Write(responseBuf[:n])
			if err != nil {
				fmt.Println("Error writing to client:", err)
				fdD.Write([]byte("Error writing to client\n"))
				return
			}
			// fdD.Write([]byte("Debug terciário\n")
			// FIca esperando termino de conexão no primario
		}
	}()

	for {
	}
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
	appConn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to app:", err)
		return
	}
	defer appConn.Close()

	// Handle the communication in a non-blocking manner
	logAndForwardCommunication(clientConn, appConn, fd)
}

// Espera em loop por novas conexões
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
