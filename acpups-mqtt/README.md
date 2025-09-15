# ACP UPS MQTT Bridge

A Go application that fetches data from an ACP UPS device and publishes events via MQTT.

## Features

- Connects to apcupsd daemon via TCP using the correct NIS (Network Information Server) protocol
- Implements proper apcupsd network protocol with 2-byte length prefixes
- Fetches UPS status data including battery level, voltage and load
- Publishes data to MQTT broker in JSON format with authentication support
- Configurable via environment variables or .env file
- Auto-reconnect functionality for both UPS and MQTT connections
- Proper parsing of apcupsd response format with units handling

## Configuration

Set the following environment variables:

- `ACPHOST` - apcupsd daemon host and port (default: `10.13.1.187:3551`)
- `MQTT_BROKER` - MQTT broker URL (default: `tcp://localhost:1883`)
- `MQTT_TOPIC` - MQTT topic to publish to (default: `ups/status`)
- `MQTT_USER` - MQTT username for authentication (optional)
- `MQTT_PASSWORD` - MQTT password for authentication (optional)
- `MQTT_CLIENT_ID` - MQTT client ID (default: `acpups-client`)
- `POLL_INTERVAL` - Polling interval in seconds (default: `30`)

## Usage

```bash
# Set environment variables
export ACPHOST=10.13.1.187:3551
export MQTT_BROKER=tcp://mqtt.example.com:1883
export MQTT_TOPIC=home/ups/status

# Run the application
./acpups-mqtt
```

## Data Format

The application publishes UPS data in JSON format:

```json
{
  "timestamp": "2025-09-15T11:30:00Z",
  "battery_level": 100.0,
  "input_voltage": 120.0,
  "load": 25.5,
  "status": "ONLINE"
}
```

## Building

```bash
go build -o acpups-mqtt
```

## Docker

You can also run this in a Docker container by creating a Dockerfile if needed.