# Local Cloud Gaming

Open-source GeForce Now alternative for local hosting (residential complexes, LAN parties, small communities).

## Overview

A management layer on top of [Sunshine](https://github.com/LizardByte/Sunshine) that enables multi-GPU, multi-user game streaming from a shared Linux server.

### Target Clients
- Steam Deck / SteamOS
- Bazzite (Fedora-based gaming distro)
- Android handhelds (ROG Ally, Lenovo Legion Go with Android, etc.)

Uses [Moonlight](https://moonlight-stream.org/) clients for streaming.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Local Cloud Gaming                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Sunshine   │  │  Sunshine   │  │  Sunshine   │   ...   │
│  │   (GPU 0)   │  │   (GPU 1)   │  │   (GPU 2)   │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
│         │                │                │                 │
│  ┌──────┴────────────────┴────────────────┴──────┐         │
│  │              Orchestrator (Go)                 │         │
│  │  - GPU pool management                         │         │
│  │  - Session allocation                          │         │
│  │  - User authentication                         │         │
│  │  - Queue management                            │         │
│  └──────────────────┬────────────────────────────┘         │
│                     │                                       │
│  ┌──────────────────┴────────────────────────────┐         │
│  │           Admin Dashboard (Web)                │         │
│  └────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
                          │
            ┌─────────────┼─────────────┐
            ▼             ▼             ▼
      ┌──────────┐  ┌──────────┐  ┌──────────┐
      │ SteamOS  │  │ Bazzite  │  │ Android  │
      │(Moonlight)│ │(Moonlight)│ │(Moonlight)│
      └──────────┘  └──────────┘  └──────────┘
```

## Features

### MVP
- [ ] Docker Compose deployment with multiple Sunshine instances
- [ ] GPU assignment API (claim/release)
- [ ] Simple web dashboard for status
- [ ] User PIN authentication (Moonlight-compatible)

### Future
- [ ] Session time limits and scheduling
- [ ] Queue system for busy periods
- [ ] Game library management
- [ ] Usage analytics
- [ ] Wake-on-LAN for power management

## Requirements

### Server
- Linux (Ubuntu 22.04+ recommended)
- NVIDIA GPU(s) with driver 535+
- Docker with NVIDIA Container Toolkit

### Network
- Gigabit LAN minimum
- Low latency (<5ms to clients)
- Ports: 47984-48010 per GPU instance

## Quick Start

```bash
# Clone
git clone https://github.com/zhiganov/local-cloud-gaming
cd local-cloud-gaming

# Configure GPUs
cp .env.example .env
# Edit .env with your GPU count

# Start
docker compose up -d

# Access dashboard
open http://localhost:8080
```

## Development

```bash
# Backend (Go)
cd orchestrator
go run .

# Dashboard
cd dashboard
npm run dev
```

## License

MIT
