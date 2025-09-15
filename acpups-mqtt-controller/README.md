# Kubernetes UPS Node Drainer

A Go application that monitors UPS status via MQTT and automatically drains Kubernetes worker nodes during power outages to ensure graceful workload migration.

## Features

- Monitors UPS status via MQTT (topic: `ups/status`)
- Drains worker nodes when UPS is on battery and battery level < 50%
- Automatically uncordons nodes when power is restored
- Preserves control plane nodes (never drains them)
- Respects DaemonSets and system pods
- Handles pod eviction with proper Kubernetes APIs

## Configuration

The application is configured via environment variables:

- `MQTT_BROKER`: MQTT broker URL (default: `tcp://localhost:1883`)
- `MQTT_CLIENT_ID`: MQTT client ID (default: `k8s-ups-drainer`)
- `MQTT_TOPIC`: MQTT topic to subscribe to (default: `ups/status`)
- `MQTT_USER`: MQTT username (optional)
- `MQTT_PASSWORD`: MQTT password (optional)
- `BATTERY_DRAIN_THRESHOLD`: Battery level threshold for draining nodes (default: `50`, range: 1-100)
- `KUBECONFIG`: Path to kubeconfig file (uses in-cluster config when running in K8s)

## Expected UPS Status Format

The application expects MQTT messages on the configured topic (default `ups/status`) with the following JSON format:

```json
{
  "timestamp": "2025-09-15T12:21:09.845003+02:00",
  "battery_level": 99,
  "input_voltage": 230,
  "load": 14,
  "status": "ONLINE"
}
```

## Logic

1. **Power Outage Detection**: When `status` is `"ONBATT"` and `battery_level` < configured threshold (default 50%)
   - Drains all worker nodes (non-control-plane)
   - Cordons nodes to prevent new pod scheduling
   - Evicts pods gracefully (respecting DaemonSets and system pods)
   - Marks nodes with annotation `ups-drainer.k8s.io/drained-by: ups-node-drainer`

2. **Power Restoration**: When `status` is `"ONLINE"`
   - Uncordons all nodes drained by this service (identified by annotation)
   - Removes the drain annotation
   - Allows normal scheduling to resume

3. **Node Classification**: 
   - Control plane nodes are identified by labels or taints:
     - `node-role.kubernetes.io/control-plane`
     - `node-role.kubernetes.io/master`

## Deployment

### 1. Build and Push Docker Image

```bash
docker build -t your-registry/k8s-ups-drainer:latest .
docker push your-registry/k8s-ups-drainer:latest
```

### 2. Update Kubernetes Manifests

Edit `k8s/configmap.yaml` to set your MQTT configuration:

```yaml
data:
  MQTT_BROKER: "tcp://your-mqtt-broker:1883"
  MQTT_TOPIC: "custom/ups/topic"  # Use custom topic instead of default ups/status
  BATTERY_DRAIN_THRESHOLD: "40"  # Drain at 40% instead of default 50%
```

If authentication is required, add credentials to `k8s/secret.yaml`:

```bash
# Base64 encode your credentials
echo -n "your-username" | base64
echo -n "your-password" | base64
```

### 3. Deploy to Kubernetes

```bash
# Apply RBAC permissions
kubectl apply -f k8s/rbac.yaml

# Apply configuration
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# Update deployment image and deploy
kubectl apply -f k8s/deployment.yaml
```

## RBAC Permissions

The application requires the following Kubernetes permissions:

- `nodes`: get, list, update, patch (for cordoning/uncordoning)
- `pods`: get, list (for finding pods to evict)  
- `pods/eviction`: create (for graceful pod eviction)
- `poddisruptionbudgets`: get, list (for respecting PDBs)

## Security Considerations

- Runs as non-root user
- Read-only root filesystem  
- Drops all capabilities
- Minimal resource requests/limits
- Only accesses necessary Kubernetes resources

## Testing

You can test the application by publishing test messages to your MQTT broker:

```bash
# Simulate power outage with low battery (using default topic)
mosquitto_pub -h your-broker -t ups/status -m '{"timestamp":"2025-09-15T12:21:09.845003+02:00","battery_level":40,"input_voltage":0,"load":14,"status":"ONBATT"}'

# Simulate power restoration (using default topic)
mosquitto_pub -h your-broker -t ups/status -m '{"timestamp":"2025-09-15T12:21:09.845003+02:00","battery_level":99,"input_voltage":230,"load":14,"status":"ONLINE"}'

# If using custom topic, replace ups/status with your configured topic
mosquitto_pub -h your-broker -t custom/ups/topic -m '{"timestamp":"2025-09-15T12:21:09.845003+02:00","battery_level":40,"input_voltage":0,"load":14,"status":"ONBATT"}'
```

## Monitoring

The application logs its actions to stdout:

- MQTT connection status
- UPS status updates
- Node draining/uncordoning actions
- Pod eviction activities
- Error conditions