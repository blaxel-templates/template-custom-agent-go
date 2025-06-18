# Template Custom Agent Golang

A powerful Go-based AI agent API with multi-MCP server support, built with Gin framework. This API provides OpenAI-compatible endpoints with advanced tool calling capabilities and intelligent agent orchestration.

## 🚀 Features

- **Multi-MCP Server Support**: Connect to unlimited MCP (Model Context Protocol) servers
- **Intelligent Agent Loop**: Classic agent architecture with tool calling and iterative reasoning
- **OpenAI-Compatible API**: Full compatibility with OpenAI chat completions format
- **Tool Routing**: Automatic routing of tool calls to appropriate MCP servers
- **Health Monitoring**: Kubernetes-ready health checks with readiness and liveness probes
- **Modular Architecture**: Clean separation of concerns with organized router structure
- **Fault Tolerance**: Graceful handling of server failures and tool errors

## 🏗️ Architecture

```
Template
├── Agent Loop Engine
│   ├── Tool Manager (OpenAI format conversion)
│   ├── Server Routing (tool-to-server mapping)
│   └── Iterative Execution
├── MCP Manager
│   ├── Server 1 (blaxel-search)
│   ├── Server 2 (weather-service)
│   └── Server N (custom-tools)
└── Router Package
    ├── Health Routes (/health/*)
    ├── Tool Routes (/tools/*)
    ├── Agent Routes (/agent/*)
    └── Chat Routes (/chat, /v1/*)
```

## 📡 API Endpoints

### Health Monitoring
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness probe (checks MCP server availability)
- `GET /health/live` - Liveness probe

### Tool Management
- `GET /tools` - List all tools from all MCP servers
- `GET /tools/servers` - List all connected MCP servers
- `GET /tools/servers/:server/tools` - List tools from specific server

### Agent Execution
- `POST /` - Stream agent response as plain text (streaming)
- `POST /agent` - Run intelligent agent with tool calling (JSON response)
- `POST /agent/run` - Alternative agent endpoint

### Chat Completions
- `POST /v1/chat/completions` - OpenAI-compatible chat completions
- `POST /chat` - Simple chat interface

### Documentation
- `GET /` - API documentation and endpoint overview

### Root Endpoints
- `GET /` - API documentation and endpoint overview
- `POST /` - Stream agent response as plain text

## 🚀 Quick Start

### Local Development

1. **Clone and setup**:
   ```bash
   git clone https://github.com/blaxel-templates/template-custom-agent-go.git
   cd template-custom-agent-go
   go mod tidy
   make dependencies
   ```

3. **Run the server**:
   ```bash
   bl serve --hotreload
   ```

### Deployment

```bash
bl deploy
```

## 📖 Usage Examples

### Streaming Agent (Text Response)
```bash
curl -X POST http://localhost:1338/ \
  -H "Content-Type: application/json" \
  -d '{
    "inputs": "What is the weather in San Francisco?",
    "max_iterations": 5,
    "model": "sandbox-openai",
    "system_prompt": "You are a helpful weather assistant."
  }'
```

### Agent with Tool Calling (JSON Response)
```bash
curl -X POST http://localhost:1338/agent \
  -H "Content-Type: application/json" \
  -d '{
    "inputs": "What is the weather in San Francisco?",
    "max_iterations": 5,
    "model": "sandbox-openai",
    "system_prompt": "You are a helpful weather assistant."
  }'
```

### List Available Tools
```bash
curl http://localhost:1338/tools
```

### Check MCP Servers
```bash
curl http://localhost:1338/tools/servers
```

### OpenAI-Compatible Chat
```bash
curl -X POST http://localhost:1338/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Explain quantum computing"
      }
    ],
    "tools": []
  }'
```

### Health Checks
```bash
# Basic health
curl http://localhost:1338/health

# Readiness probe
curl http://localhost:1338/health/ready

# Liveness probe
curl http://localhost:1338/health/live
```

## 🏗️ Project Structure

```
template-custom-agent-go/
├── main.go                    # Application entry point
├── pkg/
│   ├── agent/                 # Agent orchestration
│   │   ├── agent.go          # Agent loop implementation
│   │   └── tool_manager.go   # MCP-to-OpenAI tool conversion
│   ├── blaxel/               # Blaxel client and MCP management
│   │   ├── client.go         # Main Blaxel client
│   │   ├── mcp_manager.go    # Multi-MCP server manager
│   │   ├── mcp.go           # MCP client implementation
│   │   └── transport.go      # WebSocket transport
│   ├── middleware/           # HTTP middleware
│   │   └── middleware.go     # Logging, recovery, error handling
│   └── router/               # HTTP route organization
│       ├── router.go         # Main router setup
│       ├── health.go         # Health check routes
│       ├── tools.go          # Tool management routes
│       ├── agent.go          # Agent execution routes
│       └── chat.go           # Chat completion routes
├── go.mod                    # Go module definition
├── go.sum                    # Go dependencies
├── Dockerfile               # Docker build configuration
└── README.md               # This documentation
```

## 🔄 Agent Loop Flow

1. **User Input**: Receive user query via API
2. **Tool Discovery**: Aggregate tools from all MCP servers
3. **Agent Initialization**: Create agent with tools and configuration
4. **Iterative Processing**:
   - Send request to AI model with available tools
   - If AI requests tool usage, execute tools via appropriate MCP server
   - Add tool results to conversation context
   - Continue until task completion or max iterations
5. **Response**: Return OpenAI-compatible response

## 🛡️ Error Handling

- **MCP Server Failures**: Graceful degradation, other servers remain functional
- **Tool Execution Errors**: Detailed error messages returned to agent
- **Network Issues**: Automatic retry logic and timeout handling
- **Validation Errors**: Clear error messages for malformed requests

## 🔍 Monitoring

### Health Endpoints
- `/health` - Basic service health
- `/health/ready` - Checks MCP server connectivity
- `/health/live` - Service liveness indicator

### Logging
- Structured logging with request tracing
- Tool execution logging
- MCP server connection status
- Error tracking and debugging

## 🚀 Advanced Features

### Multi-Server Tool Routing
Tools are automatically routed to the correct MCP server based on tool name mapping.

### Configurable Agent Parameters
- Custom system prompts
- Adjustable iteration limits
- Model selection
- Temperature and other parameters

### OpenAI Compatibility
Full compatibility with OpenAI chat completions API, including:
- Tool calling format
- Message structure
- Response format
- Error handling

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For issues and questions:
1. Check the health endpoints for system status
2. Review logs for error details
3. Verify MCP server connectivity
4. Ensure environment variables are properly set