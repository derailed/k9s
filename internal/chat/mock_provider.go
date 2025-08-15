// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package chat

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// MockProvider implements a mock chat provider for testing.
type MockProvider struct {
	app App
}

// NewMockProvider creates a new mock provider.
func NewMockProvider(app App) *MockProvider {
	return &MockProvider{app: app}
}

// GetResponse returns a mock response based on the input message and k9s context.
func (m *MockProvider) GetResponse(message string, context K9sContext) (string, error) {
	// Simulate thinking time
	time.Sleep(time.Millisecond * time.Duration(500+rand.Intn(1500)))

	// Analyze the message to provide contextual responses
	lowerMsg := strings.ToLower(message)

	switch {
	case strings.Contains(lowerMsg, "hello") || strings.Contains(lowerMsg, "hi"):
		return m.getGreetingResponse(context), nil

	case strings.Contains(lowerMsg, "help"):
		return m.getHelpResponse(context), nil

	case strings.Contains(lowerMsg, "namespace") || strings.Contains(lowerMsg, "ns"):
		return m.getNamespaceResponse(context), nil

	case strings.Contains(lowerMsg, "pod") || strings.Contains(lowerMsg, "pods"):
		return m.getPodResponse(context), nil

	case strings.Contains(lowerMsg, "kubectl"):
		return m.getKubectlResponse(context), nil

	case strings.Contains(lowerMsg, "cluster"):
		return m.getClusterResponse(context), nil

	case strings.Contains(lowerMsg, "status") || strings.Contains(lowerMsg, "health"):
		return m.getStatusResponse(context), nil

	case strings.Contains(lowerMsg, "deploy") || strings.Contains(lowerMsg, "deployment"):
		return m.getDeploymentResponse(context), nil

	case strings.Contains(lowerMsg, "service") || strings.Contains(lowerMsg, "svc"):
		return m.getServiceResponse(context), nil

	case strings.Contains(lowerMsg, "log") || strings.Contains(lowerMsg, "logs"):
		return m.getLogsResponse(context), nil

	default:
		return m.getGeneralResponse(message, context), nil
	}
}

