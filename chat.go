package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Константы для настроек
const (
	//telnet localhost 8080
	serverAddress = ":8080" // Адрес и порт сервера
	exitCommand   = "Exit"  // Команда для выхода
)

// Глобальные переменные для хранения подключений
var (
	clients    = make(map[net.Conn]string) // Список клиентов
	clientsMux sync.Mutex                  // Мьютекс для синхронизации доступа к списку клиентов
)

// broadcastMessage рассылает сообщение всем клиентам
func broadcastMessage(sender net.Conn, message string) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		// Не отправляем сообщение самому себе
		if client != sender {
			_, err := client.Write([]byte(message + "\n\r"))
			if err != nil {
				log.Printf("Error sending message to client %s: %v", clients[client], err)
			}
		}
	}
}

// handleConnection обрабатывает соединение с клиентом
func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	// Добавляем клиента в список
	clientsMux.Lock()
	clients[conn] = clientAddr
	clientsMux.Unlock()

	log.Printf("Client %s connected", clientAddr)

	// Приветствуем клиента
	_, err := conn.Write([]byte("Welcome to the chat, " + clientAddr + "!\n\r"))
	if err != nil {
		log.Printf("Error writing to client %s: %v", clientAddr, err)
		return
	}

	// Уведомляем остальных о подключении нового клиента
	broadcastMessage(conn, fmt.Sprintf("%s has joined the chat", clientAddr))

	// Читаем сообщения от клиента
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()

		// Обработка команды выхода
		if text == exitCommand {
			_, _ = conn.Write([]byte("Bye!\n\r"))
			log.Printf("Client %s disconnected", clientAddr)
			break
		}

		// Рассылаем сообщение всем клиентам
		message := fmt.Sprintf("[%s]: %s", clientAddr, text)
		log.Println(message) // Логируем сообщение
		broadcastMessage(conn, message)
	}

	// Проверяем ошибки сканирования
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from client %s: %v", clientAddr, err)
	}

	// Удаляем клиента из списка
	clientsMux.Lock()
	delete(clients, conn)
	clientsMux.Unlock()

	// Уведомляем остальных о выходе клиента
	broadcastMessage(conn, fmt.Sprintf("%s has left the chat", clientAddr))
}

// startServer запускает сервер
func startServer() {
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server is running on %s", serverAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Обрабатываем каждого клиента в отдельной горутине
		go handleConnection(conn)
	}
}

// waitForShutdown ожидает сигнал завершения работы
func waitForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan // Ожидаем сигнал
	log.Println("Shutting down server...")

	// Закрываем все активные соединения
	clientsMux.Lock()
	for conn := range clients {
		conn.Close()
	}
	clientsMux.Unlock()
}

func main() {
	// Запуск сервера
	go startServer()

	// Ожидание завершения
	waitForShutdown()
}
