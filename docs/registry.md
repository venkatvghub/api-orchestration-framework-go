# Registry Package Documentation

## Overview

The `pkg/registry` package provides service discovery and component registration capabilities for the API Orchestration Framework. It offers a centralized registry for services, steps, transformers, validators, and other framework components with support for dynamic discovery, health checking, and load balancing.

## Purpose

The registry package serves as the service discovery foundation that:
- Provides centralized registration and discovery of framework components
- Enables dynamic service discovery with health checking
- Supports multiple registry backends (Consul, etcd, Kubernetes, etc.)
- Offers load balancing and failover capabilities
- Manages component lifecycle and versioning
- Provides service mesh integration capabilities
- Enables distributed configuration management

## Core Architecture

### Registry Interface

The foundation registry interface for component management:
```go
type Registry interface {
    // Service registration
    RegisterService(service *ServiceDefinition) error
    UnregisterService(serviceID string) error
    
    // Service discovery
    DiscoverServices(serviceName string) ([]*ServiceInstance, error)
    DiscoverServicesByTag(tag string) ([]*ServiceInstance, error)
    GetService(serviceID string) (*ServiceInstance, error)
    
    // Health checking
    RegisterHealthCheck(serviceID string, check HealthCheck) error
    GetServiceHealth(serviceID string) (HealthStatus, error)
    
    // Component registration
    RegisterComponent(component Component) error
    UnregisterComponent(componentID string) error
    GetComponent(componentID string) (Component, error)
    ListComponents(componentType ComponentType) ([]Component, error)
    
    // Configuration management
    SetConfig(key string, value interface{}) error
    GetConfig(key string) (interface{}, error)
    WatchConfig(key string, callback ConfigChangeCallback) error
    
    // Event handling
    Subscribe(eventType EventType, handler EventHandler) error
    Publish(event Event) error
}
```

### Service Definition

Comprehensive service definition structure:
```go
type ServiceDefinition struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Address     string            `json:"address"`
    Port        int               `json:"port"`
    Tags        []string          `json:"tags"`
    Metadata    map[string]string `json:"metadata"`
    HealthCheck *HealthCheck      `json:"health_check,omitempty"`
    
    // Service capabilities
    Endpoints   []Endpoint        `json:"endpoints"`
    Protocols   []string          `json:"protocols"`
    
    // Load balancing
    Weight      int               `json:"weight"`
    Priority    int               `json:"priority"`
    
    // Lifecycle
    TTL         time.Duration     `json:"ttl"`
    Timestamp   time.Time         `json:"timestamp"`
}

type Endpoint struct {
    Path        string            `json:"path"`
    Method      string            `json:"method"`
    Description string            `json:"description"`
    Parameters  []Parameter       `json:"parameters"`
    Responses   []Response        `json:"responses"`
}
```

### Component Registration

Framework component registration system:
```go
type Component interface {
    ID() string
    Type() ComponentType
    Name() string
    Version() string
    Description() string
    Dependencies() []string
    Metadata() map[string]interface{}
}

type ComponentType string

const (
    ComponentTypeStep        ComponentType = "step"
    ComponentTypeTransformer ComponentType = "transformer"
    ComponentTypeValidator   ComponentType = "validator"
    ComponentTypeFlow        ComponentType = "flow"
    ComponentTypeMiddleware  ComponentType = "middleware"
    ComponentTypePlugin      ComponentType = "plugin"
)
```

## Service Registration and Discovery

