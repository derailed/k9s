// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockProvider implements LLMProvider for testing.
type MockProvider struct {
	configured bool
	responses  []string
	callCount  int
}

// NewMockProvider creates a new mock provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		responses: []string{
			"I'm a mock LLM provider. I can help you with Kubernetes operations!",
			"To list pods in the current namespace, you can use: `kubectl get pods`",
			"For troubleshooting, check pod logs with: `kubectl logs <pod-name>`",
		},
	}
}

// Name returns the provider name.
func (p *MockProvider) Name() string {
	return "mock"
}

// IsConfigured returns the configuration state.
func (p *MockProvider) IsConfigured() bool {
	return p.configured
}

// Configure sets up the mock provider.
func (p *MockProvider) Configure(config *ProviderConfig) error {
	p.configured = true
	return nil
}

// SendMessage returns a mock response with beautiful formatting and k9s context awareness.
func (p *MockProvider) SendMessage(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
	if !p.configured {
		return nil, fmt.Errorf("mock provider not configured")
	}

	// Simulate API delay for realism
	time.Sleep(500 * time.Millisecond)

	// Extract context from system message
	var k9sContext K9sContextInfo
	for _, msg := range messages {
		if msg.Role == "system" {
			k9sContext = extractK9sContext(msg.Content)
			break
		}
	}

	// Get the last user message
	var userMsg string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			userMsg = strings.ToLower(messages[i].Content)
			break
		}
	}

	// Generate context-aware response
	response := p.generateContextAwareResponse(userMsg, k9sContext)

	return &Response{
		Content: response,
		Model:   "mock-k9s-assistant",
		Usage: UsageInfo{
			PromptTokens:     len(userMsg) / 4,
			CompletionTokens: len(response) / 4,
			TotalTokens:      (len(userMsg) + len(response)) / 4,
		},
	}, nil
}

// K9sContextInfo represents extracted k9s context information.
type K9sContextInfo struct {
	Cluster   string
	Context   string
	Namespace string
	View      string
	Resource  string
}

// extractK9sContext extracts k9s context from system message.
func extractK9sContext(systemMsg string) K9sContextInfo {
	ctx := K9sContextInfo{}
	
	lines := strings.Split(systemMsg, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Cluster:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ctx.Cluster = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Context:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ctx.Context = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Namespace:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ctx.Namespace = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Current View:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ctx.View = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Current Resource:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ctx.Resource = strings.TrimSpace(parts[1])
			}
		}
	}
	
	return ctx
}

// generateContextAwareResponse creates beautiful, context-aware responses.
func (p *MockProvider) generateContextAwareResponse(userMsg string, ctx K9sContextInfo) string {
	switch {
	case strings.Contains(userMsg, "pod") || strings.Contains(userMsg, "running"):
		return p.generatePodResponse(ctx)
	case strings.Contains(userMsg, "namespace") || strings.Contains(userMsg, "ns"):
		return p.generateNamespaceResponse(ctx)
	case strings.Contains(userMsg, "deployment") || strings.Contains(userMsg, "deploy"):
		return p.generateDeploymentResponse(ctx)
	case strings.Contains(userMsg, "service") || strings.Contains(userMsg, "svc"):
		return p.generateServiceResponse(ctx)
	case strings.Contains(userMsg, "current") || strings.Contains(userMsg, "status") || strings.Contains(userMsg, "state"):
		return p.generateCurrentStatusResponse(ctx)
	case strings.Contains(userMsg, "help") || strings.Contains(userMsg, "what"):
		return p.generateHelpResponse(ctx)
	case strings.Contains(userMsg, "context") || strings.Contains(userMsg, "cluster"):
		return p.generateContextResponse(ctx)
	default:
		return p.generateDefaultResponse(ctx)
	}
}

func (p *MockProvider) generatePodResponse(ctx K9sContextInfo) string {
	nsInfo := ctx.Namespace
	if nsInfo == "" || nsInfo == "all namespaces" {
		nsInfo = "all namespaces"
	}

	return fmt.Sprintf(`## 🚀 Pods in %s

Based on your current k9s context, here's how to work with pods:

### 📊 Current Context
- **Cluster:** %s
- **Namespace:** %s
- **View:** %s

### 🔍 Pod Operations

` + "```bash" + `
# List pods in current namespace
kubectl get pods

# List pods in all namespaces
kubectl get pods --all-namespaces

# Get detailed pod information
kubectl describe pod <pod-name>

# View pod logs
kubectl logs <pod-name>

# Follow pod logs
kubectl logs -f <pod-name>
` + "```" + `

### 💡 Pro Tips
- Use **Tab** in k9s to see additional pod details
- Press **l** to view logs directly in k9s
- Press **d** to describe the selected pod
- Press **s** to shell into a running pod

*Currently viewing: %s*`, nsInfo, ctx.Cluster, ctx.Namespace, ctx.View, ctx.View)
}

