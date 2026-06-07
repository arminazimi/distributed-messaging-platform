package config

import (
	"gateway/pkg/env"
	"strconv"
)

var (
	AppName         string
	AppBaseURL      string
	AppListenAddr   string
	DBUsername      string
	DBPassword      string
	DBHost          string
	DBPort          int
	DBName          string
	RabbitmqUri     string
	MessageExchange string
	ExpressQueue    string
	NormalQueue     string

	// Capacity knobs
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeSec int

	// Rate limiting knobs
	GlobalRateLimitRPS   int
	GlobalRateLimitBurst int
	UserRateLimitRPS     int
	UserRateLimitBurst   int
)

func Init() {
	AppName = env.Default("APP_NAME", "gateway")
	AppListenAddr = env.RequiredNotEmpty("LISTEN_ADDR")
	DBUsername = env.RequiredNotEmpty("DB_USER_NAME")
	DBPassword = env.RequiredNotEmpty("DB_PASSWORD")
	DBHost = env.RequiredNotEmpty("DB_HOST")
	port, err := strconv.Atoi(env.RequiredNotEmpty("DB_PORT"))
	if err != nil {
		panic("invalid DB_PORT: " + err.Error())
	}
	DBPort = port
	DBName = env.RequiredNotEmpty("DB_NAME")
	RabbitmqUri = env.RequiredNotEmpty("RABBIT_URI")
	MessageExchange = env.RequiredNotEmpty("RABBIT_MESSAGE_EXCHANGE")
	ExpressQueue = env.RequiredNotEmpty("EXPRESS_QUEUE")
	NormalQueue = env.RequiredNotEmpty("NORMAL_QUEUE")

	DBMaxOpenConns = env.DefaultInt("DB_MAX_OPEN_CONNS", 50)
	DBMaxIdleConns = env.DefaultInt("DB_MAX_IDLE_CONNS", 25)
	DBConnMaxLifetimeSec = env.DefaultInt("DB_CONN_MAX_LIFETIME_SEC", 300)

	GlobalRateLimitRPS = env.DefaultInt("GLOBAL_RATE_LIMIT_RPS", 1000)
	GlobalRateLimitBurst = env.DefaultInt("GLOBAL_RATE_LIMIT_BURST", 1000)
	UserRateLimitRPS = env.DefaultInt("USER_RATE_LIMIT_RPS", 20)
	UserRateLimitBurst = env.DefaultInt("USER_RATE_LIMIT_BURST", 20)
}
