// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package k8s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/daytonaio/runner-manager/pkg/provider/types"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sProviderConfig struct {
	Namespace   string // Namespace for placeholder pods
	WaitTimeout int    // Timeout in seconds for waiting for pod scheduling
	Kubeconfig  string // Path to kubeconfig file (optional, for local development)
}

type K8sProvider struct {
	clientset  *kubernetes.Clientset
	config     K8sProviderConfig
	jobTracker map[string]*JobStatus
	jobMutex   sync.RWMutex
}

type JobStatus struct {
	JobID     string
	PodNames  []string
	Status    string // "pending", "running", "completed", "timeout"
	StartedAt time.Time
}

// NewK8sProvider creates a new Kubernetes provider instance
func NewK8sProvider(config K8sProviderConfig) (*K8sProvider, error) {
	var k8sConfig *rest.Config
	var err error

	// If kubeconfig path is provided, use it (for local development)
	if config.Kubeconfig != "" {
		log.Infof("Using kubeconfig from: %s", config.Kubeconfig)
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create config from kubeconfig: %w", err)
		}
	} else {
		// Try in-cluster configuration (for production deployment)
		log.Info("Attempting to use in-cluster configuration")
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			// If in-cluster fails and no kubeconfig provided, try default kubeconfig location
			kubeconfigEnv := os.Getenv("KUBECONFIG")
			if kubeconfigEnv != "" {
				log.Infof("In-cluster config failed, trying KUBECONFIG env var: %s", kubeconfigEnv)
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
				if err != nil {
					return nil, fmt.Errorf("failed to create config from KUBECONFIG env: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create in-cluster config and no KUBECONFIG found: %w", err)
			}
		} else {
			log.Info("Successfully using in-cluster configuration")
		}
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &K8sProvider{
		clientset:  clientset,
		config:     config,
		jobTracker: make(map[string]*JobStatus),
	}, nil
}

func (p *K8sProvider) AddRunners(ctx context.Context, instances int) (*types.AddRunnerResponse, error) {
	if instances <= 0 {
		return nil, errors.New("instances must be greater than 0")
	}

	// Generate unique job ID
	jobID := uuid.New().String()
	podNames := make([]string, 0, instances)

	// Create placeholder pods
	for i := 0; i < instances; i++ {
		podName := fmt.Sprintf("node-placeholder-%s", uuid.New().String())
		podNames = append(podNames, podName)

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: p.config.Namespace,
				Labels: map[string]string{
					"app": "node-placeholder",
				},
				Annotations: map[string]string{
					"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyAlways,
				NodeSelector: map[string]string{
					"daytona-sandbox-c": "true",
				},
				Tolerations: []corev1.Toleration{
					{
						Key:      "sandbox",
						Operator: corev1.TolerationOpEqual,
						Value:    "true",
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "app",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{"node-placeholder"},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:  "pause",
						Image: "registry.k8s.io/pause:3.6",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		}

		_, err := p.clientset.CoreV1().Pods(p.config.Namespace).Create(ctx, pod, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create placeholder pod %s: %w", podName, err)
		}
	}

	// Store job status
	p.jobMutex.Lock()
	p.jobTracker[jobID] = &JobStatus{
		JobID:     jobID,
		PodNames:  podNames,
		Status:    "pending",
		StartedAt: time.Now(),
	}
	p.jobMutex.Unlock()

	// Launch background goroutine to wait for pod scheduling
	go p.waitForPodScheduling(context.Background(), jobID, podNames)

	return &types.AddRunnerResponse{
		JobID:    jobID,
		PodNames: podNames,
		Message:  "Runner provisioning started",
	}, nil
}

func (p *K8sProvider) RemoveRunners(ctx context.Context, instances int) error {
	return errors.New("RemoveRunners not implemented for Kubernetes provider")
}

func (p *K8sProvider) ListRunners(ctx context.Context) ([]types.RunnerInfo, error) {
	// Query all placeholder pods
	pods, err := p.clientset.CoreV1().Pods(p.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=node-placeholder",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list placeholder pods: %w", err)
	}

	runners := make([]types.RunnerInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		runnerInfo := p.buildRunnerInfo(ctx, &pod)
		runners = append(runners, runnerInfo)
	}

	return runners, nil
}

func (p *K8sProvider) GetRunner(ctx context.Context, runnerId string) (*types.RunnerInfo, error) {
	// Get the specific placeholder pod by name (runnerId is the placeholder pod name)
	pod, err := p.clientset.CoreV1().Pods(p.config.Namespace).Get(ctx, runnerId, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get placeholder pod %s: %w", runnerId, err)
	}

	// Verify it's a placeholder pod
	if pod.Labels["app"] != "node-placeholder" {
		return nil, fmt.Errorf("pod %s is not a placeholder pod", runnerId)
	}

	runnerInfo := p.buildRunnerInfo(ctx, pod)
	return &runnerInfo, nil
}

func (p *K8sProvider) GetProviderName() string {
	return "kubernetes"
}

// waitForPodScheduling waits for pods to be scheduled to nodes in the background
func (p *K8sProvider) waitForPodScheduling(ctx context.Context, jobID string, podNames []string) {
	timeout := time.Duration(p.config.WaitTimeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Mark job as running
	p.updateJobStatus(jobID, "running")

	log.Infof("Background job %s started: waiting for %d pods to be scheduled", jobID, len(podNames))

	// Poll each pod until scheduled
	for _, podName := range podNames {
		scheduled := p.waitForSinglePod(ctx, podName)
		if !scheduled {
			log.Warnf("Pod %s not scheduled within timeout", podName)
		} else {
			log.Infof("Pod %s successfully scheduled to node", podName)
		}
	}

	// Mark job as completed or timeout
	if ctx.Err() == context.DeadlineExceeded {
		log.Warnf("Background job %s timed out", jobID)
		p.updateJobStatus(jobID, "timeout")
	} else {
		log.Infof("Background job %s completed successfully", jobID)
		p.updateJobStatus(jobID, "completed")
	}
}

// waitForSinglePod waits for a single pod to be scheduled to a node
func (p *K8sProvider) waitForSinglePod(ctx context.Context, podName string) bool {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			pod, err := p.clientset.CoreV1().Pods(p.config.Namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				log.Errorf("Failed to get pod %s: %v", podName, err)
				continue
			}

			if pod.Spec.NodeName != "" {
				log.Infof("Pod %s scheduled to node %s", podName, pod.Spec.NodeName)
				return true
			}
		}
	}
}

