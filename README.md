# RSS Digest

Daily RSS feed summarizer powered by Z.AI GLM Coding Plan (glm-4.7-flash).

## What it does

- Fetches multiple RSS feeds across news, tech, finance, science, and more
- Summarizes them into a concise daily digest using GLM-4.7-Flash
- Groups related stories, highlights what matters, skips duplicates
- Shows clickable source links for each article
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
| `GROQ_API_KEY` | (required) | Your Z.AI GLM Coding Plan API key (legacy name) |
| `DATA_DIR` | `/data` | Storage directory for digest and config |
| `PORT` | `8080` | Server port |

## API Configuration

This app uses the **Z.AI GLM Coding Plan** for summarization:

- **Endpoint:** `https://api.z.ai/api/coding/paas/v4/chat/completions`
- **Model:** `glm-4.7-flash`
- **Get API key:** https://z.ai/subscribe

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Web frontend |
| GET | `/api/digest` | Get cached digest (includes articles array) |
| POST | `/api/refresh` | Regenerate digest |
| GET | `/api/feeds` | List all configured feeds |
| POST | `/api/feeds/add` | Add new feed |
| POST | `/api/feeds/toggle` | Enable/disable feed |
| POST | `/api/feeds/remove` | Remove feed |
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

### 3. Create config with feeds

Create `config.json` in your data directory:

```json
{
  "feeds": [
    {"url": "https://hnrss.org/frontpage", "name": "Hacker News", "enabled": true},
    {"url": "https://feeds.bbci.co.uk/news/world/rss.xml", "name": "BBC World", "enabled": true}
  ]
}
```

### 4. Deploy

```bash
kubectl apply -k k8s/
```

This deploys:
- **Deployment** - Web server (1 replica)
- **CronJob** - Daily digest generation at 7 UTC (8 CET)
- **PVC** - Persistent storage for digest and config
- **Service** - ClusterIP service
- **Ingress** - https://rss.germainleignel.com

### 5. Verify

```bash
kubectl get pods -l app=rss-digest
kubectl logs -l app=rss-digest
```

## Default Feeds (49 total)

**Global News:** Reuters, AP, BBC, Al Jazeera, NPR, FT, Guardian, Economist, DW, France 24, Le Monde, NYT, WSJ, Fox

**France:** Franceinfo, Le Figaro, Le Monde, Libération

**Cybersecurity:** Schneier, The Hacker News, Krebs on Security

**AI:** MIT Tech Review AI, OpenAI Blog

**Finance:** Seeking Alpha, FT, WSJ Personal Finance/Economy/Markets/Business

**Tech:** Ars Technica, Hacker News, TechCrunch, The Verge, Wired

**Dev:** CSS-Tricks, Joel on Software, Martin Fowler, Mozilla Hacks, Node Weekly, NVIDIA Dev, Signal, Smashing

**Science:** NASA, Nature, Scientific American

**Startups:** TechCrunch Startups, YC Blog, Crunchbase News

Config stored in `/data/config.json`.

## Features

### Digest Cards
- Parsed into topic-based sections
- Staggered animations on load
- Markdown-style formatting

### Sources Section
- Shows top 50 articles with clickable links
- Article title + source name
- Opens in new tab

### Feed Management
- Toggle feeds on/off via UI
- Add/remove feeds dynamically
- Persists to config.json

## Requirements

- Go 1.23+ (for building)
- Docker (for deployment)
- Z.AI GLM Coding Plan API key

## License

MIT
