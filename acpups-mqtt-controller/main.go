package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s-ups-drainer/pkg/nodedrainer"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	log.Println("Starting Kubernetes UPS Node Drainer")

	// Initialize Kubernetes client
	k8sClient, err := initKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}
	log.Println("Initializing Kubernetes Controllers")

	// Initialize MQTT client
	mqttClient, err := initMQTTClient()
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}
	log.Println("Connected to MQTT broker")

	// Create node drainer
	drainer := nodedrainer.New(k8sClient, mqttClient)

	// Subscribe to UPS status with configuration
	config := createConfig()
	if err := drainer.Subscribe(config); err != nil {
		log.Fatalf("Failed to subscribe to MQTT: %v", err)
	}
	log.Println("Subscribed to UPS status updates")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	mqttClient.Disconnect(250)
}

func initKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}

func initMQTTClient() (mqtt.Client, error) {
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "tcp://localhost:1883"
	}

	clientID := os.Getenv("MQTT_CLIENT_ID")
	if clientID == "" {
		clientID = "k8s-ups-drainer"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")
	if username != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", broker)
	return client, nil
}

func createConfig() *nodedrainer.Config {
	config := nodedrainer.DefaultConfig()

	// Read MQTT topic from environment
	if topic := os.Getenv("MQTT_TOPIC"); topic != "" {
		config.MQTTTopic = topic
		log.Printf("Using MQTT topic from env: %s", topic)
	}

	// Read battery drain threshold from environment
	if thresholdStr := os.Getenv("BATTERY_DRAIN_THRESHOLD"); thresholdStr != "" {
		if threshold, err := strconv.Atoi(thresholdStr); err == nil && threshold > 0 && threshold <= 100 {
			config.BatteryDrainThreshold = threshold
			log.Printf("Using battery drain threshold from env: %d%%", threshold)
		} else {
			log.Printf("Invalid BATTERY_DRAIN_THRESHOLD value '%s', using default: %d%%", thresholdStr, config.BatteryDrainThreshold)
		}
	}

	return config
}
