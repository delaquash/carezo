package configs


import (
	"logs"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost 		string
	DBPort 		string
	DBUser 		string
	DBPassword 	string
	DBName		string

	// Redis Setting
}