package nodedrainer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// UPSDrainerAnnotation is used to mark nodes that were drained by this service
	UPSDrainerAnnotation = "ups-drainer.k8s.io/drained-by"
	// UPSDrainerValue is the value set in the annotation
	UPSDrainerValue = "ups-node-drainer"
)

// New creates a new NodeDrainer instance
func New(clientset *kubernetes.Clientset, mqttClient mqtt.Client) *NodeDrainer {
	return &NodeDrainer{
		clientset:  clientset,
		mqttClient: mqttClient,
	}
}

// Subscribe subscribes to the MQTT topic and starts handling UPS status messages
func (nd *NodeDrainer) Subscribe(config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	// Store config in the drainer instance
	nd.config = config

	token := nd.mqttClient.Subscribe(config.MQTTTopic, config.QoS, func(client mqtt.Client, msg mqtt.Message) {
		var status UPSStatus
		if err := json.Unmarshal(msg.Payload(), &status); err != nil {
			log.Printf("Failed to parse UPS status: %v", err)
			return
		}

		log.Printf("Received UPS status: %+v", status)
		nd.lastStatus = &status
		nd.handleUPSStatus(&status)
	})

	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", config.MQTTTopic, token.Error())
	}

	log.Printf("Subscribed to MQTT topic: %s with battery drain threshold: %d%%", config.MQTTTopic, config.BatteryDrainThreshold)
	return nil
}

// GetLastStatus returns the last received UPS status
func (nd *NodeDrainer) GetLastStatus() *UPSStatus {
	nd.mutex.RLock()
	defer nd.mutex.RUnlock()
	return nd.lastStatus
}

// GetDrainedNodes returns the currently drained worker nodes by checking node state
func (nd *NodeDrainer) GetDrainedNodes() ([]string, error) {
	ctx := context.Background()
	nodes, err := nd.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var drainedNodes []string
	for _, node := range nodes.Items {
		// Skip control plane nodes
		if nd.isControlPlaneNode(&node) {
			continue
		}

		// Check if node is cordoned and has our annotation
		if node.Spec.Unschedulable {
			if annotation, exists := node.Annotations[UPSDrainerAnnotation]; exists && annotation == UPSDrainerValue {
				drainedNodes = append(drainedNodes, node.Name)
			}
		}
	}

	return drainedNodes, nil
}

func (nd *NodeDrainer) handleUPSStatus(status *UPSStatus) {
	nd.mutex.Lock()
	defer nd.mutex.Unlock()

	// Use configured battery threshold
	threshold := nd.config.BatteryDrainThreshold

	// Check if we need to drain nodes (ONBATT and battery < threshold)
	if status.Status == "ONBATT" && status.BatteryLevel < threshold {
		log.Printf("UPS on battery with %d%% (threshold: %d%%) - ensuring all worker nodes are drained", status.BatteryLevel, threshold)
		if err := nd.ensureWorkerNodesDrained(); err != nil {
			log.Printf("Failed to drain worker nodes: %v", err)
		}
	} else if status.Status == "ONLINE" {
		// Power is back, uncordon any nodes that were drained by us
		log.Println("Power restored - uncordoning nodes drained by UPS drainer")
		if err := nd.uncordonUPSDrainedNodes(); err != nil {
			log.Printf("Failed to uncordon nodes: %v", err)
		}
	}
}

func (nd *NodeDrainer) ensureWorkerNodesDrained() error {
	ctx := context.Background()

	nodes, err := nd.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes.Items {
		// Skip control plane nodes
		if nd.isControlPlaneNode(&node) {
			log.Printf("Skipping control plane node: %s", node.Name)
			continue
		}

		// Check if node is already cordoned and annotated by us
		isAlreadyDrained := node.Spec.Unschedulable
		hasDrainerAnnotation := false
		if node.Annotations != nil {
			if annotation, exists := node.Annotations[UPSDrainerAnnotation]; exists && annotation == UPSDrainerValue {
				hasDrainerAnnotation = true
			}
		}

		// If already drained by us, skip
		if isAlreadyDrained && hasDrainerAnnotation {
			log.Printf("Node %s already drained by UPS drainer", node.Name)
			continue
		}

		log.Printf("Draining node: %s", node.Name)
		if err := nd.drainNode(&node); err != nil {
			log.Printf("Failed to drain node %s: %v", node.Name, err)
			continue
		}
	}

	return nil
}