// updateJobStatus updates the status of a background job
func (p *K8sProvider) updateJobStatus(jobID string, status string) {
	p.jobMutex.Lock()
	defer p.jobMutex.Unlock()

	if job, exists := p.jobTracker[jobID]; exists {
		job.Status = status
	}
}

// buildRunnerInfo constructs RunnerInfo from a placeholder pod
func (p *K8sProvider) buildRunnerInfo(ctx context.Context, pod *corev1.Pod) types.RunnerInfo {
	runnerInfo := types.RunnerInfo{
		Id:       pod.Name,
		Metadata: make(map[string]string),
	}

	runnerInfo.Metadata["placeholder_pod"] = pod.Name

	// Check if pod is scheduled to a node
	if pod.Spec.NodeName == "" {
		runnerInfo.Status = "pending"
		return runnerInfo
	}

	runnerInfo.Metadata["node_name"] = pod.Spec.NodeName

	// Get the node to retrieve internal IP
	node, err := p.clientset.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		runnerInfo.Status = "failed"
		runnerInfo.Metadata["error"] = fmt.Sprintf("failed to get node: %v", err)
		return runnerInfo
	}

	// Get node internal IP
	nodeIP := p.getNodeInternalIP(node)
	if nodeIP != "" {
		runnerInfo.Metadata["node_internal_ip"] = nodeIP
	}

	// Check if runner DaemonSet pod exists on this node
	runnerPod, err := p.findRunnerPodOnNode(ctx, pod.Spec.NodeName)
	if err != nil {
		runnerInfo.Status = "provisioning"
		return runnerInfo
	}

	if runnerPod != nil {
		runnerInfo.Metadata["runner_pod"] = runnerPod.Name

		// Check runner pod status
		if runnerPod.Status.Phase == corev1.PodRunning {
			runnerInfo.Status = "ready"

			// TODO: RUNNER REGISTRATION LOGIC
			// ============================================================================
			// After the runner DaemonSet pod is running, we need to register this runner
			// with the main Daytona application.
			//
			// REGISTRATION REQUIREMENTS:
			// ---------------------------
			// 1. Runner Daemon URL: Construct the full URL to access the runner daemon
			//    - The runner uses hostNetwork: true and hostPort, so it's accessible via node IP
			//    - Format: http://{node_internal_ip}:{runner_api_port}
			//    - Node IP is already available: nodeIP (from getNodeInternalIP)
			//    - Runner API port should come from configuration (e.g., RUNNER_API_PORT env var)
			//    - Default port is 8080 (see runner.yaml service.apiPort in values.yaml)
			//
			// 2. Registration Endpoint:
			//    - Main application API endpoint: POST {DAYTONA_API_URL}/runners/register
			//    - Should be configurable via environment variable: DAYTONA_API_URL
			//    - Example: http://api.daytona-dev.svc.cluster.local:3000/api/runners/register
			//
			// 3. Registration Payload:
			//    {
			//      "runner_id": "<placeholder-pod-name>",           // Use pod.Name
			//      "runner_url": "http://10.128.0.15:3001",         // Constructed URL
			//      "node_name": "<k8s-node-name>",                  // From pod.Spec.NodeName
			//      "node_internal_ip": "<node-internal-ip>",        // From nodeIP variable
			//      "status": "ready",
			//      "metadata": {
			//        "runner_pod_name": "<daytona-runner-pod-name>", // From runnerPod.Name
			//        "namespace": "<namespace>",                     // From p.config.Namespace
			//        "provisioned_at": "<timestamp>"                 // Current time
			//      }
			//    }
			//
			// 4. Authentication:
			//    - Include authentication token in request headers
			//    - Header: Authorization: Bearer {DAYTONA_API_TOKEN}
			//    - Token should come from configuration (env var: DAYTONA_API_TOKEN)
			//
			// IMPLEMENTATION STEPS:
			// ---------------------
			// Step 1: Add configuration fields to K8sProviderConfig:
			//         - RunnerAPIPort int    // Port exposed by runner daemon (from runner.yaml)
			//         - DaytonaAPIURL string // Main application API URL
			//         - DaytonaAPIToken string // Authentication token for API
			//
			// Step 2: Construct runner URL:
			//         runnerURL := fmt.Sprintf("http://%s:%d", nodeIP, p.config.RunnerAPIPort)
			//
			// Step 3: Create registration request payload (use a struct):
			//         type RunnerRegistrationRequest struct {
			//             RunnerID       string            `json:"runner_id"`
			//             RunnerURL      string            `json:"runner_url"`
			//             NodeName       string            `json:"node_name"`
			//             NodeInternalIP string            `json:"node_internal_ip"`
			//             Status         string            `json:"status"`
			//             Metadata       map[string]string `json:"metadata"`
			//         }
			//
			// Step 4: Make HTTP POST request to Daytona API:
			//         - Use standard http.Client
			//         - Set proper headers (Content-Type, Authorization)
			//         - Handle errors appropriately (log but don't fail the buildRunnerInfo)
			//         - Consider retry logic for transient failures
			//
			// Step 5: Track registration status:
			//         - Add field to JobStatus: RegisteredRunners []string
			//         - Mark runner as registered to avoid duplicate registrations
			//         - Store registration timestamp in metadata
			//
			// ERROR HANDLING:
			// ---------------
			// - Log registration failures but don't change runner status
			// - Runner should still be marked as "ready" even if registration fails
			// - Consider implementing retry mechanism for failed registrations
			// - Store registration error in metadata for debugging
			//
			// IDEMPOTENCY:
			// ------------
			// - Check if runner is already registered before making request
			// - Main API should handle duplicate registrations gracefully
			// - Consider adding "last_registered_at" timestamp to track re-registrations
			//
			// EXAMPLE IMPLEMENTATION LOCATION:
			// ---------------------------------
			// Create a new method: p.registerRunner(ctx, runnerInfo, nodeIP, runnerPod)
			// Call it here when status is "ready" and runner not yet registered
			// ============================================================================
		} else {
			runnerInfo.Status = "provisioning"
		}
	} else {
		runnerInfo.Status = "provisioning"
	}

	// Check if placeholder pod is in failed state
	if pod.Status.Phase == corev1.PodFailed {
		runnerInfo.Status = "failed"
	}

	return runnerInfo
}

// getNodeInternalIP extracts the internal IP address from a node
func (p *K8sProvider) getNodeInternalIP(node *corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

// findRunnerPodOnNode finds the runner DaemonSet pod on a specific node
func (p *K8sProvider) findRunnerPodOnNode(ctx context.Context, nodeName string) (*corev1.Pod, error) {
	pods, err := p.clientset.CoreV1().Pods(p.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=daytona-runner",
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list runner pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, nil
	}

	return &pods.Items[0], nil
}