### Service Registration
Register services with the registry:
```go
// Register a service
serviceDefinition := &registry.ServiceDefinition{
    ID:      "user-api-001",
    Name:    "user-api",
    Version: "v1.2.3",
    Address: "192.168.1.100",
    Port:    8080,
    Tags:    []string{"api", "user", "production"},
    Metadata: map[string]string{
        "environment": "production",
        "region":      "us-west-2",
        "datacenter":  "dc1",
    },
    HealthCheck: &registry.HealthCheck{
        HTTP:     "http://192.168.1.100:8080/health",
        Interval: 30 * time.Second,
        Timeout:  5 * time.Second,
    },
    Endpoints: []registry.Endpoint{
        {
            Path:        "/api/users",
            Method:      "GET",
            Description: "List users",
        },
        {
            Path:        "/api/users/{id}",
            Method:      "GET",
            Description: "Get user by ID",
        },
    },
    Weight: 100,
    TTL:    5 * time.Minute,
}

err := registry.RegisterService(serviceDefinition)
```

### Service Discovery
Discover services from the registry:
```go
// Discover services by name
userServices, err := registry.DiscoverServices("user-api")
if err != nil {
    return err
}

// Filter healthy services
healthyServices := make([]*registry.ServiceInstance, 0)
for _, service := range userServices {
    health, err := registry.GetServiceHealth(service.ID)
    if err == nil && health.Status == registry.HealthStatusPassing {
        healthyServices = append(healthyServices, service)
    }
}

// Select service using load balancing
selectedService := loadBalancer.Select(healthyServices)
```

### Dynamic Service Discovery
Implement dynamic service discovery with caching:
```go
type ServiceDiscovery struct {
    registry    registry.Registry
    cache       map[string][]*registry.ServiceInstance
    cacheTTL    time.Duration
    lastUpdate  map[string]time.Time
    mu          sync.RWMutex
}

func (sd *ServiceDiscovery) GetServices(serviceName string) ([]*registry.ServiceInstance, error) {
    sd.mu.RLock()
    cached, exists := sd.cache[serviceName]
    lastUpdate := sd.lastUpdate[serviceName]
    sd.mu.RUnlock()
    
    // Check cache validity
    if exists && time.Since(lastUpdate) < sd.cacheTTL {
        return cached, nil
    }
    
    // Fetch from registry
    services, err := sd.registry.DiscoverServices(serviceName)
    if err != nil {
        // Return cached data if available
        if exists {
            return cached, nil
        }
        return nil, err
    }
    
    // Update cache
    sd.mu.Lock()
    sd.cache[serviceName] = services
    sd.lastUpdate[serviceName] = time.Now()
    sd.mu.Unlock()
    
    return services, nil
}
```

## Health Checking

### Health Check Types
Support multiple health check mechanisms:
```go
type HealthCheck struct {
    // HTTP health check
    HTTP     string        `json:"http,omitempty"`
    Method   string        `json:"method,omitempty"`
    Headers  map[string]string `json:"headers,omitempty"`
    
    // TCP health check
    TCP      string        `json:"tcp,omitempty"`
    
    // Script health check
    Script   string        `json:"script,omitempty"`
    Args     []string      `json:"args,omitempty"`
    
    // gRPC health check
    GRPC     string        `json:"grpc,omitempty"`
    
    // Common settings
    Interval time.Duration `json:"interval"`
    Timeout  time.Duration `json:"timeout"`
    
    // Failure handling
    DeregisterCriticalServiceAfter time.Duration `json:"deregister_critical_service_after,omitempty"`
}

type HealthStatus struct {
    Status      HealthStatusType `json:"status"`
    Output      string          `json:"output"`
    LastCheck   time.Time       `json:"last_check"`
    CheckCount  int             `json:"check_count"`
    FailureCount int            `json:"failure_count"`
}

type HealthStatusType string

const (
    HealthStatusPassing  HealthStatusType = "passing"
    HealthStatusWarning  HealthStatusType = "warning"
    HealthStatusCritical HealthStatusType = "critical"
    HealthStatusUnknown  HealthStatusType = "unknown"
)
```

