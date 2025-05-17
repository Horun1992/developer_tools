package server

import (
	"database/sql"
	"developers_tools/internal/handlers"
	"developers_tools/internal/middleware"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

const (
	devUser = "DEV_USER"
	devPass = "DEV_PASS"
	dbName  = "DATABASE_NAME"
	dbUser  = "DATABASE_USER"
	dbPass  = "DATABASE_PASSWORD"
	dbHost  = "DATABASE_HOST"
	dbPort  = "DATABASE_PORT"
)

var DB *sql.DB

func Start(registerRoutes func(), addr string) {
	loadEnv(".env")

	env := getRequiredEnvVars([]string{
		devUser, devPass, dbName, dbUser, dbPass, dbHost, dbPort,
	})

	dsn := buildDSN(env)
	DB = initDB(dsn)

	middleware.InitMiddleware(env[devUser], env[devPass])
	handlers.InitPlateHandler(DB)

	registerRoutes()

	fmt.Printf("🔥 Server started on https://dev.e-mcg.ru")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getRequiredEnvVars(keys []string) map[string]string {
	env := make(map[string]string)
	for _, key := range keys {
		val := os.Getenv(key)
		if val == "" {
			log.Fatalf("❌ Переменная окружения %s не установлена", key)
		}
		env[key] = val
	}
	return env
}

func buildDSN(env map[string]string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env[dbUser],
		env[dbPass],
		env[dbHost],
		env[dbPort],
		env[dbName],
	)
}

func initDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Ошибка при проверке соединения с БД: %v", err)
	}

	return db
}

func loadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		wd, _ := os.Getwd()
		log.Println("📂 Working directory:", wd)
		log.Println("⚠️ Не удалось загрузить .env файл. Используются переменные окружения из системы.")
	}
}
