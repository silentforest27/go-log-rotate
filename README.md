# go-log-rotate

`go-log-rotate` is a simple, thread-safe log rotation library for Go. It ensures that log files do not grow indefinitely by rotating them once they reach a specified size limit.

## Usage

```go
rotator, err := logrotate.NewRotator("app.log", 10*1024*1024) // 10MB
if err != nil {
    log.Fatal(err)
}

// Use it with the standard log package
log.SetOutput(rotator)
log.Println("This will be rotated!")
```