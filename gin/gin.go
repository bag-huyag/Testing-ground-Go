package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// User структура для хранения данных пользователя
type User struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// База данных в памяти (потокобезопасная)
var (
	DB    = make(map[string]int)
	DBMux sync.Mutex
)

// Инициализация конфигурации
func initConfig() {
	viper.SetConfigName("config") // Имя файла конфигурации (config.yaml)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Путь к файлу конфигурации

	// Установка значений по умолчанию
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("auth.username", "admin")
	viper.SetDefault("auth.password", "password")

	// Чтение конфигурации
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Не удалось прочитать файл конфигурации: %v", err)
	}
}

// Инициализация маршрутизатора
func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Маршруты
	r.GET("/ping", pingHandler)
	r.GET("/users/:name", getUserHandler)

	// Группа маршрутов с авторизацией
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		viper.GetString("auth.username"): viper.GetString("auth.password"),
	}))
	authorized.POST("/users", updateUserHandler)

	return r
}

// Handler для /ping
func pingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// Handler для GET /users/:name
func getUserHandler(c *gin.Context) {
	name := c.Param("name")

	DBMux.Lock()
	value, exists := DB[name]
	DBMux.Unlock()

	if exists {
		c.JSON(http.StatusOK, gin.H{"user": name, "value": value})
	} else {
		c.JSON(http.StatusOK, gin.H{"user": name, "status": "no value"})
	}
}

// Handler для POST /users (требует авторизации)
func updateUserHandler(c *gin.Context) {
	// Получение имени авторизованного пользователя
	user := c.MustGet(gin.AuthUserKey).(string)

	// Парсинг JSON
	var params struct {
		Value int `json:"value" binding:"required,min=1,max=1000"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "validation_error", "error": err.Error()})
		return
	}

	// Обновление значения в базе данных
	DBMux.Lock()
	DB[user] = params.Value
	DBMux.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Основная функция
func main() {
	// Инициализация конфигурации
	initConfig()

	// Настройка маршрутизатора
	router := setupRouter()

	// Запуск сервера
	port := viper.GetString("server.port")
	log.Printf("Сервер запущен на порту %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
