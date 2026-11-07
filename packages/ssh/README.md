# @opentui/ssh
SSH server integration for OpenTUI Go

## Installation

```sh
go get PixTui/packages/ssh
```

## Usage

```go
package main

import (
	"PixTui/packages/ssh"
)

func main() {
	server := ssh.NewServer("0.0.0.0:2222", func(session *ssh.Session) {
		// Handle SSH session
	})
	server.Listen()
	defer server.Close()
}
```
