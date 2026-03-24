# RSS Digest

Daily RSS feed summarizer powered by Groq (Llama 3.3 70B).

## What it does

- Fetches multiple RSS feeds (tech news, programming, AI, DevOps)
- Summarizes them into a concise daily digest using Groq's LLM
- Groups related stories, highlights what matters, skips duplicates
- Web UI + daily cron for automated updates

## Quick Start

```bash
export GROQ_API_KEY=your_key_here
./rss-digest              # CLI mode (print to stdout)
./rss-digest --serve      # Web server at http://localhost:8080
./rss-digest --generate   # Generate and save to storage
```

## Modes

| Mode | Command | Description |
|------|---------|-------------|
| CLI | `./rss-digest` | Print digest to stdout |
| Server | `./rss-digest --serve` | Start web server |
| Generate | `./rss-digest --generate` | Generate and save to storage (for cron) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GROQ_API_KEY` | (required) | Your Groq API key |
| `DATA_DIR` | `/data` | Storage directory for digest |
| `PORT` | `8080` | Server port |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Web frontend |
| GET | `/api/digest` | Get cached digest |
| POST | `/api/refresh` | Regenerate digest |
| GET | `/health` | Health check |

## Deployment (Kubernetes)

### 1. Build and push image

```bash
./build-and-push.sh
```

### 2. Create secret with API key

```bash
kubectl create secret generic rss-digest-secrets \
  --from-literal=GROQ_API_KEY=your_key_here \
  -n default
```

### 3. Deploy

```bash
kubectl apply -k k8s/
```

This deploys:
- **Deployment** - Web server (1 replica)
- **CronJob** - Daily digest generation at 7 UTC (8 CET)
- **PVC** - Persistent storage for digest
- **Service** - ClusterIP service
- **Ingress** - https://rss.germainleignel.com

### 4. Verify

```bash
kubectl get pods -l app=rss-digest
kubectl logs -l app=rss-digest
```

## Feeds

Default feeds include:
- Hacker News (front page)
- Ars Technica
- The Verge
- TechCrunch
- Go Blog
- r/golang, r/programming
- r/MachineLearning, r/artificial
- r/kubernetes, r/devops

Edit `feeds_list.go` to customize.

## Requirements

- Go 1.21+ (for building)
- Docker (for deployment)
- Groq API key

## License

MIT
