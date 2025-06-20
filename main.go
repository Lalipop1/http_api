package main

import (
	"internal/handlers"
	"internal/services"
	"internal/storage"
	"log"
	"net/http"
)

func main() {
	// Инициализация хранилища в памяти
	taskStorage := storage.NewInMemoryTaskStorage()

	// Сервис для работы с задачами
	taskService := services.NewTaskService(taskStorage)

	// HTTP обработчики
	taskHandler := handlers.NewTaskHandler(taskService)

	// Настройка маршрутов
	http.HandleFunc("/tasks", taskHandler.HandleTasks)
	http.HandleFunc("/tasks/", taskHandler.HandleTaskByID)

	// Запуск сервера
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
