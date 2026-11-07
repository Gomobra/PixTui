# PixTui SSH

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

SSH server integration for PixTui.

## About

PixTui SSH provides an SSH server implementation for PixTui applications. It supports password, public key, and keyboard-interactive authentication methods.

## Features

- **SSH server** — Listen on TCP port with graceful shutdown
- **Authentication** — Password, public key, keyboard-interactive, and no-auth methods
- **Host keys** — RSA, ECDSA, and Ed25519 key generation and loading
- **Authorized keys** — Load from file or configure programmatically
- **Session management** — PTY allocation, window resize, environment variables
- **Middleware** — Request/response middleware chain
- **Session limits** — Per-connection and global session limits
- **ANSI rendering** — Write frames to SSH sessions with SGR codes

## Installation

```bash
go get github.com/Gomobra/PixTui/packages/ssh
```

## Quick Start

```go
package main

import (
	"github.com/Gomobra/PixTui/packages/ssh"
)

func main() {
	server := ssh.NewServer("0.0.0.0:2222", func(session *ssh.Session) {
		// Handle SSH session
	})
	server.SetPasswordAuth(func(user, password string) bool {
		return user == "admin" && password == "secret"
	})
	server.Listen()
	defer server.Close()
}
```

## Authentication Methods

| Method | Function | Description |
|---|---|---|
| Password | `SetPasswordAuth` | Username/password authentication |
| Public Key | `SetPublicKeyAuth` | Public key authentication |
| Authorized Keys | `SetAuthorizedKeys` | Load authorized_keys file |
| Keyboard-Interactive | `SetKeyboardInteractiveAuth` | Interactive challenge/response |
| No Auth | `SetNoAuth` | No authentication required |

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)