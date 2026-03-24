package main

// Real RSS feeds for tech news
func getDefaultFeeds() []string {
	return []string{
		// Tech News
		"https://hnrss.org/frontpage",                    // Hacker News front page
		"https://feeds.arstechnica.com/arstechnica/index", // Ars Technica
		"https://www.theverge.com/rss/index.xml",         // The Verge
		"https://techcrunch.com/feed/",                   // TechCrunch

		// Programming & Dev
		"https://blog.golang.org/feed.atom",              // Go Blog
		"https://www.reddit.com/r/golang/.rss",           // r/golang
		"https://www.reddit.com/r/programming/.rss",      // r/programming

		// AI & ML
		"https://www.reddit.com/r/MachineLearning/.rss",  // r/MachineLearning
		"https://www.reddit.com/r/artificial/.rss",       // r/artificial

		// DevOps & Cloud
		"https://www.reddit.com/r/kubernetes/.rss",       // r/kubernetes
		"https://www.reddit.com/r/devops/.rss",           // r/devops
	}
}