func (p *MockProvider) generateNamespaceResponse(ctx K9sContextInfo) string {
	return fmt.Sprintf(`## 📦 Namespace Management

### 🎯 Current Namespace Context
You're currently in namespace: **%s**

### 🔄 Namespace Operations

` + "```bash" + `
# List all namespaces
kubectl get namespaces

# Switch to a different namespace
kubectl config set-context --current --namespace=<namespace-name>

# Create a new namespace
kubectl create namespace <namespace-name>

# Delete a namespace (be careful!)
kubectl delete namespace <namespace-name>
` + "```" + `

### 🚀 k9s Namespace Tips
- Type **:ns** in k9s to switch namespaces quickly
- Use **@<namespace>** in commands to specify namespace
- Press **Ctrl+A** to toggle between all namespaces view

### 📈 Namespace Resources Summary
In your current namespace (%s), you can explore:
- Pods, Deployments, Services
- ConfigMaps, Secrets
- Ingress, NetworkPolicies
- And much more!`, ctx.Namespace, ctx.Namespace)
}

func (p *MockProvider) generateDeploymentResponse(ctx K9sContextInfo) string {
	return fmt.Sprintf(`## 🚢 Deployment Management

### 📋 Deployment Operations in %s

` + "```bash" + `
# List deployments
kubectl get deployments

# Get deployment details
kubectl describe deployment <deployment-name>

# Scale a deployment
kubectl scale deployment <deployment-name> --replicas=3

# Update deployment image
kubectl set image deployment/<deployment-name> container=image:tag

# Check rollout status
kubectl rollout status deployment/<deployment-name>

# Rollback deployment
kubectl rollout undo deployment/<deployment-name>
` + "```" + `

### 🎯 k9s Deployment Features
- Press **s** to scale deployment
- Press **r** to restart deployment
- Press **u** to rollback deployment
- Press **y** to view YAML

### 📊 Current Context
- **Cluster:** %s
- **Namespace:** %s
- **Current View:** %s

### 🔍 Best Practices
- Always check deployment status after updates
- Use resource limits and requests
- Implement health checks (readiness/liveness probes)
- Tag your images properly for easier rollbacks`, ctx.Namespace, ctx.Cluster, ctx.Namespace, ctx.View)
}

func (p *MockProvider) generateServiceResponse(ctx K9sContextInfo) string {
	return fmt.Sprintf(`## 🌐 Service Management

### 🔗 Service Operations in %s

` + "```bash" + `
# List services
kubectl get services

# Get service details
kubectl describe service <service-name>

# Expose a deployment as a service
kubectl expose deployment <deployment-name> --port=80 --target-port=8080

# Port forward to a service
kubectl port-forward service/<service-name> 8080:80

# Test service connectivity
kubectl run test-pod --image=busybox --rm -it -- wget -qO- <service-name>
` + "```" + `

### 🚀 Service Types
- **ClusterIP**: Internal cluster access only
- **NodePort**: External access via node IP:port
- **LoadBalancer**: Cloud provider load balancer
- **ExternalName**: DNS alias for external services

### 🎯 k9s Service Features
- Press **p** to view port-forwards
- Press **y** to view service YAML
- Press **d** to describe service endpoints

### 📊 Current Context
- **Cluster:** %s  
- **Namespace:** %s
- **View:** %s`, ctx.Namespace, ctx.Cluster, ctx.Namespace, ctx.View)
}

func (p *MockProvider) generateCurrentStatusResponse(ctx K9sContextInfo) string {
	return fmt.Sprintf(`## 📊 Current k9s Status

### 🎯 Your Current Context
` + "```yaml" + `
Cluster:    %s
Context:    %s
Namespace:  %s
View:       %s
Resource:   %s
` + "```" + `

### 🚀 Quick Actions Available
- **:pods** - View pods in current namespace
- **:deployments** - View deployments
- **:services** - View services
- **:ns** - Switch namespaces
- **:ctx** - Switch contexts

### 🔍 Current View Features
Based on your current view (**%s**), you can:
- Navigate with **↑/↓** arrow keys
- Press **Enter** to drill down
- Press **Esc** to go back
- Press **/** to filter resources
- Press **?** for help

### 💡 Power User Tips
- Use **Ctrl+A** to toggle all namespaces
- Press **:xray <resource>** for dependency graphs
- Use **:alias** to see available shortcuts
- Press **:q** to quit k9s

*Everything looks good! 🎉*`, 
		ctx.Cluster, ctx.Context, ctx.Namespace, ctx.View, ctx.Resource, ctx.View)
}

