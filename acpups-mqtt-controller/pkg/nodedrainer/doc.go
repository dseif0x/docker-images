// Package nodedrainer provides functionality to drain Kubernetes worker nodes
// based on UPS status received via MQTT messages.
//
// The package monitors UPS status and automatically:
//   - Drains worker nodes when UPS is on battery and battery level < 50%
//   - Uncordons nodes when power is restored
//   - Preserves control plane nodes (never drains them)
//   - Handles graceful pod eviction respecting DaemonSets and system pods
//
// Example usage:
//
//	// Create Kubernetes and MQTT clients
//	k8sClient, _ := kubernetes.NewForConfig(config)
//	mqttClient := mqtt.NewClient(opts)
//
//	// Create drainer and subscribe to UPS status
//	drainer := nodedrainer.New(k8sClient, mqttClient)
//	err := drainer.Subscribe(nodedrainer.DefaultConfig())
package nodedrainer