func (m *MockProvider) getGreetingResponse(context K9sContext) string {
	return fmt.Sprintf(`# Hello! 👋

Welcome to your Kubernetes cluster! I can see you're currently working with:

**Current Context:**
- 🏢 **Cluster:** %s
- 🏠 **Namespace:** %s  
- 👁️ **Current View:** %s

I'm here to help you with Kubernetes operations. Feel free to ask me about:
- Current resources in your namespace
- kubectl commands for your current context
- Troubleshooting tips
- Best practices

What would you like to know?`, context.ClusterName, context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getHelpResponse(context K9sContext) string {
	return fmt.Sprintf(`# K9s Chat Assistant Help 🆘

I can help you with various Kubernetes operations in your current context:

## 📍 Current Context
- **Cluster:** %s
- **Namespace:** %s
- **View:** %s

## 🔧 What I Can Do

### **Resource Management**
- Get information about pods, services, deployments
- Suggest kubectl commands for your current namespace
- Help troubleshoot resource issues

### **Context Awareness**
- Provide commands specific to your current namespace (%s)
- Suggest operations relevant to your current view (%s)
- Help navigate between different resources

### **Examples of Questions You Can Ask:**
- *"Show me pods in this namespace"*
- *"How do I check deployment status?"*
- *"What kubectl command shows service details?"*
- *"Help me troubleshoot a failing pod"*

### **Quick Commands:**
- **Tab** - Switch focus between input/messages
- **Ctrl+L** - Clear chat history
- **Esc** - Close chat

Try asking me something specific about your cluster!`,
		context.ClusterName, context.CurrentNamespace, context.CurrentView,
		context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getNamespaceResponse(context K9sContext) string {
	return fmt.Sprintf(`# Namespace Information 🏠

## Current Namespace: **%s**

Here are some useful kubectl commands for working with this namespace:

### **View Resources**
`+"```bash"+`
# List all pods in current namespace
kubectl get pods -n %s

# Get detailed pod information
kubectl get pods -n %s -o wide

# Show all resources in namespace
kubectl get all -n %s
`+"```"+`

### **Common Operations**
`+"```bash"+`
# Describe a specific pod
kubectl describe pod <pod-name> -n %s

# Get pod logs
kubectl logs <pod-name> -n %s

# Execute commands in a pod
kubectl exec -it <pod-name> -n %s -- /bin/bash
`+"```"+`

### **Namespace Management**
`+"```bash"+`
# Switch to different namespace in k9s
# Use ':ns' command or press '0' to see all namespaces

# List all namespaces
kubectl get namespaces

# Get namespace details
kubectl describe namespace %s
`+"```"+`

💡 **Tip:** In k9s, you can quickly switch namespaces using the ':ns' command or by pressing '0' to see all namespaces!`,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace)
}

func (m *MockProvider) getPodResponse(context K9sContext) string {
	return fmt.Sprintf(`# Pod Management 🚀

## Working with Pods in Namespace: **%s**

### **Basic Pod Commands**
`+"```bash"+`
# List pods in current namespace
kubectl get pods -n %s

# Watch pods in real-time
kubectl get pods -n %s --watch

# Get pod details with node information
kubectl get pods -n %s -o wide
`+"```"+`

### **Pod Troubleshooting**
`+"```bash"+`
# Check pod status and events
kubectl describe pod <pod-name> -n %s

# View pod logs
kubectl logs <pod-name> -n %s

# Follow logs in real-time
kubectl logs -f <pod-name> -n %s

# Get logs from previous container instance
kubectl logs <pod-name> -n %s --previous
`+"```"+`

### **Interactive Operations**
`+"```bash"+`
# Execute shell in running pod
kubectl exec -it <pod-name> -n %s -- /bin/bash

# Run a specific command
kubectl exec <pod-name> -n %s -- ls -la /app

# Copy files from/to pod
kubectl cp <pod-name>:/path/to/file ./local-file -n %s
`+"```"+`

### **Resource Usage**
`+"```bash"+`
# Check pod resource usage (requires metrics-server)
kubectl top pods -n %s

# Get pod YAML configuration
kubectl get pod <pod-name> -n %s -o yaml
`+"```"+`

🔍 **Current View:** %s
💡 **Tip:** In k9s, press 'l' to view logs, 's' to shell into a pod, and 'd' to describe!`,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getKubectlResponse(context K9sContext) string {
	return fmt.Sprintf(`# Kubectl Commands for Your Context 🛠️

## Current Context Configuration
- **Cluster:** %s
- **Namespace:** %s

### **Essential kubectl Commands**

#### **Viewing Resources**
`+"```bash"+`
# Quick resource overview
kubectl get all -n %s

# Specific resource types
kubectl get pods,svc,deploy -n %s

# With additional details
kubectl get pods -n %s -o wide
`+"```"+`

#### **Detailed Information**
`+"```bash"+`
# Describe any resource
kubectl describe pod <name> -n %s
kubectl describe service <name> -n %s
kubectl describe deployment <name> -n %s

# Get resource YAML
kubectl get pod <name> -n %s -o yaml
`+"```"+`

#### **Logs and Debugging**
`+"```bash"+`
# Pod logs
kubectl logs <pod-name> -n %s
kubectl logs -f <pod-name> -n %s  # Follow logs

# Multiple containers
kubectl logs <pod-name> -c <container-name> -n %s
`+"```"+`

#### **Resource Management**
`+"```bash"+`
# Scale deployments
kubectl scale deployment <name> --replicas=3 -n %s

# Delete resources
kubectl delete pod <name> -n %s
kubectl delete deployment <name> -n %s
`+"```"+`

#### **Port Forwarding**
`+"```bash"+`
# Forward local port to pod
kubectl port-forward pod/<name> 8080:80 -n %s

# Forward to service
kubectl port-forward service/<name> 8080:80 -n %s
`+"```"+`

💡 **Pro Tip:** k9s provides visual shortcuts for most of these commands. Use 'Ctrl+A' to see available actions for the selected resource!`,
		context.ClusterName, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace)
}

func (m *MockProvider) getClusterResponse(context K9sContext) string {
	return fmt.Sprintf(`# Cluster Information 🏢

## Current Cluster: **%s**

### **Cluster Overview Commands**
`+"```bash"+`
# View cluster information
kubectl cluster-info

# Check cluster status
kubectl get componentstatuses

# List all nodes
kubectl get nodes

# Node details with resource usage
kubectl get nodes -o wide
`+"```"+`

### **Cluster Resources**
`+"```bash"+`
# View all namespaces
kubectl get namespaces

# Cluster-wide resources
kubectl get pods --all-namespaces
kubectl get services --all-namespaces

# Storage classes and persistent volumes
kubectl get sc,pv
`+"```"+`

### **Resource Usage**
`+"```bash"+`
# Node resource usage (requires metrics-server)
kubectl top nodes

# Pod resource usage across cluster
kubectl top pods --all-namespaces
`+"```"+`

### **Cluster Administration**
`+"```bash"+`
# View cluster events
kubectl get events --all-namespaces --sort-by='.lastTimestamp'

# Check API versions
kubectl api-versions

# View cluster configuration
kubectl config view
`+"```"+`

### **Current Working Context**
- **Namespace:** %s
- **Current View:** %s

🌟 **k9s Tips:**
- Press ':pulses' to see cluster health overview
- Press ':contexts' to switch between clusters
- Press ':nodes' to view cluster nodes`,
		context.ClusterName, context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getStatusResponse(context K9sContext) string {
	responses := []string{
		fmt.Sprintf(`# Cluster Status Check ✅

## Current Context Status
- **Cluster:** %s ✅ Connected
- **Namespace:** %s ✅ Active  
- **View:** %s 👁️ Current

### **Quick Health Check Commands**
`+"```bash"+`
# Check cluster health
kubectl get componentstatuses

# View cluster events (recent issues)
kubectl get events -n %s --sort-by='.lastTimestamp'

# Check node status
kubectl get nodes

# Pod status in current namespace
kubectl get pods -n %s
`+"```"+`

### **Common Status Checks**
`+"```bash"+`
# All resources status
kubectl get all -n %s

# Check for failing pods
kubectl get pods -n %s --field-selector=status.phase!=Running

# Resource usage
kubectl top pods -n %s 2>/dev/null || echo "Metrics not available"
`+"```"+`

💚 **System appears healthy!** All connections working properly.`,
			context.ClusterName, context.CurrentNamespace, context.CurrentView,
			context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
			context.CurrentNamespace, context.CurrentNamespace),
	}

	return responses[rand.Intn(len(responses))]
}

func (m *MockProvider) getDeploymentResponse(context K9sContext) string {
	return fmt.Sprintf(`# Deployment Management 🚀

## Working with Deployments in: **%s**

### **View Deployments**
`+"```bash"+`
# List deployments
kubectl get deployments -n %s

# Detailed deployment info
kubectl get deployments -n %s -o wide

# Watch deployment changes
kubectl get deployments -n %s --watch
`+"```"+`

### **Deployment Details**
`+"```bash"+`
# Describe deployment
kubectl describe deployment <name> -n %s

# Get deployment YAML
kubectl get deployment <name> -n %s -o yaml

# Check rollout status
kubectl rollout status deployment/<name> -n %s
`+"```"+`

### **Scaling Operations**
`+"```bash"+`
# Scale deployment
kubectl scale deployment <name> --replicas=5 -n %s

# Auto-scaling (HPA)
kubectl autoscale deployment <name> --cpu-percent=70 --min=2 --max=10 -n %s
`+"```"+`

### **Rolling Updates**
`+"```bash"+`
# Update deployment image
kubectl set image deployment/<name> container=new-image:tag -n %s

# Rollback to previous version
kubectl rollout undo deployment/<name> -n %s

# View rollout history
kubectl rollout history deployment/<name> -n %s
`+"```"+`

### **Troubleshooting**
`+"```bash"+`
# Check deployment events
kubectl describe deployment <name> -n %s

# View replica set status
kubectl get rs -n %s

# Check pod logs from deployment
kubectl logs -l app=<deployment-label> -n %s
`+"```"+`

🎯 **Current View:** %s
💡 **k9s Tip:** Press 's' to scale, 'R' to restart, and 'l' to view logs!`,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getServiceResponse(context K9sContext) string {
	return fmt.Sprintf(`# Service Management 🌐

## Working with Services in: **%s**

### **View Services**
`+"```bash"+`
# List services
kubectl get services -n %s
kubectl get svc -n %s  # Short form

# Detailed service info
kubectl get svc -n %s -o wide

# Show service endpoints
kubectl get endpoints -n %s
`+"```"+`

### **Service Details**
`+"```bash"+`
# Describe service
kubectl describe service <name> -n %s

# Get service YAML
kubectl get service <name> -n %s -o yaml

# Check service endpoints
kubectl describe endpoints <name> -n %s
`+"```"+`

### **Service Types and Access**
`+"```bash"+`
# Port forwarding to service
kubectl port-forward service/<name> 8080:80 -n %s

# Create temporary pod to test service
kubectl run test-pod --image=busybox -it --rm -n %s -- sh

# Test service connectivity from inside cluster
# (from within test pod)
wget -qO- http://<service-name>.<namespace>.svc.cluster.local:80
`+"```"+`

### **Service Discovery**
`+"```bash"+`
# Find services by label
kubectl get svc -l app=<label> -n %s

# Check service DNS resolution
kubectl exec <pod-name> -n %s -- nslookup <service-name>

# View service account
kubectl get serviceaccount -n %s
`+"```"+`

### **LoadBalancer & NodePort Services**
`+"```bash"+`
# Check external access (if LoadBalancer/NodePort)
kubectl get svc -n %s | grep -E 'LoadBalancer|NodePort'

# Get external IPs
kubectl get svc -n %s -o jsonpath='{.items[*].status.loadBalancer.ingress[*].ip}'
`+"```"+`

🔗 **Current View:** %s
💡 **k9s Tip:** Press 'Shift+F' for port-forward, 'd' to describe service details!`,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getLogsResponse(context K9sContext) string {
	return fmt.Sprintf(`# Log Management 📋

## Viewing Logs in Namespace: **%s**

### **Basic Log Commands**
`+"```bash"+`
# View pod logs
kubectl logs <pod-name> -n %s

# Follow logs in real-time
kubectl logs -f <pod-name> -n %s

# Previous container logs (after restart)
kubectl logs <pod-name> --previous -n %s
`+"```"+`

### **Multi-Container Pods**
`+"```bash"+`
# List containers in pod
kubectl get pod <pod-name> -n %s -o jsonpath='{.spec.containers[*].name}'

# Logs from specific container
kubectl logs <pod-name> -c <container-name> -n %s

# All containers in pod
kubectl logs <pod-name> --all-containers -n %s
`+"```"+`

### **Advanced Log Options**
`+"```bash"+`
# Logs with timestamps
kubectl logs <pod-name> -n %s --timestamps

# Last N lines
kubectl logs <pod-name> -n %s --tail=100

# Logs since specific time
kubectl logs <pod-name> -n %s --since=1h

# Logs from multiple pods (by label)
kubectl logs -l app=<label> -n %s
`+"```"+`

### **Log Filtering & Search**
`+"```bash"+`
# Filter logs with grep
kubectl logs <pod-name> -n %s | grep ERROR

# Save logs to file
kubectl logs <pod-name> -n %s > pod-logs.txt

# Monitor logs with tail
kubectl logs -f <pod-name> -n %s | tail -f
`+"```"+`

### **Deployment/Service Logs**
`+"```bash"+`
# All pods in deployment
kubectl logs -l app=<deployment-name> -n %s --all-containers

# Streaming logs from all replicas
kubectl logs -f deployment/<name> -n %s
`+"```"+`

🎯 **Current View:** %s  
📱 **k9s Shortcuts:**
- 'l' - View logs for selected resource
- 'Shift+F' - Follow logs in real-time  
- 'p' - Previous logs (after pod restart)`,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
		context.CurrentNamespace, context.CurrentNamespace, context.CurrentView)
}

func (m *MockProvider) getGeneralResponse(message string, context K9sContext) string {
	responses := []string{
		fmt.Sprintf(`# Thanks for your question! 🤔

I understand you're asking about: **"%s"**

## Current Context:
- **Cluster:** %s
- **Namespace:** %s  
- **View:** %s

While I'm a mock assistant, I can help you with:

### **Common Operations**
- Viewing pods, services, deployments
- kubectl command suggestions
- Namespace and cluster information
- Log viewing and troubleshooting

### **Try asking me:**
- *"Show me pods"* - for pod management commands
- *"Help with kubectl"* - for command references  
- *"Check cluster status"* - for health information
- *"How to view logs?"* - for logging commands

**What specific Kubernetes task can I help you with?**`,
			message, context.ClusterName, context.CurrentNamespace, context.CurrentView),

		fmt.Sprintf(`# Interesting question! 💭

You asked: **"%s"**

## Let me provide some context-aware help:

**Current Environment:**
- 🏢 Cluster: %s
- 🏠 Namespace: %s
- 👁️ View: %s

Based on your current context, here are some things you might want to explore:

### **Quick Commands for %s namespace:**
`+"```bash"+`
kubectl get all -n %s
kubectl describe namespace %s
kubectl get events -n %s
`+"```"+`

### **Common Next Steps:**
1. **Explore resources** - 'kubectl get pods,svc,deploy -n %s'
2. **Check status** - 'kubectl get events -n %s --sort-by=.lastTimestamp'
3. **View details** - 'kubectl describe <resource> <name> -n %s'

**What would you like to explore further?**`,
			message, context.ClusterName, context.CurrentNamespace, context.CurrentView,
			context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
			context.CurrentNamespace, context.CurrentNamespace, context.CurrentNamespace,
			context.CurrentNamespace),
	}

	return responses[rand.Intn(len(responses))]
}