func (p *MockProvider) generateContextResponse(ctx K9sContextInfo) string {
	return fmt.Sprintf(`## 🌍 Kubernetes Context Information

### 🎯 Current Connection
` + "```yaml" + `
Cluster:  %s
Context:  %s
Region:   Multi-zone (simulated)
Version:  v1.29.0 (simulated)
` + "```" + `

### 🔄 Context Management

` + "```bash" + `
# List all contexts
kubectl config get-contexts

# Switch contexts
kubectl config use-context <context-name>

# Current context info
kubectl config current-context

# View cluster info
kubectl cluster-info
` + "```" + `

### 🏢 Cluster Health (Simulated)
- ✅ **API Server**: Healthy
- ✅ **ETCD**: Healthy  
- ✅ **Controller Manager**: Healthy
- ✅ **Scheduler**: Healthy
- ✅ **Nodes**: 3/3 Ready

### 🚀 k9s Context Tips
- Type **:ctx** to switch contexts quickly
- Use **:cluster** to view cluster information
- Press **Ctrl+R** to refresh current view

*Connected to %s cluster successfully! 🎉*`, ctx.Cluster, ctx.Context, ctx.Cluster)
}

func (p *MockProvider) generateHelpResponse(ctx K9sContextInfo) string {
	return fmt.Sprintf(`## 🤖 K9s Chat Assistant Help

### 👋 Welcome!
I'm your intelligent Kubernetes assistant integrated into k9s. I can help you with:

### 🎯 What I Can Do
- **Resource Management**: Pods, Deployments, Services, etc.
- **Context Awareness**: I know your current cluster, namespace, and view
- **Command Suggestions**: kubectl and k9s commands
- **Troubleshooting**: Common issues and solutions
- **Best Practices**: Kubernetes recommendations

### 💬 Example Questions
- *"What pods are running in my namespace?"*
- *"How do I scale a deployment?"*
- *"Show me service information"*
- *"What's my current context?"*
- *"Help with troubleshooting pods"*

### 🔮 Current Session
` + "```yaml" + `
Cluster:   %s
Namespace: %s  
View:      %s
Status:    Ready to help! 🚀
` + "```" + `

### 🎨 Features
- **Rich Markdown**: Beautiful formatted responses
- **Code Blocks**: Syntax-highlighted commands
- **Context Aware**: Responses based on your k9s state
- **Real-time**: Always up-to-date with your current view

### 🚀 Getting Started
Just type your Kubernetes questions naturally, and I'll provide helpful, context-aware responses!

*Ready to assist with your Kubernetes operations! 🎉*`, ctx.Cluster, ctx.Namespace, ctx.View)
}

func (p *MockProvider) generateDefaultResponse(ctx K9sContextInfo) string {
	responses := []string{
		fmt.Sprintf(`## 🎯 K9s Assistant Ready!

I'm here to help with your Kubernetes operations in **%s** namespace.

### 💡 Quick Suggestions
- Ask about **pods**, **deployments**, or **services**
- Need help with **kubectl** commands?
- Want to know your **current status**?
- Looking for **troubleshooting** tips?

Just ask me anything about Kubernetes! 🚀`, ctx.Namespace),

		fmt.Sprintf(`## 🔍 Exploring Your Cluster

Currently connected to **%s** cluster in the **%s** namespace.

### 🚀 Popular Actions
- **:pods** - View all pods
- **:deploy** - View deployments  
- **:svc** - View services
- **:logs** - View logs

What would you like to explore next? 🤔`, ctx.Cluster, ctx.Namespace),

		fmt.Sprintf(`## 📊 Cluster Overview

### 🎯 Your Current View: **%s**

Based on what you're currently viewing, I can help you:
- Understand the resources you're seeing
- Navigate to related resources
- Execute relevant commands
- Troubleshoot any issues

What specific help do you need? 💪`, ctx.View),
	}
	
	p.callCount++
	return responses[p.callCount%len(responses)]
}

// GetModels returns mock models.
func (p *MockProvider) GetModels(ctx context.Context) ([]Model, error) {
	if !p.configured {
		return nil, fmt.Errorf("mock provider not configured")
	}

	return []Model{
		{
			ID:          "mock-model",
			Name:        "Mock Model",
			Description: "A mock model for testing purposes",
		},
		{
			ID:          "mock-model-large",
			Name:        "Mock Large Model",
			Description: "A larger mock model for testing",
		},
	}, nil
}
