# capture [![Build Status](https://travis-ci.org/mssola/capture.svg?branch=master)](https://travis-ci.org/mssola/capture) [![GoDoc](https://godoc.org/github.com/mssola/capture?status.png)](http://godoc.org/github.com/mssola/capture)

This is a super simple package that basically allows you to capture the output
from os.Stdout and os.Stderr in a safe and clean manner. This is how you should
use it:

```go
package main

import (
	"fmt"
	"os"

	"github.com/mssola/capture"
)

func willPanic() {
	panic("Told ya!")
}

func foo() {
	fmt.Fprintf(os.Stdout, "Hello")
	fmt.Fprintf(os.Stderr, "World")
}

func main() {
	res := capture.All(func() { willPanic() })
	fmt.Printf("%v\n", res.Error)
	// Output: "Panic: Told ya!"

	res = capture.All(func() { foo() })
	fmt.Printf("%s %s\n", res.Stdout, res.Stderr)
	// Output: "Hello World"
}
```

Copyright &copy; 2015-2016 Miquel Sabaté Solà, released under the MIT License.
