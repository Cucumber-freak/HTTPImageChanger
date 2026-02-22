package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL      string
	S3Endpoint string
	S3Key      string
	S3Secret   string
	S3Bucket   string
	RabbitURL  string
	QueueName  string
	Port       string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DBURL:      getEnv("DB_URL", "postgres://user:password@localhost:5432/db"),
		S3Endpoint: getEnv("S3_ENDPOINT", "localhost:9000"),
		S3Key:      getEnv("S3_ACCESS_KEY", "admin"),
		S3Secret:   getEnv("S3_SECRET_KEY", "admin"),
		S3Bucket:   getEnv("S3_BUCKET", "images"),
		RabbitURL:  getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:  getEnv("RABBIT_QUEUE", "task_queue"),
		Port:       getEnv("SERVER_PORT", ":8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
