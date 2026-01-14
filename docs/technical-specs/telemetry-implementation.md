# Technical Specification: CLI Telemetry Implementation

## Document Information

- **Status**: Draft
- **Author**: DataRobot Platform Team
- **Created**: 2025-12-31
- **Last Updated**: 2025-12-31
- **Version**: 1.0

## Table of Contents

- [1. Overview](#1-overview)
- [2. Goals and Non-Goals](#2-goals-and-non-goals)
- [3. Background and Context](#3-background-and-context)
- [4. Architecture Options](#4-architecture-options)
- [5. Recommended Approach](#5-recommended-approach)
- [6. Implementation Plan](#6-implementation-plan)
- [7. Privacy and Compliance](#7-privacy-and-compliance)
- [8. Testing Strategy](#8-testing-strategy)
- [9. Rollout Plan](#9-rollout-plan)
- [10. Monitoring and Success Metrics](#10-monitoring-and-success-metrics)
- [11. Open Questions](#11-open-questions)
- [12. References](#12-references)

---

## 1. Overview

This document outlines the technical specification for implementing telemetry in the DataRobot CLI (`dr`). The telemetry system will track CLI usage patterns, command execution, errors, and user interactions to improve product development, identify issues, and understand user behavior.

### 1.1 Summary

We will integrate Amplitude as the telemetry backend to capture CLI usage events. The implementation will be privacy-conscious, with opt-out mechanisms and minimal data collection focused on improving the CLI experience.

---

## 2. Goals and Non-Goals

### 2.1 Goals

1. **Track CLI Usage**: Capture command execution patterns, frequency, and success rates
2. **Error Tracking**: Identify common failure modes and error patterns
3. **Feature Adoption**: Understand which commands and features are most/least used
4. **Performance Monitoring**: Track command execution times and identify bottlenecks
5. **User Journey Analysis**: Understand common workflows and command sequences
6. **Privacy-First**: Implement with user consent and easy opt-out mechanisms
7. **Minimal Performance Impact**: Ensure telemetry doesn't degrade CLI performance

### 2.2 Non-Goals

1. **Collect Sensitive Data**: No PII, API keys, environment variables, or user-generated content
2. **Real-Time Monitoring**: Not building a real-time alerting system (use existing tools)
3. **Custom Analytics Dashboard**: Use Amplitude's existing dashboards
4. **Track Every Keystroke**: Only track command-level events, not individual keystrokes
5. **Replace Application Logging**: Telemetry supplements, not replaces, debug logs

---

## 3. Background and Context

### 3.1 Current State

The DataRobot CLI currently has:
- Basic logging framework (`charmbracelet/log`)
- HTTP client infrastructure in `internal/drapi/` and `internal/config/auth.go`
- User-Agent header already set: `GetUserAgentHeader()` returns `"DataRobot CLI version: {version}"`
- No telemetry or analytics tracking
- Version information tracked in `internal/version/version.go`

### 3.2 Technology Stack

- **Language**: Go 1.25.5+
- **CLI Framework**: Cobra
- **TUI Framework**: Bubble Tea (for interactive commands)
- **Config Management**: Viper
- **HTTP Client**: Standard `net/http`

### 3.3 Existing Infrastructure

```go
// internal/version/version.go
const CliName = "dr"
const AppName = "DataRobot CLI"
var Version = "dev"
var GitCommit = "unknown"
var BuildDate = "unknown"

// internal/config/api.go
func GetUserAgentHeader() string {
    return version.GetAppNameVersionText() // "DataRobot CLI version: {version}"
}
```

HTTP clients already exist in:
- `internal/drapi/templates.go:141` - Template API client
- `internal/config/auth.go:33` - Authentication verification

---

## 4. Architecture Options

### 4.1 Option A: Amplitude Go SDK (Ampli)

**Description**: Use Amplitude's official Go SDK ([Ampli](https://amplitude.com/docs/sdks/analytics/go/ampli-for-go)) for type-safe event tracking.

**Pros**:
- Type-safe event definitions
- Built-in batching and retry logic
- Official support and documentation
- Handles network failures gracefully
- Automatic session tracking

**Cons**:
- Adds external dependency (~100KB)
- Requires code generation for event schemas
- More complex setup process
- May have learning curve for team

**Implementation Complexity**: Medium-High

### 4.2 Option B: HTTP API with Custom Headers

**Description**: Use DataRobot API endpoints with custom telemetry headers (`X-DataRobot-Api-Consumer`) for tracking.

**Pros**:
- Leverages existing infrastructure
- No external dependencies
- Full control over data sent
- Can integrate with existing DataRobot systems

**Cons**:
- Requires building custom analytics backend
- Limited analytics capabilities without additional infrastructure
- Need to maintain custom tracking code
- No built-in batching or retry logic
- Requires coordination with backend team

**Implementation Complexity**: High (requires backend work)

### 4.3 Option C: Lightweight HTTP Amplitude Client

**Description**: Implement a minimal HTTP client that sends events directly to Amplitude's HTTP API without the full SDK.

**Pros**:
- Minimal external dependencies
- Full control over implementation
- Lightweight (<50KB code)
- Simple to understand and maintain
- Can batch events manually

**Cons**:
- Need to implement retry logic
- Need to handle rate limiting
- Manual event schema management
- Less robust than official SDK
- Need to implement session tracking

**Implementation Complexity**: Medium

### 4.4 Option D: User-Agent Enhancement Only

**Description**: Enhance existing User-Agent header with additional metadata without separate telemetry system.

**Pros**:
- Zero external dependencies
- No additional HTTP requests
- Uses existing infrastructure
- Already implemented in codebase

**Cons**:
- Very limited analytics capabilities
- Can only track API calls to DataRobot
- No visibility into command failures
- No offline command tracking
- Cannot track TUI interactions

**Implementation Complexity**: Low

---

## 5. Recommended Approach

### 5.1 Recommended Solution: Option C (Lightweight HTTP Amplitude Client)

**Rationale**:

1. **Balance of Control and Simplicity**: Option C provides enough functionality without the complexity of Option A or the infrastructure overhead of Option B
2. **Minimal Dependencies**: Keeps the CLI lightweight and reduces attack surface
3. **Go Ecosystem**: Standard library HTTP client is mature and well-tested
4. **Team Familiarity**: Team already maintains HTTP clients in the codebase
5. **Flexibility**: Can upgrade to Option A later if needed without major refactoring

### 5.2 Hybrid Approach

**Primary**: Option C (Lightweight HTTP client to Amplitude)
**Secondary**: Option D (Enhanced User-Agent for existing API calls)

This hybrid approach provides:
- Comprehensive telemetry for all CLI operations (Option C)
- Passive tracking on DataRobot API interactions (Option D)
- Redundancy if Amplitude is unreachable

---

## 6. Implementation Plan

### 6.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     DataRobot CLI (dr)                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐       ┌─────────────────────────────┐   │
│  │   Commands   │──────▶│   Telemetry Middleware      │   │
│  │  (Cobra)     │       │   (PreRunE/PostRunE hooks)  │   │
│  └──────────────┘       └─────────────────────────────┘   │
│                                     │                       │
│                                     ▼                       │
│                         ┌───────────────────────┐          │
│                         │  Telemetry Client     │          │
│                         │  (internal/telemetry) │          │
│                         └───────────────────────┘          │
│                                     │                       │
│                         ┌───────────┴────────────┐         │
│                         ▼                        ▼         │
│                  ┌──────────────┐      ┌──────────────┐   │
│                  │ Event Queue  │      │ Config Store │   │
│                  │  (in-memory) │      │   (Viper)    │   │
│                  └──────────────┘      └──────────────┘   │
│                         │                                   │
└─────────────────────────┼───────────────────────────────────┘
                          │
                          ▼
                 ┌────────────────────┐
                 │  Amplitude HTTP    │
                 │      API v2        │
                 └────────────────────┘
```

### 6.2 Component Design

#### 6.2.1 Telemetry Client (`internal/telemetry/client.go`)

```go
package telemetry

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "sync"
    "time"

    "github.com/charmbracelet/log"
    "github.com/datarobot/cli/internal/config"
    "github.com/datarobot/cli/internal/version"
)

// Client handles telemetry event collection and transmission
type Client struct {
    apiKey       string
    endpoint     string
    userID       string
    sessionID    string
    deviceID     string
    queue        []Event
    queueMutex   sync.Mutex
    httpClient   *http.Client
    enabled      bool
    batchSize    int
    flushTimer   *time.Timer
    shutdownCh   chan struct{}
}

// Event represents a telemetry event
type Event struct {
    EventType      string                 `json:"event_type"`
    UserID         string                 `json:"user_id,omitempty"`
    DeviceID       string                 `json:"device_id"`
    SessionID      int64                  `json:"session_id"`
    Time           int64                  `json:"time"`
    EventProperties map[string]interface{} `json:"event_properties,omitempty"`
    UserProperties  map[string]interface{} `json:"user_properties,omitempty"`
    AppVersion     string                 `json:"app_version"`
    Platform       string                 `json:"platform"`
    OSName         string                 `json:"os_name"`
    OSVersion      string                 `json:"os_version"`
}

// NewClient creates a new telemetry client
func NewClient(opts ...Option) (*Client, error) {
    // Implementation details
}

// Track records an event
func (c *Client) Track(ctx context.Context, eventType string, properties map[string]interface{}) error {
    if !c.enabled {
        return nil
    }

    event := Event{
        EventType:      eventType,
        DeviceID:       c.deviceID,
        SessionID:      c.sessionID,
        Time:           time.Now().UnixMilli(),
        EventProperties: properties,
        AppVersion:     version.Version,
        Platform:       runtime.GOOS,
        OSName:         runtime.GOOS,
        OSVersion:      getOSVersion(),
    }

    c.queueMutex.Lock()
    c.queue = append(c.queue, event)
    shouldFlush := len(c.queue) >= c.batchSize
    c.queueMutex.Unlock()

    if shouldFlush {
        go c.Flush(ctx)
    }

    return nil
}

// Flush sends queued events to Amplitude
func (c *Client) Flush(ctx context.Context) error {
    // Implementation details
}

// Shutdown gracefully shuts down the client
func (c *Client) Shutdown(ctx context.Context) error {
    // Implementation details
}
```

#### 6.2.2 Configuration (`internal/telemetry/config.go`)

```go
package telemetry

import (
    "github.com/spf13/viper"
)

const (
    // Config keys
    ConfigKeyEnabled        = "telemetry.enabled"
    ConfigKeyDeviceID       = "telemetry.device_id"
    ConfigKeyOptInTimestamp = "telemetry.opt_in_timestamp"
    ConfigKeyOptOutReason   = "telemetry.opt_out_reason"

    // Default values
    DefaultBatchSize     = 10
    DefaultFlushInterval = 30 * time.Second
    DefaultTimeout       = 10 * time.Second
)

// IsEnabled returns whether telemetry is enabled
func IsEnabled() bool {
    // Check env var first (for CI/CD)
    if envVal := os.Getenv("DR_TELEMETRY_ENABLED"); envVal != "" {
        return envVal == "true" || envVal == "1"
    }

    // Default to opt-out (enabled) with prompt on first use
    return viper.GetBool(ConfigKeyEnabled)
}

// GetDeviceID returns a stable device identifier
func GetDeviceID() string {
    deviceID := viper.GetString(ConfigKeyDeviceID)
    if deviceID == "" {
        deviceID = generateDeviceID()
        viper.Set(ConfigKeyDeviceID, deviceID)
        _ = viper.WriteConfig()
    }
    return deviceID
}

// OptIn enables telemetry
func OptIn() error {
    viper.Set(ConfigKeyEnabled, true)
    viper.Set(ConfigKeyOptInTimestamp, time.Now().Unix())
    return viper.WriteConfig()
}

// OptOut disables telemetry
func OptOut(reason string) error {
    viper.Set(ConfigKeyEnabled, false)
    viper.Set(ConfigKeyOptOutReason, reason)
    return viper.WriteConfig()
}
```

#### 6.2.3 Middleware (`internal/telemetry/middleware.go`)

```go
package telemetry

import (
    "context"
    "time"

    "github.com/spf13/cobra"
)

var globalClient *Client

// InitializeGlobalClient sets up the global telemetry client
func InitializeGlobalClient() error {
    if !IsEnabled() {
        return nil
    }

    client, err := NewClient(
        WithAPIKey(getAmplitudeAPIKey()),
        WithDeviceID(GetDeviceID()),
        WithBatchSize(DefaultBatchSize),
        WithFlushInterval(DefaultFlushInterval),
    )
    if err != nil {
        return err
    }

    globalClient = client
    return nil
}

// InjectMiddleware adds telemetry hooks to a Cobra command
func InjectMiddleware(cmd *cobra.Command) {
    // Store original hooks
    originalPreRunE := cmd.PreRunE
    originalPostRunE := cmd.PostRunE

    // Inject pre-run hook
    cmd.PreRunE = func(c *cobra.Command, args []string) error {
        startTime := time.Now()
        ctx := context.WithValue(c.Context(), "telemetry_start_time", startTime)
        c.SetContext(ctx)

        // Track command start
        if globalClient != nil {
            _ = globalClient.Track(ctx, "command_started", map[string]interface{}{
                "command":      c.CommandPath(),
                "args_count":   len(args),
                "flags":        getFlagValues(c),
            })
        }

        // Call original pre-run
        if originalPreRunE != nil {
            return originalPreRunE(c, args)
        }
        return nil
    }

    // Inject post-run hook
    cmd.PostRunE = func(c *cobra.Command, args []string) error {
        ctx := c.Context()
        startTime, _ := ctx.Value("telemetry_start_time").(time.Time)
        duration := time.Since(startTime)

        // Track command completion
        if globalClient != nil {
            _ = globalClient.Track(ctx, "command_completed", map[string]interface{}{
                "command":      c.CommandPath(),
                "duration_ms":  duration.Milliseconds(),
                "success":      true,
            })
        }

        // Call original post-run
        if originalPostRunE != nil {
            return originalPostRunE(c, args)
        }
        return nil
    }

    // Recursively inject into subcommands
    for _, subCmd := range cmd.Commands() {
        InjectMiddleware(subCmd)
    }
}

// ShutdownGlobalClient gracefully shuts down telemetry
func ShutdownGlobalClient(ctx context.Context) error {
    if globalClient != nil {
        return globalClient.Shutdown(ctx)
    }
    return nil
}
```

#### 6.2.4 Events Schema (`internal/telemetry/events.go`)

```go
package telemetry

const (
    // Command events
    EventCommandStarted   = "command_started"
    EventCommandCompleted = "command_completed"
    EventCommandFailed    = "command_failed"

    // Authentication events
    EventAuthLoginStarted   = "auth_login_started"
    EventAuthLoginSucceeded = "auth_login_succeeded"
    EventAuthLoginFailed    = "auth_login_failed"
    EventAuthLogout         = "auth_logout"
    EventAuthSetURL         = "auth_set_url"

    // Template events
    EventTemplateList    = "template_list"
    EventTemplateClone   = "template_clone"
    EventTemplateSetup   = "template_setup"

    // Environment events
    EventDotenvSetup    = "dotenv_setup"
    EventDotenvEdit     = "dotenv_edit"
    EventDotenvValidate = "dotenv_validate"

    // Task events
    EventTaskList    = "task_list"
    EventTaskRun     = "task_run"
    EventTaskSuccess = "task_success"
    EventTaskFailed  = "task_failed"

    // Self-management events
    EventSelfUpdate  = "self_update"
    EventSelfVersion = "self_version"

    // Error events
    EventError = "error"

    // TUI events
    EventTUIInteraction = "tui_interaction"
)

// CommandProperties returns standard command properties
func CommandProperties(cmd string, args []string) map[string]interface{} {
    return map[string]interface{}{
        "command":    cmd,
        "args_count": len(args),
    }
}

// ErrorProperties returns standard error properties
func ErrorProperties(err error, context string) map[string]interface{} {
    return map[string]interface{}{
        "error_type":    fmt.Sprintf("%T", err),
        "error_message": err.Error(),
        "context":       context,
    }
}
```

#### 6.2.5 Enhanced User-Agent (`internal/config/api.go`)

```go
// GetUserAgentHeader returns the User-Agent header value with enhanced telemetry info
func GetUserAgentHeader() string {
    base := version.GetAppNameVersionText()
    
    // Add platform info
    platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
    
    // Add session ID if telemetry enabled
    sessionInfo := ""
    if telemetry.IsEnabled() {
        sessionInfo = fmt.Sprintf("; session=%s", telemetry.GetSessionID())
    }
    
    return fmt.Sprintf("%s (%s%s)", base, platform, sessionInfo)
}

// GetTelemetryHeaders returns additional headers for DataRobot API calls
func GetTelemetryHeaders() map[string]string {
    headers := make(map[string]string)
    
    headers["X-DataRobot-Api-Consumer"] = "cli"
    headers["X-DataRobot-CLI-Version"] = version.Version
    headers["X-DataRobot-CLI-Platform"] = runtime.GOOS
    
    if telemetry.IsEnabled() {
        headers["X-DataRobot-CLI-Session"] = telemetry.GetSessionID()
        headers["X-DataRobot-CLI-Device"] = telemetry.GetDeviceID()
    }
    
    return headers
}
```

### 6.3 Integration Points

#### 6.3.1 Root Command (`cmd/root.go`)

```go
func ExecuteContext(ctx context.Context) error {
    // Initialize telemetry
    if err := telemetry.InitializeGlobalClient(); err != nil {
        log.Debug("Failed to initialize telemetry", "error", err)
    }

    // Inject middleware into all commands
    telemetry.InjectMiddleware(RootCmd)

    // Execute command
    err := RootCmd.ExecuteContext(ctx)

    // Shutdown telemetry
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = telemetry.ShutdownGlobalClient(shutdownCtx)

    return err
}
```

#### 6.3.2 Error Handling

```go
// In any command that returns an error
func (cmd *Command) RunE(c *cobra.Command, args []string) error {
    err := doSomething()
    if err != nil {
        // Track error
        if telemetry.IsEnabled() {
            _ = telemetry.Track(c.Context(), telemetry.EventError, 
                telemetry.ErrorProperties(err, c.CommandPath()))
        }
        return err
    }
    return nil
}
```

### 6.4 File Structure

```
internal/telemetry/
├── client.go           # Main telemetry client implementation
├── client_test.go      # Client unit tests
├── config.go           # Configuration management
├── config_test.go      # Config unit tests
├── middleware.go       # Cobra middleware integration
├── middleware_test.go  # Middleware unit tests
├── events.go           # Event type definitions and helpers
├── device.go           # Device ID generation and storage
├── mock.go             # Mock client for testing
└── README.md           # Package documentation
```

### 6.5 Configuration Schema

Add to `~/.config/datarobot/drconfig.yaml`:

```yaml
telemetry:
  enabled: true  # Default: true (opt-out model)
  device_id: "uuid-generated-once"
  opt_in_timestamp: 1704067200
  last_prompt_version: "1.0"  # Track which consent version was shown
```

### 6.6 Environment Variables

```bash
# Disable telemetry (useful for CI/CD)
DR_TELEMETRY_ENABLED=false

# Custom Amplitude endpoint (for testing)
DR_TELEMETRY_ENDPOINT=https://custom-endpoint.example.com

# Enable telemetry debug logging
DR_TELEMETRY_DEBUG=true
```

---

## 7. Privacy and Compliance

### 7.1 Data Collection Policy

**What We Collect**:
- Command names and subcommands (e.g., `dr templates list`)
- Flag names (not values)
- Command execution duration
- Success/failure status
- Error types (not error messages with user data)
- CLI version and platform information
- Anonymized device ID (generated UUID)
- Session ID (generated per CLI invocation)

**What We DO NOT Collect**:
- Personally Identifiable Information (PII)
- DataRobot API keys or credentials
- Environment variable values
- File paths or contents
- User input values
- Command arguments with potentially sensitive data
- IP addresses (Amplitude can be configured to not store IPs)
- Template names or custom data

### 7.2 Consent Mechanism

#### 7.2.1 First-Run Experience

On first CLI use, show a consent prompt:

```
╭─────────────────────────────────────────────────────────────╮
│                                                             │
│  📊 Help us improve the DataRobot CLI                       │
│                                                             │
│  We'd like to collect anonymous usage data to improve      │
│  the CLI. This includes:                                   │
│                                                             │
│  • Commands you run (e.g., 'dr templates list')            │
│  • Success/failure rates                                   │
│  • Performance metrics                                     │
│                                                             │
│  We never collect:                                         │
│  • Personal information or credentials                     │
│  • File contents or paths                                  │
│  • Input values or arguments                               │
│                                                             │
│  You can opt-out anytime with: dr self telemetry disable   │
│                                                             │
│  Learn more: https://docs.datarobot.com/cli/telemetry      │
│                                                             │
│  Allow anonymous usage tracking? [Y/n]                     │
│                                                             │
╰─────────────────────────────────────────────────────────────╯
```

#### 7.2.2 Telemetry Commands

```bash
# Check telemetry status
dr self telemetry status

# Disable telemetry
dr self telemetry disable [--reason "privacy concerns"]

# Enable telemetry
dr self telemetry enable

# View what data is collected
dr self telemetry info
```

Implementation in `cmd/self/telemetry.go`:

```go
var telemetryCmd = &cobra.Command{
    Use:   "telemetry",
    Short: "Manage telemetry settings",
    Long: `View and configure CLI telemetry settings.
    
Telemetry helps us understand how the CLI is used and improve it.
All data collected is anonymous and does not include sensitive information.`,
}

var telemetryStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show current telemetry status",
    RunE: func(cmd *cobra.Command, args []string) error {
        enabled := telemetry.IsEnabled()
        deviceID := telemetry.GetDeviceID()
        
        if enabled {
            fmt.Println("✅ Telemetry is enabled")
            fmt.Printf("📱 Device ID: %s\n", deviceID)
        } else {
            fmt.Println("❌ Telemetry is disabled")
        }
        return nil
    },
}

var telemetryDisableCmd = &cobra.Command{
    Use:   "disable",
    Short: "Disable telemetry",
    RunE: func(cmd *cobra.Command, args []string) error {
        reason, _ := cmd.Flags().GetString("reason")
        if err := telemetry.OptOut(reason); err != nil {
            return err
        }
        fmt.Println("✅ Telemetry disabled successfully")
        return nil
    },
}
```

### 7.3 GDPR and Privacy Compliance

1. **Right to Access**: Users can see their device ID with `dr self telemetry status`
2. **Right to Erasure**: Contact support to delete data associated with device ID
3. **Right to Object**: Simple opt-out with `dr self telemetry disable`
4. **Data Minimization**: Only collect essential metrics
5. **Purpose Limitation**: Data only used for CLI improvement
6. **Storage Limitation**: Amplitude retention period: 90 days (configurable)

### 7.4 Security Considerations

1. **Data in Transit**: All data sent over HTTPS (TLS 1.2+)
2. **API Key Storage**: Amplitude API key embedded in binary (low risk for client-side analytics)
3. **No Credentials**: Never transmit DataRobot credentials or API keys
4. **Rate Limiting**: Implement client-side rate limiting to prevent data exfiltration
5. **Timeout Handling**: All telemetry requests timeout after 10 seconds
6. **Fail Silent**: Telemetry failures never block CLI operations

---

## 8. Testing Strategy

### 8.1 Unit Tests

```go
// internal/telemetry/client_test.go
func TestClientTrack(t *testing.T) {
    // Test event tracking
    client := NewMockClient()
    err := client.Track(context.Background(), "test_event", map[string]interface{}{
        "property": "value",
    })
    assert.NoError(t, err)
}

func TestClientFlush(t *testing.T) {
    // Test batch flushing
}

func TestClientDisabled(t *testing.T) {
    // Test that disabled client doesn't send events
}
```

### 8.2 Integration Tests

```go
// internal/telemetry/integration_test.go
func TestEndToEndTracking(t *testing.T) {
    // Set up test Amplitude endpoint
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request format
        assert.Equal(t, "POST", r.Method)
        assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
        
        // Decode body
        var payload map[string]interface{}
        json.NewDecoder(r.Body).Decode(&payload)
        
        // Verify event structure
        assert.Contains(t, payload, "events")
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{"code": 200})
    }))
    defer mockServer.Close()
    
    // Test with mock server
}
```

### 8.3 Manual Testing

```bash
# Enable debug mode
export DR_TELEMETRY_DEBUG=true

# Run commands and verify events
dr --debug templates list

# Check that events are sent
# (View network requests in debug logs)

# Test opt-out
dr self telemetry disable

# Verify no events sent
dr templates list
```

### 8.4 Smoke Tests

Add to `smoke_test_scripts/`:

```bash
#!/bin/bash
# Test telemetry functionality

# Test 1: Verify telemetry status command
dr self telemetry status

# Test 2: Disable telemetry
dr self telemetry disable --reason "testing"

# Test 3: Run command with telemetry disabled
dr templates list

# Test 4: Enable telemetry
dr self telemetry enable

# Test 5: Run command with telemetry enabled
dr templates list
```

---

## 9. Rollout Plan

### 9.1 Phase 1: Foundation (Sprint 1-2)

**Goal**: Build core telemetry infrastructure

**Tasks**:
1. Create `internal/telemetry` package structure
2. Implement `Client` with basic HTTP functionality
3. Implement device ID generation and storage
4. Add configuration management (enable/disable)
5. Create unit tests for core functionality
6. Document package API

**Success Criteria**:
- Unit tests pass with >80% coverage
- Mock client can track events
- Configuration persists correctly

### 9.2 Phase 2: Integration (Sprint 3)

**Goal**: Integrate telemetry into CLI

**Tasks**:
1. Implement Cobra middleware
2. Add telemetry to root command
3. Implement command start/complete events
4. Add error tracking
5. Create telemetry management commands (`dr self telemetry`)
6. Add consent prompt for first-time users

**Success Criteria**:
- Events tracked for all commands
- Opt-in/opt-out flow works
- Middleware doesn't break existing commands

### 9.3 Phase 3: Enhanced Tracking (Sprint 4)

**Goal**: Add detailed event tracking

**Tasks**:
1. Add specific events for auth commands
2. Add events for template operations
3. Add events for dotenv operations
4. Add events for task operations
5. Implement TUI interaction tracking (limited)
6. Enhance User-Agent header

**Success Criteria**:
- All major command groups emit events
- User-Agent includes platform info
- Custom headers added to DataRobot API calls

### 9.4 Phase 4: Polish and Release (Sprint 5)

**Goal**: Production-ready release

**Tasks**:
1. Performance optimization
2. Error handling improvements
3. Documentation (user-facing and internal)
4. Privacy policy updates
5. Integration tests
6. Beta testing with internal users
7. Amplitude dashboard setup

**Success Criteria**:
- Performance impact <10ms per command
- 100% of events successfully batched
- Documentation complete
- Beta testers provide positive feedback

### 9.5 Rollout Strategy

**Week 1-2**: Internal alpha testing (dev team only)
**Week 3-4**: Beta release (opt-in for early adopters)
**Week 5**: General availability with opt-out model
**Week 6+**: Monitor and iterate based on feedback

---

## 10. Monitoring and Success Metrics

### 10.1 Technical Metrics

**Performance**:
- Telemetry overhead per command: <10ms (p95)
- Event queue flush time: <500ms (p95)
- Memory overhead: <5MB
- Batch success rate: >99%
- Event loss rate: <0.1%

**Reliability**:
- Telemetry failures don't block CLI: 100%
- Graceful degradation on network failure: 100%
- Config corruption rate: 0%

**Adoption**:
- Opt-out rate: <20% (industry standard)
- Active devices (7-day): Track growth
- Events per user per day: Track trends

### 10.2 Product Metrics (via Amplitude)

**Usage Patterns**:
- Most/least used commands
- Command completion rates
- Average session duration
- Command sequences (funnels)

**Error Rates**:
- Errors by command
- Errors by platform
- Errors by version
- Error trends over time

**Feature Adoption**:
- New feature usage rates
- Template selection distribution
- Authentication method distribution

**Performance**:
- Command duration by type
- Slow commands identification
- Performance regression detection

### 10.3 Amplitude Dashboard Setup

**Dashboards to Create**:

1. **Executive Dashboard**:
   - Daily Active Users (DAU)
   - Weekly Active Users (WAU)
   - Monthly Active Users (MAU)
   - Commands per user
   - Version distribution

2. **Command Usage Dashboard**:
   - Command frequency
   - Command success rates
   - Command duration
   - Command sequences

3. **Error Dashboard**:
   - Error rate trends
   - Top errors by type
   - Errors by command
   - Errors by platform/version

4. **Feature Adoption Dashboard**:
   - New feature usage
   - Template selection
   - TUI vs direct command usage

---

## 11. Open Questions

### 11.1 Technical Questions

1. **Q**: Should we implement local event persistence for offline scenarios?
   - **A**: Phase 2 consideration - implement if users frequently work offline
   - **Decision**: Start without, add if needed based on feedback

2. **Q**: How do we handle telemetry in CI/CD environments?
   - **A**: Auto-detect CI environments (check `CI` env var) and disable by default
   - **Decision**: Add to Phase 1

3. **Q**: Should we track TUI interactions (e.g., arrow key presses in lists)?
   - **A**: Too granular and privacy-invasive
   - **Decision**: Only track TUI screen transitions and selections

4. **Q**: What's the retry policy for failed event sends?
   - **A**: Exponential backoff: 1s, 2s, 4s, then drop
   - **Decision**: Implement in Phase 1

5. **Q**: Should we correlate CLI events with DataRobot backend events?
   - **A**: Yes, via session ID in custom headers
   - **Decision**: Implement in Phase 3

### 11.2 Product Questions

1. **Q**: Should telemetry be opt-in or opt-out?
   - **Current Approach**: Opt-out (enabled by default with prominent first-run notice)
   - **Rationale**: Industry standard for dev tools, maximizes data collection while respecting privacy

2. **Q**: How do we handle enterprise customers with strict privacy policies?
   - **A**: Provide environment variable override: `DR_TELEMETRY_ENABLED=false`
   - **Decision**: Document in enterprise deployment guides

3. **Q**: Should we expose telemetry data to users?
   - **A**: Limited exposure via `dr self telemetry info` showing what would be sent
   - **Decision**: Phase 4 feature

### 11.3 Privacy Questions

1. **Q**: Can users request data deletion?
   - **A**: Yes, contact support with device ID
   - **Decision**: Document in privacy policy

2. **Q**: What's the data retention period?
   - **A**: 90 days in Amplitude (configurable)
   - **Decision**: Align with company policy

3. **Q**: Do we need explicit consent for GDPR?
   - **A**: Opt-out model with clear notice is acceptable for legitimate interest
   - **Decision**: Legal team to confirm

---

## 12. References

### 12.1 External Documentation

- [Amplitude HTTP API v2](https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/)
- [Amplitude Go SDK](https://amplitude.com/docs/sdks/analytics/go/ampli-for-go)
- [Amplitude Best Practices](https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/#best-practices)
- [GDPR Compliance for Analytics](https://gdpr.eu/cookies/)

### 12.2 Similar Implementations

Reference implementations to study:
- [Homebrew telemetry](https://github.com/Homebrew/brew/blob/master/Library/Homebrew/utils/analytics.sh)
- [Netlify CLI telemetry](https://github.com/netlify/cli/blob/main/src/utils/telemetry/index.ts)
- [Azure CLI telemetry](https://github.com/Azure/azure-cli/tree/dev/src/azure-cli-telemetry)
- [GitHub CLI telemetry](https://github.com/cli/cli/tree/trunk/pkg/cmd/config)

### 12.3 Internal Documentation

- CLI Architecture: `docs/development/building.md`
- Configuration System: `docs/user-guide/configuration.md`
- Command Structure: `cmd/README.md`
- API Client: `internal/drapi/templates.go`

### 12.4 Amplitude Configuration

**API Key Management**:
- Development: Use separate Amplitude project
- Production: Use production Amplitude project
- Store API key: Embed in binary (acceptable for client-side analytics)

**Amplitude Setup**:
```go
const (
    AmplitudeEndpoint = "https://api2.amplitude.com/2/httpapi"
    AmplitudeAPIKey   = "REPLACE_WITH_ACTUAL_KEY" // From Amplitude console
)
```

---

## 13. Appendix

### 13.1 Example Event Payloads

#### Command Execution Event

```json
{
  "api_key": "AMPLITUDE_API_KEY",
  "events": [
    {
      "event_type": "command_completed",
      "user_id": null,
      "device_id": "550e8400-e29b-41d4-a716-446655440000",
      "session_id": 1704067200000,
      "time": 1704067201500,
      "event_properties": {
        "command": "dr templates list",
        "duration_ms": 1500,
        "success": true,
        "flags": ["--verbose"]
      },
      "app_version": "0.3.0",
      "platform": "darwin",
      "os_name": "darwin",
      "os_version": "14.2"
    }
  ]
}
```

#### Error Event

```json
{
  "api_key": "AMPLITUDE_API_KEY",
  "events": [
    {
      "event_type": "error",
      "device_id": "550e8400-e29b-41d4-a716-446655440000",
      "session_id": 1704067200000,
      "time": 1704067201500,
      "event_properties": {
        "error_type": "*url.Error",
        "context": "dr templates list",
        "command": "dr templates list"
      },
      "app_version": "0.3.0",
      "platform": "darwin"
    }
  ]
}
```

### 13.2 Configuration Examples

**Disable telemetry globally**:

```bash
# Via environment variable (recommended for CI/CD)
export DR_TELEMETRY_ENABLED=false

# Via CLI command (persists to config file)
dr self telemetry disable
```

**Config file after opt-out**:

```yaml
# ~/.config/datarobot/drconfig.yaml
telemetry:
  enabled: false
  device_id: "550e8400-e29b-41d4-a716-446655440000"
  opt_out_reason: "privacy concerns"
  last_updated: 1704067200
```

### 13.3 Performance Benchmarks

Target benchmarks for telemetry operations:

```
BenchmarkTrackEvent-8             100000     10523 ns/op      2048 B/op      10 allocs/op
BenchmarkBatchFlush-8              10000    156789 ns/op     16384 B/op      50 allocs/op
BenchmarkDeviceIDGeneration-8    1000000       823 ns/op       256 B/op       5 allocs/op
BenchmarkConfigRead-8             500000      2341 ns/op       512 B/op       8 allocs/op
```

### 13.4 Error Handling Examples

```go
// Example: Track error without blocking command
func executeCommand(ctx context.Context, cmd string) error {
    err := doCommand(cmd)
    if err != nil {
        // Track error (non-blocking)
        go func() {
            trackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            _ = telemetry.Track(trackCtx, telemetry.EventError, 
                telemetry.ErrorProperties(err, cmd))
        }()
        return err
    }
    return nil
}
```

---

## Document Change Log

| Version | Date       | Author | Changes |
|---------|------------|--------|---------|
| 1.0     | 2025-12-31 | DataRobot Platform Team | Initial draft |

---

## Approval

**Technical Review**:
- [ ] Engineering Lead
- [ ] Platform Architect
- [ ] Security Team

**Product Review**:
- [ ] Product Manager
- [ ] UX Designer

**Legal Review**:
- [ ] Legal/Privacy Team

**Final Approval**:
- [ ] Engineering Director
