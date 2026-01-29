# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Local Cloud Gaming is a multi-GPU management layer for Sunshine game streaming. It enables shared gaming servers in residential complexes or small communities.

## Architecture

- **Orchestrator** (`/orchestrator`) - Go backend managing GPU allocation, sessions, and users
- **Dashboard** (`/dashboard`) - Web UI for admins (status, users, settings)
- **Docker** - Each GPU runs an isolated Sunshine container

## Commands

```bash
# Run orchestrator locally
cd orchestrator && go run .

# Run tests
cd orchestrator && go test ./...

# Build orchestrator
cd orchestrator && go build -o bin/orchestrator

# Start full stack
docker compose up -d

# View logs
docker compose logs -f orchestrator
```

## Key Concepts

### GPU Pool
GPUs are managed as a pool. Each GPU can be assigned to one user at a time. The orchestrator tracks:
- Available GPUs
- Active sessions (user → GPU mapping)
- Session start time and limits

### Sunshine Integration
Each Sunshine instance runs in a container with GPU passthrough. The orchestrator:
1. Assigns a port range per instance
2. Generates pairing PINs
3. Monitors health via Sunshine API

### Client Flow
1. User opens Moonlight, enters server IP
2. Orchestrator assigns available GPU, returns connection details
3. User enters PIN to pair
4. Sunshine handles streaming directly to client

## Environment Variables

```bash
GPU_COUNT=2              # Number of GPUs to manage
BASE_PORT=47984          # Starting port for Sunshine instances
ADMIN_PASSWORD=xxx       # Dashboard admin password
SESSION_TIMEOUT=4h       # Auto-release after inactivity
```
