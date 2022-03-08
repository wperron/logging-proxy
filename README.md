# Logging Proxy

This is a small application meant to be used as part of the development cycle as
an easy way to inspect outbound HTTP requests from an application.

For any sufficiently complex application, it's not hard to imagine there be
multiple layers of libraries all composed together just to send an HTTP request
to an external service. Sticking this proxy in between the application and the
downstream server allows for an easy inspection of HTTP requests sent along with
their headers and body.

## Usage

```bash
# listens on `-addr` and proxies calls to `-proxy`
go run main.go -addr :3098 -proxy http://my-service:8080
```

