# go-diff
[![GoDoc](https://godoc.org/github.com/jakebowkett/go-diff/diff?status.svg)](https://godoc.org/github.com/jakebowkett/go-diff/diff)

Package diff provides a simple way to diff objects. It
is intended to assist with tasks such as logging changes
to config structs in a running program. Please see the
godoc link above for documentation.

```go
type Config struct {
    Debug   bool
    Version string
    Timeout int
}

c1 := Config{}
c2 := Config{"0.0.1", true, 30}

changes, _ := diff.Objects(c1, c2)

fmt.Println(changes[0]) // `.Debug changed from false to true`
fmt.Println(changes[1]) // `.Version changed from "" to "0.0.1"`
fmt.Println(changes[2]) // `.Timeout changed from 0 to 30`
```