### Custom Health Checks
Implement custom health check logic:
```go
// Custom health check for database connectivity
type DatabaseHealthCheck struct {
    db       *sql.DB
    query    string
    timeout  time.Duration
}

func (dhc *DatabaseHealthCheck) Check() registry.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), dhc.timeout)
    defer cancel()
    
    start := time.Now()
    err := dhc.db.PingContext(ctx)
    duration := time.Since(start)
    
    if err != nil {
        return registry.HealthStatus{
            Status:    registry.HealthStatusCritical,
            Output:    fmt.Sprintf("Database ping failed: %v", err),
            LastCheck: time.Now(),
        }
    }
    
    if duration > dhc.timeout/2 {
        return registry.HealthStatus{
            Status:    registry.HealthStatusWarning,
            Output:    fmt.Sprintf("Database ping slow: %v", duration),
            LastCheck: time.Now(),
        }
    }
    
    return registry.HealthStatus{
        Status:    registry.HealthStatusPassing,
        Output:    fmt.Sprintf("Database ping successful: %v", duration),
        LastCheck: time.Now(),
    }
}
```

## Component Registration

### Step Registration
Register custom steps with the registry:
```go
// Register a custom step
type CustomAPIStep struct {
    id          string
    name        string
    version     string
    description string
    apiURL      string
}

func (cas *CustomAPIStep) ID() string { return cas.id }
func (cas *CustomAPIStep) Type() registry.ComponentType { return registry.ComponentTypeStep }
func (cas *CustomAPIStep) Name() string { return cas.name }
func (cas *CustomAPIStep) Version() string { return cas.version }
func (cas *CustomAPIStep) Description() string { return cas.description }
func (cas *CustomAPIStep) Dependencies() []string { return []string{"http-client"} }
func (cas *CustomAPIStep) Metadata() map[string]interface{} {
    return map[string]interface{}{
        "api_url": cas.apiURL,
        "timeout": "30s",
    }
}

// Register the step
customStep := &CustomAPIStep{
    id:          "custom-api-step-v1",
    name:        "Custom API Step",
    version:     "1.0.0",
    description: "Custom step for API integration",
    apiURL:      "https://api.example.com",
}

err := registry.RegisterComponent(customStep)
```

### Transformer Registration
Register custom transformers:
```go
// Register a custom transformer
type BusinessTransformer struct {
    id          string
    name        string
    version     string
    description string
    rules       map[string]interface{}
}

func (bt *BusinessTransformer) ID() string { return bt.id }
func (bt *BusinessTransformer) Type() registry.ComponentType { return registry.ComponentTypeTransformer }
func (bt *BusinessTransformer) Name() string { return bt.name }
func (bt *BusinessTransformer) Version() string { return bt.version }
func (bt *BusinessTransformer) Description() string { return bt.description }
func (bt *BusinessTransformer) Dependencies() []string { return []string{"base-transformer"} }
func (bt *BusinessTransformer) Metadata() map[string]interface{} {
    return map[string]interface{}{
        "rules": bt.rules,
        "type":  "business-logic",
    }
}

// Register the transformer
businessTransformer := &BusinessTransformer{
    id:          "business-transformer-v2",
    name:        "Business Logic Transformer",
    version:     "2.1.0",
    description: "Applies business logic transformations",
    rules:       map[string]interface{}{"discount": 0.1},
}

err := registry.RegisterComponent(businessTransformer)
```

### Component Discovery
Discover and use registered components:
```go
// Discover steps by type
steps, err := registry.ListComponents(registry.ComponentTypeStep)
if err != nil {
    return err
}

// Find specific step
var customStep registry.Component
for _, step := range steps {
    if step.Name() == "Custom API Step" {
        customStep = step
        break
    }
}

// Use the discovered step
if customStep != nil {
    stepInstance := createStepInstance(customStep)
    flow.Step("customAPI", stepInstance)
}
```

## Load Balancing and Failover

