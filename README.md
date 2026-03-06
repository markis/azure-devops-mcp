# Azure DevOps MCP Server

[![CI](https://github.com/markis/azure-devops-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/markis/azure-devops-mcp/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-52.1%25-yellow.svg)](https://github.com/markis/azure-devops-mcp)
[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://go.dev/dl/)

MCP (Model Context Protocol) server for Azure DevOps work item management. Allows AI assistants to interact with Azure DevOps work items through a standardized protocol.

## Features

- ✅ Get work items by ID
- ✅ List work items with WIQL queries
- ✅ List work items assigned to authenticated user
- ✅ Create new work items
- ✅ Update existing work items
- ✅ Add comments to work items
- ✅ Full field support (story points, acceptance criteria, area/iteration paths, etc.)
- ✅ HTML to Markdown conversion for descriptions

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/markis/azure-devops-mcp/releases).

#### macOS

```bash
# Intel Mac
curl -LO https://github.com/markis/azure-devops-mcp/releases/latest/download/azure-devops-mcp-darwin-amd64
chmod +x azure-devops-mcp-darwin-amd64
mv azure-devops-mcp-darwin-amd64 /usr/local/bin/azure-devops-mcp

# Apple Silicon (M1/M2/M3)
curl -LO https://github.com/markis/azure-devops-mcp/releases/latest/download/azure-devops-mcp-darwin-arm64
chmod +x azure-devops-mcp-darwin-arm64
mv azure-devops-mcp-darwin-arm64 /usr/local/bin/azure-devops-mcp
```

#### Linux

```bash
# AMD64
curl -LO https://github.com/markis/azure-devops-mcp/releases/latest/download/azure-devops-mcp-linux-amd64
chmod +x azure-devops-mcp-linux-amd64
sudo mv azure-devops-mcp-linux-amd64 /usr/local/bin/azure-devops-mcp

# ARM64
curl -LO https://github.com/markis/azure-devops-mcp/releases/latest/download/azure-devops-mcp-linux-arm64
chmod +x azure-devops-mcp-linux-arm64
sudo mv azure-devops-mcp-linux-arm64 /usr/local/bin/azure-devops-mcp
```

#### Windows

Download the appropriate `.exe` file:
- `azure-devops-mcp-windows-amd64.exe` (64-bit Intel/AMD)
- `azure-devops-mcp-windows-arm64.exe` (ARM64)

Add the binary to your PATH.

### Build from Source

```bash
git clone https://github.com/markis/azure-devops-mcp.git
cd azure-devops-mcp
go build -o azure-devops-mcp ./cmd/azure-devops-mcp
```

## Configuration

Set the following environment variables:

```bash
export AZURE_DEVOPS_ORG_URL="https://dev.azure.com/your-org"
export AZURE_DEVOPS_PAT="your-personal-access-token"
export AZURE_DEVOPS_PROJECT="your-project-name"
```

### Creating a Personal Access Token (PAT)

1. Go to Azure DevOps → User Settings → Personal Access Tokens
2. Create a new token with **Work Items (Read & Write)** scope
3. Copy the token (you won't see it again!)

## Usage

Start the MCP server:

```bash
azure-devops-mcp
```

The server runs over stdio and follows the [Model Context Protocol](https://modelcontextprotocol.io) specification.

### Configuring MCP Clients

#### Claude Desktop

Add to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "azure-devops": {
      "command": "/usr/local/bin/azure-devops-mcp",
      "env": {
        "AZURE_DEVOPS_ORG_URL": "https://dev.azure.com/your-org",
        "AZURE_DEVOPS_PAT": "your-personal-access-token",
        "AZURE_DEVOPS_PROJECT": "your-project-name"
      }
    }
  }
}
```

#### Cline (VS Code Extension)

Add to your Cline MCP settings file:

**macOS/Linux**: `~/.config/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`
**Windows**: `%APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json`

```json
{
  "mcpServers": {
    "azure-devops": {
      "command": "/usr/local/bin/azure-devops-mcp",
      "env": {
        "AZURE_DEVOPS_ORG_URL": "https://dev.azure.com/your-org",
        "AZURE_DEVOPS_PAT": "your-personal-access-token",
        "AZURE_DEVOPS_PROJECT": "your-project-name"
      }
    }
  }
}
```

#### OpenCode

Add to your OpenCode MCP configuration:

**macOS/Linux**: `~/.config/opencode/mcp.json`

```json
{
  "mcpServers": {
    "azure-devops": {
      "command": "/usr/local/bin/azure-devops-mcp",
      "env": {
        "AZURE_DEVOPS_ORG_URL": "https://dev.azure.com/your-org",
        "AZURE_DEVOPS_PAT": "your-personal-access-token",
        "AZURE_DEVOPS_PROJECT": "your-project-name"
      }
    }
  }
}
```

**Note**: After adding the configuration, restart your MCP client for the changes to take effect.

### Available Tools

- `get_work_item` - Fetch a single work item by ID
- `list_work_items` - Run a WIQL query and return matching work items
- `list_my_work_items` - Return active work items assigned to you
- `create_work_item` - Create a new work item
- `update_work_item` - Update an existing work item
- `add_comment` - Add a comment to a work item

## Development

### Prerequisites

- Go 1.26+
- golangci-lint (for linting)

### Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run linter
golangci-lint run

# Build
go build -o azure-devops-mcp ./cmd/azure-devops-mcp
```

### Project Structure

```
.
├── cmd/azure-devops-mcp/   # Entry point
├── internal/
│   ├── client/             # Azure DevOps API client
│   ├── controller/         # MCP server setup and tool registration
│   └── tools/              # Business logic for MCP tools
├── .github/workflows/      # CI/CD pipelines
└── AGENTS.md               # Guidelines for AI coding agents
```

## Test Coverage

Current coverage: **52.1%**

- `internal/tools`: 96.8% (business logic)
- `internal/client`: 51.2% (API helpers)
- `internal/controller`: 36.5% (MCP wiring)

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `go test ./...`
2. Linter passes: `golangci-lint run`
3. Code coverage is maintained or improved
4. Follow conventional commit messages

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Uses [Azure DevOps Go API](https://github.com/microsoft/azure-devops-go-api)
- HTML to Markdown conversion by [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown)
