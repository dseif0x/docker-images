package nodedrainer

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/client-go/kubernetes"
)

// UPSStatus represents the structure of UPS status messages from MQTT
type UPSStatus struct {
	Timestamp    string `json:"timestamp"`
	BatteryLevel int    `json:"battery_level"`
	InputVoltage int    `json:"input_voltage"`
	Load         int    `json:"load"`
	Status       string `json:"status"`
}

// NodeDrainer manages Kubernetes node draining based on UPS status
type NodeDrainer struct {
	clientset  *kubernetes.Clientset
	mutex      sync.RWMutex
	mqttClient mqtt.Client
	lastStatus *UPSStatus
	config     *Config
}

// Config holds configuration options for the NodeDrainer
type Config struct {
	MQTTTopic             string
	QoS                   byte
	BatteryDrainThreshold int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MQTTTopic:             "ups/status",
		QoS:                   0,
		BatteryDrainThreshold: 50,
	}
}