### Load Balancing Strategies
Implement various load balancing algorithms:
```go
type LoadBalancer interface {
    Select(services []*registry.ServiceInstance) *registry.ServiceInstance
    UpdateWeights(serviceID string, weight int) error
}

// Round-robin load balancer
type RoundRobinLoadBalancer struct {
    counter uint64
}

func (rr *RoundRobinLoadBalancer) Select(services []*registry.ServiceInstance) *registry.ServiceInstance {
    if len(services) == 0 {
        return nil
    }
    
    index := atomic.AddUint64(&rr.counter, 1) % uint64(len(services))
    return services[index]
}

// Weighted load balancer
type WeightedLoadBalancer struct {
    random *rand.Rand
    mu     sync.Mutex
}

func (wlb *WeightedLoadBalancer) Select(services []*registry.ServiceInstance) *registry.ServiceInstance {
    if len(services) == 0 {
        return nil
    }
    
    totalWeight := 0
    for _, service := range services {
        totalWeight += service.Weight
    }
    
    if totalWeight == 0 {
        // Fallback to round-robin
        wlb.mu.Lock()
        index := wlb.random.Intn(len(services))
        wlb.mu.Unlock()
        return services[index]
    }
    
    wlb.mu.Lock()
    target := wlb.random.Intn(totalWeight)
    wlb.mu.Unlock()
    
    current := 0
    for _, service := range services {
        current += service.Weight
        if current > target {
            return service
        }
    }
    
    return services[len(services)-1]
}
```

### Circuit Breaker Integration
Integrate circuit breakers with service discovery:
```go
type CircuitBreakerRegistry struct {
    registry        registry.Registry
    circuitBreakers map[string]*CircuitBreaker
    mu              sync.RWMutex
}

func (cbr *CircuitBreakerRegistry) GetHealthyServices(serviceName string) ([]*registry.ServiceInstance, error) {
    services, err := cbr.registry.DiscoverServices(serviceName)
    if err != nil {
        return nil, err
    }
    
    healthyServices := make([]*registry.ServiceInstance, 0)
    
    for _, service := range services {
        cbr.mu.RLock()
        cb, exists := cbr.circuitBreakers[service.ID]
        cbr.mu.RUnlock()
        
        if !exists {
            // Create circuit breaker for new service
            cb = NewCircuitBreaker(service.ID)
            cbr.mu.Lock()
            cbr.circuitBreakers[service.ID] = cb
            cbr.mu.Unlock()
        }
        
        if cb.IsAvailable() {
            healthyServices = append(healthyServices, service)
        }
    }
    
    return healthyServices, nil
}
```

## Registry Backends

### Consul Backend
Integration with HashiCorp Consul:
```go
// Consul registry configuration
consulConfig := &registry.ConsulConfig{
    Address:    "localhost:8500",
    Datacenter: "dc1",
    Token:      "consul-token",
    Scheme:     "http",
    
    // Service registration defaults
    DefaultTTL:      30 * time.Second,
    DefaultInterval: 10 * time.Second,
    DefaultTimeout:  5 * time.Second,
}

// Create Consul registry
consulRegistry := registry.NewConsulRegistry(consulConfig)

// Register service with Consul
err := consulRegistry.RegisterService(serviceDefinition)
```

### etcd Backend
Integration with etcd:
```go
// etcd registry configuration
etcdConfig := &registry.EtcdConfig{
    Endpoints:   []string{"localhost:2379"},
    DialTimeout: 5 * time.Second,
    Username:    "etcd-user",
    Password:    "etcd-password",
    
    // Key prefix for services
    ServicePrefix: "/services/",
    ConfigPrefix:  "/config/",
}

// Create etcd registry
etcdRegistry := registry.NewEtcdRegistry(etcdConfig)
```

### Kubernetes Backend
Integration with Kubernetes service discovery:
```go
// Kubernetes registry configuration
k8sConfig := &registry.KubernetesConfig{
    Namespace:  "default",
    KubeConfig: "/path/to/kubeconfig",
    
    // Service discovery settings
    LabelSelector: "app=api-orchestration",
    FieldSelector: "status.phase=Running",
}

// Create Kubernetes registry
k8sRegistry := registry.NewKubernetesRegistry(k8sConfig)
```

