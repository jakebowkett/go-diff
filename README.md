# go-diff
[![GoDoc](https://godoc.org/github.com/jakebowkett/go-diff/diff?status.svg)](https://godoc.org/github.com/jakebowkett/go-diff/diff)

Package diff provides a simple way to diff objects. It
is intended to assist with tasks such as logging changes
to config structs in a running program. Please see the
godoc link above for documentation.

```go
type Config struct {
    Debug      bool
    Path       string
    timeout    int // unexported fields will be diffed
    EmailOnErr []string
}

c1 := Config{}
c2 := Config{
    true,
    "path/to/thing",
    30,
    []string{"person@domain.me"},
}

changes, _ := diff.Objects(c1, c2)

fmt.Println(changes[0]) // `.Debug changed from false to true`
fmt.Println(changes[1]) // `.Path changed from "" to "path/to/thing"`
fmt.Println(changes[2]) // `.timeout changed from 0 to 30`
fmt.Println(changes[3]) // `.EmailOnErr[0] added "person@domain.me"`
```