func (nd *NodeDrainer) isControlPlaneNode(node *corev1.Node) bool {
	// Check common control plane labels/taints
	labels := node.Labels
	taints := node.Spec.Taints

	// Check for control plane labels
	if labels != nil {
		if _, exists := labels["node-role.kubernetes.io/control-plane"]; exists {
			return true
		}
		if _, exists := labels["node-role.kubernetes.io/master"]; exists {
			return true
		}
	}

	// Check for control plane taints
	for _, taint := range taints {
		if taint.Key == "node-role.kubernetes.io/control-plane" ||
			taint.Key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false
}

func (nd *NodeDrainer) drainNode(node *corev1.Node) error {
	ctx := context.Background()

	// First, cordon the node and add our annotation
	nodeCopy := node.DeepCopy()
	nodeCopy.Spec.Unschedulable = true

	if nodeCopy.Annotations == nil {
		nodeCopy.Annotations = make(map[string]string)
	}
	nodeCopy.Annotations[UPSDrainerAnnotation] = UPSDrainerValue

	_, err := nd.clientset.CoreV1().Nodes().Update(ctx, nodeCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to cordon node %s: %w", node.Name, err)
	}

	log.Printf("Cordoned node: %s", node.Name)

	// Get all pods on the node
	pods, err := nd.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
	})
	if err != nil {
		return fmt.Errorf("failed to list pods on node %s: %w", node.Name, err)
	}

	// Evict pods (excluding DaemonSets and system pods)
	for _, pod := range pods.Items {
		if nd.shouldEvictPod(&pod) {
			if err := nd.evictPod(&pod); err != nil {
				log.Printf("Failed to evict pod %s/%s: %v", pod.Namespace, pod.Name, err)
			}
		}
	}

	return nil
}

func (nd *NodeDrainer) shouldEvictPod(pod *corev1.Pod) bool {
	// Skip system namespaces
	systemNamespaces := []string{"kube-system", "kube-public", "kube-node-lease"}
	for _, ns := range systemNamespaces {
		if pod.Namespace == ns {
			return false
		}
	}

	// Skip if managed by DaemonSet
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return false
		}
	}

	// Skip completed pods
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return false
	}

	// Skip pods with local storage (unless explicitly allowed)
	for _, volume := range pod.Spec.Volumes {
		if volume.EmptyDir != nil || volume.HostPath != nil {
			return false
		}
	}

	return true
}

func (nd *NodeDrainer) evictPod(pod *corev1.Pod) error {
	ctx := context.Background()

	eviction := &policyv1beta1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}

	err := nd.clientset.PolicyV1beta1().Evictions(pod.Namespace).Evict(ctx, eviction)
	if err != nil {
		return err
	}

	log.Printf("Evicted pod: %s/%s", pod.Namespace, pod.Name)
	return nil
}

func (nd *NodeDrainer) uncordonUPSDrainedNodes() error {
	ctx := context.Background()

	// Find all nodes that were drained by us
	nodes, err := nd.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes.Items {
		// Skip control plane nodes
		if nd.isControlPlaneNode(&node) {
			continue
		}

		// Only uncordon nodes that were drained by us
		if node.Annotations != nil {
			if annotation, exists := node.Annotations[UPSDrainerAnnotation]; exists && annotation == UPSDrainerValue {
				// Uncordon the node and remove our annotation
				nodeCopy := node.DeepCopy()
				nodeCopy.Spec.Unschedulable = false
				delete(nodeCopy.Annotations, UPSDrainerAnnotation)

				_, err = nd.clientset.CoreV1().Nodes().Update(ctx, nodeCopy, metav1.UpdateOptions{})
				if err != nil {
					log.Printf("Failed to uncordon node %s: %v", node.Name, err)
					continue
				}

				log.Printf("Uncordoned node: %s", node.Name)
			}
		}
	}

	return nil
}