### Multi-Registry Support
Use multiple registries simultaneously:
```go
// Create multi-registry
multiRegistry := registry.NewMultiRegistry(
    consulRegistry,
    etcdRegistry,
    k8sRegistry,
).WithStrategy(registry.StrategyFirstSuccess) // or StrategyAll, StrategyMajority

// Services will be registered to all registries
err := multiRegistry.RegisterService(serviceDefinition)

// Discovery will query all registries and merge results
services, err := multiRegistry.DiscoverServices("user-api")
```

## Configuration Management

### Distributed Configuration
Manage configuration across services:
```go
// Set configuration
err := registry.SetConfig("database.connection_string", "postgres://user:pass@localhost/db")
err = registry.SetConfig("cache.ttl", 300)
err = registry.SetConfig("api.rate_limit", 1000)

// Get configuration
dbConfig, err := registry.GetConfig("database.connection_string")
cacheConfig, err := registry.GetConfig("cache.ttl")

// Watch for configuration changes
err = registry.WatchConfig("api.rate_limit", func(key string, oldValue, newValue interface{}) {
    log.Info("Configuration changed",
        zap.String("key", key),
        zap.Any("old_value", oldValue),
        zap.Any("new_value", newValue))
    
    // Update rate limiter
    updateRateLimit(newValue.(int))
})
```

### Environment-specific Configuration
Manage configuration for different environments:
```go
// Environment-specific configuration
type ConfigManager struct {
    registry    registry.Registry
    environment string
}

func (cm *ConfigManager) GetConfig(key string) (interface{}, error) {
    // Try environment-specific config first
    envKey := fmt.Sprintf("%s.%s", cm.environment, key)
    value, err := cm.registry.GetConfig(envKey)
    if err == nil {
        return value, nil
    }
    
    // Fallback to global config
    return cm.registry.GetConfig(key)
}

func (cm *ConfigManager) SetConfig(key string, value interface{}) error {
    envKey := fmt.Sprintf("%s.%s", cm.environment, key)
    return cm.registry.SetConfig(envKey, value)
}
```

## Event System

### Event Publishing and Subscription
Implement event-driven architecture:
```go
type Event struct {
    Type      EventType              `json:"type"`
    Source    string                 `json:"source"`
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
    ID        string                 `json:"id"`
}

type EventType string

const (
    EventTypeServiceRegistered   EventType = "service.registered"
    EventTypeServiceUnregistered EventType = "service.unregistered"
    EventTypeServiceHealthChanged EventType = "service.health.changed"
    EventTypeConfigChanged       EventType = "config.changed"
    EventTypeComponentRegistered EventType = "component.registered"
)

// Subscribe to events
err := registry.Subscribe(registry.EventTypeServiceRegistered, func(event registry.Event) {
    log.Info("New service registered",
        zap.String("service_id", event.Data["service_id"].(string)),
        zap.String("service_name", event.Data["service_name"].(string)))
    
    // Update load balancer
    updateLoadBalancer()
})

// Publish events
event := registry.Event{
    Type:   registry.EventTypeServiceRegistered,
    Source: "api-orchestration-framework",
    Data: map[string]interface{}{
        "service_id":   "user-api-001",
        "service_name": "user-api",
        "address":      "192.168.1.100:8080",
    },
    Timestamp: time.Now(),
    ID:        generateEventID(),
}

err = registry.Publish(event)
```

## Integration with Framework Components

