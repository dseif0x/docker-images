package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"acpups-mqtt/ups"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

type Config struct {
	ACPHost      string
	MQTTBroker   string
	MQTTTopic    string
	MQTTClientID string
	MQTTUser     string
	MQTTPassword string
	Interval     time.Duration
}

func loadConfig() *Config {
	config := &Config{
		ACPHost:      getEnv("ACPHOST", "10.13.1.187:3551"),
		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTTopic:    getEnv("MQTT_TOPIC", "ups/status"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "acpups-client"),
		MQTTUser:     getEnv("MQTT_USER", ""),
		MQTTPassword: getEnv("MQTT_PASSWORD", ""),
		Interval:     time.Duration(getEnvInt("POLL_INTERVAL", 30)) * time.Second,
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

func createMQTTClient(config *Config) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTTBroker)
	opts.SetClientID(config.MQTTClientID)

	// Set username and password if provided
	if config.MQTTUser != "" {
		opts.SetUsername(config.MQTTUser)
	}
	if config.MQTTPassword != "" {
		opts.SetPassword(config.MQTTPassword)
	}

	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	client := mqtt.NewClient(opts)
	return client
}

func publishUPSData(client mqtt.Client, topic string, data *ups.Data) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal UPS data: %v", err)
	}

	token := client.Publish(topic, 0, false, jsonData)
	token.Wait()
	return token.Error()
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
		config.ACPHost, config.MQTTBroker, config.MQTTTopic, config.Interval)

	// Create MQTT client
	mqttClient := createMQTTClient(config)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
	defer mqttClient.Disconnect(250)

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
		if err := publishUPSData(mqttClient, config.MQTTTopic, upsData); err != nil {
			log.Printf("Error publishing to MQTT: %v", err)
		} else {
			log.Printf("Published UPS data: Battery=%g%%, Load=%g%%, Status=%s",
				upsData.BatteryLevel, upsData.Load, upsData.Status)
		}

		// Wait for next tick
		<-ticker.C
	}
}
