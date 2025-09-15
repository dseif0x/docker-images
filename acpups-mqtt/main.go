package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"acpups-mqtt/mqtt"
	"acpups-mqtt/ups"

	"github.com/joho/godotenv"
)

type Config struct {
	ACPHost  string
	MQTT     mqtt.Config
	Interval time.Duration
}

func loadConfig() *Config {
	config := &Config{
		ACPHost: getEnv("ACPHOST", "10.13.1.187:3551"),
		MQTT: mqtt.Config{
			Broker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
			Topic:    getEnv("MQTT_TOPIC", "ups/status"),
			ClientID: getEnv("MQTT_CLIENT_ID", "acpups-client"),
			User:     getEnv("MQTT_USER", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
		},
		Interval: time.Duration(getEnvInt("POLL_INTERVAL", 30)) * time.Second,
	}
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func main() {
	// Load environment variables from .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	log.Println("Starting ACP UPS to MQTT bridge...")

	config := loadConfig()
	log.Printf("Configuration: ACP Host=%s, MQTT Broker=%s, Topic=%s, Interval=%v",
		config.ACPHost, config.MQTT.Broker, config.MQTT.Topic, config.Interval)

	// Create MQTT client
	mqttClient := mqtt.NewClient(&config.MQTT)
	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	// Create UPS client
	upsClient := ups.NewClient(config.ACPHost)

	// Main loop
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for {
		// Connect to UPS
		conn, err := upsClient.Connect()
		if err != nil {
			log.Printf("Error connecting to UPS: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Fetch UPS data
		upsData, err := upsClient.FetchData(conn)
		conn.Close()

		if err != nil {
			log.Printf("Error fetching UPS data: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Publish to MQTT
		if err := mqttClient.PublishJSON(upsData); err != nil {
			log.Printf("Error publishing to MQTT: %v", err)
		} else {
			log.Printf("Published UPS data: Battery=%g%%, Load=%g%%, Status=%s",
				upsData.BatteryLevel, upsData.Load, upsData.Status)
		}

		// Wait for next tick
		<-ticker.C
	}
}