### HTTP Steps Integration
Integrate service discovery with HTTP steps:
```go
// Service-aware HTTP step
type ServiceHTTPStep struct {
    serviceName string
    endpoint    string
    registry    registry.Registry
    loadBalancer registry.LoadBalancer
}

func (shs *ServiceHTTPStep) Run(ctx *flow.Context) error {
    // Discover services
    services, err := shs.registry.DiscoverServices(shs.serviceName)
    if err != nil {
        return err
    }
    
    // Filter healthy services
    healthyServices := filterHealthyServices(services)
    if len(healthyServices) == 0 {
        return fmt.Errorf("no healthy services available for %s", shs.serviceName)
    }
    
    // Select service using load balancer
    selectedService := shs.loadBalancer.Select(healthyServices)
    
    // Build URL
    url := fmt.Sprintf("http://%s:%d%s", 
        selectedService.Address, selectedService.Port, shs.endpoint)
    
    // Execute HTTP request
    httpStep := http.GET(url)
    return httpStep.Run(ctx)
}
```

### Flow Registration
Register flows as discoverable components:
```go
// Register flow as component
type FlowComponent struct {
    flow        *flow.Flow
    id          string
    name        string
    version     string
    description string
}

func (fc *FlowComponent) ID() string { return fc.id }
func (fc *FlowComponent) Type() registry.ComponentType { return registry.ComponentTypeFlow }
func (fc *FlowComponent) Name() string { return fc.name }
func (fc *FlowComponent) Version() string { return fc.version }
func (fc *FlowComponent) Description() string { return fc.description }
func (fc *FlowComponent) Dependencies() []string { return fc.flow.Dependencies() }
func (fc *FlowComponent) Metadata() map[string]interface{} {
    return map[string]interface{}{
        "steps": fc.flow.StepNames(),
        "timeout": fc.flow.Timeout().String(),
    }
}

// Register flow
flowComponent := &FlowComponent{
    flow:        userRegistrationFlow,
    id:          "user-registration-flow-v1",
    name:        "User Registration Flow",
    version:     "1.0.0",
    description: "Complete user registration workflow",
}

err := registry.RegisterComponent(flowComponent)
```

## Best Practices

### Service Registration
1. **Use Meaningful IDs**: Create unique, descriptive service IDs
2. **Include Metadata**: Add relevant metadata for service discovery
3. **Set Appropriate TTL**: Configure TTL based on service lifecycle
4. **Health Check Configuration**: Implement comprehensive health checks
5. **Version Management**: Include version information for compatibility

### Service Discovery
1. **Cache Results**: Cache discovery results to reduce registry load
2. **Handle Failures**: Implement fallback mechanisms for registry failures
3. **Load Balancing**: Use appropriate load balancing strategies
4. **Circuit Breakers**: Integrate circuit breakers for resilience
5. **Monitoring**: Monitor service discovery performance and health

### Component Management
1. **Dependency Tracking**: Track component dependencies accurately
2. **Version Compatibility**: Ensure version compatibility between components
3. **Lifecycle Management**: Properly manage component lifecycle
4. **Documentation**: Include comprehensive component documentation
5. **Testing**: Test component registration and discovery

## Examples

### Complete Service Setup
```go
func SetupServiceRegistry() {
    // Create registry with multiple backends
    registry := registry.NewMultiRegistry(
        registry.NewConsulRegistry(&registry.ConsulConfig{
            Address: "consul:8500",
        }),
        registry.NewEtcdRegistry(&registry.EtcdConfig{
            Endpoints: []string{"etcd:2379"},
        }),
    )
    
    // Register current service
    serviceDefinition := &registry.ServiceDefinition{
        ID:      fmt.Sprintf("api-orchestration-%s", generateInstanceID()),
        Name:    "api-orchestration",
        Version: "2.0.0",
        Address: getLocalIP(),
        Port:    8080,
        Tags:    []string{"api", "orchestration", "production"},
        Metadata: map[string]string{
            "environment": "production",
            "region":      "us-west-2",
        },
        HealthCheck: &registry.HealthCheck{
            HTTP:     "http://localhost:8080/health",
            Interval: 30 * time.Second,
            Timeout:  5 * time.Second,
        },
        Weight: 100,
        TTL:    5 * time.Minute,
    }
    
    err := registry.RegisterService(serviceDefinition)
    if err != nil {
        log.Fatal("Failed to register service", zap.Error(err))
    }
    
    // Setup service discovery
    serviceDiscovery := NewServiceDiscovery(registry)
    
    // Setup load balancer
    loadBalancer := NewWeightedLoadBalancer()
    
    // Register components
    registerFrameworkComponents(registry)
    
    // Setup event handlers
    setupEventHandlers(registry)
    
    log.Info("Service registry setup complete")
}
```

