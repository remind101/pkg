# Instructions for Modernizing Go Repositories with OpenHands

This document provides a structured approach to modernizing Go codebases efficiently using OpenHands, minimizing unnecessary work and costs.

## 1. Initial Setup

### Installing Go
```bash
# Check if Go is already installed
go version

# If not installed or outdated, download and install the latest version
# For Linux (adjust URL for other platforms)
wget https://go.dev/dl/go1.22.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

### Create a Tracking File
Create a file called `modernization.md` to track progress:

```bash
touch modernization.md
```

Initial content structure:
```markdown
# Go Modernization Tracking

## Identified Issues
- [ ] Issue 1
- [ ] Issue 2

## Completed Changes
- [ ] Change 1
- [ ] Change 2

## Testing Status
- [ ] Package 1 tests passing
- [ ] Package 2 tests passing
```

## 2. Repository Analysis

### Analyze Go Version
```bash
# Check Go version used in the project
grep -r "go [0-9]" --include="*.mod" .
```

### Identify Deprecated APIs
Common deprecated APIs to look for:
```bash
# HTTP Transport using Dial instead of DialContext
grep -r "Dial:" --include="*.go" .
grep -r "http.Transport{" --include="*.go" .

# HTTP requests without context
grep -r "http.NewRequest(" --include="*.go" .

# Context-free operations
grep -r "WithContext" --include="*.go" .
```

### Check Dependencies
```bash
# List and check dependencies
go list -m all
```

## 3. Systematic Modernization Approach

### 1. Update Core Infrastructure First
Start with the most foundational packages that other code depends on:

1. **HTTP Transport Configuration**:
   - Replace `Dial` with `DialContext`
   - Add modern connection parameters:
     - `MaxIdleConns`
     - `MaxIdleConnsPerHost`
     - `IdleConnTimeout`
     - `TLSHandshakeTimeout`

2. **HTTP Client Code**:
   - Replace `http.NewRequest` with `http.NewRequestWithContext`
   - Ensure context is properly propagated

### 2. Update Dependent Packages
After updating core infrastructure, move to packages that depend on them:

1. **Service Clients**:
   - Update to use context-aware methods
   - Ensure proper error handling with contexts

2. **Middleware**:
   - Update timeout middleware to use context cancellation
   - Update authentication/authorization middleware

### 3. Update Tests
Update test files to match the modernized implementation:
   - Update mock expectations
   - Ensure tests use context-aware methods
   - Add new tests for context cancellation if needed

## 4. Common Modernization Patterns

### HTTP Transport Modernization
```go
// OLD
transport := &http.Transport{
    Dial: (&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }).Dial,
    TLSHandshakeTimeout: 10 * time.Second,
}

// NEW
transport := &http.Transport{
    DialContext: (&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }).DialContext,
    TLSHandshakeTimeout: 10 * time.Second,
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 100,
    IdleConnTimeout:     90 * time.Second,
}
```

### HTTP Request Modernization
```go
// OLD
req, err := http.NewRequest("GET", url, body)

// NEW
ctx := context.Background() // or use a passed context
req, err := http.NewRequestWithContext(ctx, "GET", url, body)
```

### Context Propagation
```go
// OLD
func DoSomething(params Params) (*Result, error) {
    // ...
}

// NEW
func DoSomething(ctx context.Context, params Params) (*Result, error) {
    // Use ctx for cancellation, timeouts, etc.
    // ...
}
```

## 5. Efficient Testing Strategy

### Run Tests Incrementally
```bash
# Test specific package after changes
go test ./path/to/package

# Test all packages
go test ./...
```

### Use Build Tags for Compatibility Testing
If maintaining backward compatibility:
```go
// +build go1.16

package example
```

## 6. Commit Strategy

### Atomic Commits
Make focused commits that address specific modernization concerns:

1. Create a branch for modernization:
   ```bash
   git checkout -b modernize-go-packages
   ```

2. Make atomic commits:
   ```bash
   # Example commit structure
   git add path/to/updated/transport.go
   git commit -m "Update HTTP transport to use DialContext and add modern connection parameters"
   
   git add path/to/client/
   git commit -m "Update client package to use context-aware HTTP requests"
   ```

3. Update the tracking file with each completed change:
   ```bash
   git add modernization.md
   git commit -m "Update modernization tracking"
   ```

## 7. Final Verification

### Comprehensive Testing
```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...
```

### Documentation Update
Create a summary document (like `MODERNIZATION_SUMMARY.md`) that explains:
- What changes were made
- Why they were necessary
- Benefits of the modernization
- Any breaking changes and migration notes

## 8. OpenHands Efficiency Tips

1. **Use Batch Operations**:
   - Combine multiple grep commands into one search
   - Use find with exec to perform operations on multiple files

2. **Minimize Tool Calls**:
   - Combine multiple bash commands with `&&` or `;`
   - Use sed for batch file edits instead of opening each file

3. **Track Progress Explicitly**:
   - Update the tracking file regularly
   - Use git status to verify changes before committing

4. **Focus on High-Impact Changes First**:
   - Prioritize changes to core libraries and utilities
   - Address the most widely used deprecated APIs first

5. **Use Templates for Common Changes**:
   - Create templates for common modernization patterns
   - Apply them consistently across the codebase

By following these structured instructions, you can modernize Go repositories efficiently with OpenHands, minimizing costs and avoiding unnecessary work.