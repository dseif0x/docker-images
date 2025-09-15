package mqtt

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	Broker   string
	Topic    string
	ClientID string
	User     string
	Password string
}

type Client struct {
	client mqtt.Client
	topic  string
}

func NewClient(config *Config) *Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetClientID(config.ClientID)

	// Set username and password if provided
	if config.User != "" {
		opts.SetUsername(config.User)
	}
	if config.Password != "" {
		opts.SetPassword(config.Password)
	}

	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	mqttClient := mqtt.NewClient(opts)

	return &Client{
		client: mqttClient,
		topic:  config.Topic,
	}
}

func (c *Client) Connect() error {
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}
	return nil
}

func (c *Client) Disconnect() {
	c.client.Disconnect(250)
}

func (c *Client) PublishJSON(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	token := c.client.Publish(c.topic, 0, false, jsonData)
	token.Wait()
	return token.Error()
}
