# RSS Digest

Daily RSS feed summarizer powered by Groq (Llama 3.3 70B).

## What it does

- Fetches multiple RSS feeds (tech news, programming, AI, DevOps)
- Summarizes them into a concise daily digest using Groq's LLM
- Groups related stories, highlights what matters, skips duplicates

## Setup

1. Get a Groq API key from [console.groq.com](https://console.groq.com)
2. Set environment variable:
   ```bash
   export GROQ_API_KEY=your_key_here
   ```

## Run

```bash
go run .
```

Or build and run:
```bash
go build -o rss-digest .
./rss-digest
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

- Go 1.21+
- Groq API key

## License

MIT
