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

var DB *sql.DB

func Start(registerRoutes func(), addr string) {
	loadEnv(".env")

	env := getRequiredEnvVars([]string{
		"DEV_USER",
		"DEV_PASS",
		"DATABASE_NAME",
		"DATABASE_USER",
		"DATABASE_PASSWORD",
		"DATABASE_HOST",
		"DATABASE_PORT",
	})

	dsn := buildDSN(env)
	DB = initDB(dsn)

	middleware.InitMiddleware(env["DEV_USER"], env["DEV_PASS"])
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
		env["DATABASE_USER"],
		env["DATABASE_PASSWORD"],
		env["DATABASE_HOST"],
		env["DATABASE_PORT"],
		env["DATABASE_NAME"],
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
