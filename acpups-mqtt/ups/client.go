package ups

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	Timestamp    time.Time `json:"timestamp"`
	BatteryLevel float64   `json:"battery_level"`
	InputVoltage float64   `json:"input_voltage"`
	Load         float64   `json:"load"`
	Status       string    `json:"status"`
}

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{host: host}
}

func (c *Client) Connect() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", c.host, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to UPS: %v", err)
	}
	return conn, nil
}

func (c *Client) FetchData(conn net.Conn) (*Data, error) {
	// Set read/write timeouts
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Send "status" command using NIS protocol
	if err := writeNISMessage(conn, "status"); err != nil {
		return nil, fmt.Errorf("failed to send status command: %v", err)
	}

	// Read all response messages until EOF
	var responseBuilder strings.Builder
	for {
		message, err := readNISMessage(conn)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}
		responseBuilder.WriteString(message)
	}

	response := responseBuilder.String()
	return parseResponse(response)
}

// writeNISMessage writes a message using the apcupsd NIS protocol format
func writeNISMessage(conn net.Conn, message string) error {
	data := []byte(message)
	length := uint16(len(data))

	// Write 2-byte length prefix (big-endian)
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Write message data
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %v", err)
	}

	return nil
}

// readNISMessage reads a message using the apcupsd NIS protocol format
func readNISMessage(conn net.Conn) (string, error) {
	// Read 2-byte length prefix
	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		if err == io.EOF {
			return "", err
		}
		return "", fmt.Errorf("failed to read message length: %v", err)
	}

	// If length is 0, return EOF
	if length == 0 {
		return "", io.EOF
	}

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return "", fmt.Errorf("failed to read message data: %v", err)
	}

	return string(data), nil
}

func parseResponse(response string) (*Data, error) {
	data := &Data{
		Timestamp: time.Now(),
	}

	scanner := bufio.NewScanner(strings.NewReader(response))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse key-value pairs in format "key : value"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Handle different apcupsd status fields
		switch strings.ToUpper(key) {
		case "BCHARGE":
			// Battery charge percentage
			if strings.HasSuffix(value, " Percent") {
				value = strings.TrimSuffix(value, " Percent")
			}
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				data.BatteryLevel = val
			}
		case "LINEV":
			// Input line voltage
			if strings.HasSuffix(value, " Volts") {
				value = strings.TrimSuffix(value, " Volts")
			}
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				data.InputVoltage = val
			}
		case "LOADPCT":
			// Load percentage
			if strings.HasSuffix(value, " Percent") {
				value = strings.TrimSuffix(value, " Percent")
			}
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				data.Load = val
			}
		case "STATUS":
			// UPS status string
			data.Status = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing UPS response: %v", err)
	}

	return data, nil
}