### Dynamic Flow Composition
```go
func CreateDynamicFlow(flowName string) (*flow.Flow, error) {
    // Discover available components
    steps, err := registry.ListComponents(registry.ComponentTypeStep)
    if err != nil {
        return nil, err
    }
    
    transformers, err := registry.ListComponents(registry.ComponentTypeTransformer)
    if err != nil {
        return nil, err
    }
    
    // Create flow based on available components
    dynamicFlow := flow.NewFlow(flowName)
    
    // Add steps based on discovery
    for _, step := range steps {
        if isCompatible(step, flowName) {
            stepInstance := createStepFromComponent(step)
            dynamicFlow.Step(step.Name(), stepInstance)
        }
    }
    
    // Add transformers
    for _, transformer := range transformers {
        if isCompatible(transformer, flowName) {
            transformerInstance := createTransformerFromComponent(transformer)
            dynamicFlow.Transform(transformer.Name(), transformerInstance)
        }
    }
    
    return dynamicFlow, nil
}
```

## Troubleshooting

### Common Issues

1. **Service Registration Failures**
   ```go
   // Implement retry logic for registration
   func registerServiceWithRetry(registry registry.Registry, service *registry.ServiceDefinition) error {
       maxRetries := 3
       backoff := time.Second
       
       for i := 0; i < maxRetries; i++ {
           err := registry.RegisterService(service)
           if err == nil {
               return nil
           }
           
           log.Warn("Service registration failed, retrying",
               zap.Int("attempt", i+1),
               zap.Error(err))
           
           time.Sleep(backoff)
           backoff *= 2
       }
       
       return fmt.Errorf("failed to register service after %d attempts", maxRetries)
   }
   ```

2. **Service Discovery Timeouts**
   ```go
   // Implement timeout handling
   func discoverServicesWithTimeout(registry registry.Registry, serviceName string, timeout time.Duration) ([]*registry.ServiceInstance, error) {
       ctx, cancel := context.WithTimeout(context.Background(), timeout)
       defer cancel()
       
       resultChan := make(chan []*registry.ServiceInstance, 1)
       errorChan := make(chan error, 1)
       
       go func() {
           services, err := registry.DiscoverServices(serviceName)
           if err != nil {
               errorChan <- err
               return
           }
           resultChan <- services
       }()
       
       select {
       case services := <-resultChan:
           return services, nil
       case err := <-errorChan:
           return nil, err
       case <-ctx.Done():
           return nil, fmt.Errorf("service discovery timeout for %s", serviceName)
       }
   }
   ```

3. **Registry Backend Failures**
   ```go
   // Monitor registry health
   func monitorRegistryHealth(registry registry.Registry) {
       ticker := time.NewTicker(30 * time.Second)
       defer ticker.Stop()
       
       for range ticker.C {
           // Test registry connectivity
           testService := &registry.ServiceDefinition{
               ID:   "health-check-test",
               Name: "health-check",
           }
           
           err := registry.RegisterService(testService)
           if err != nil {
               log.Error("Registry health check failed", zap.Error(err))
               // Trigger failover or alerting
               handleRegistryFailure(err)
           } else {
               // Clean up test service
               registry.UnregisterService(testService.ID)
           }
       }
   }
   ```