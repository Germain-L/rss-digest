// RSS Digest Frontend
// ===================

const API_BASE = '';

// DOM Elements
const elements = {
    refreshBtn: document.getElementById('refreshBtn'),
    copyBtn: document.getElementById('copyBtn'),
    status: document.getElementById('status'),
    digest: document.getElementById('digest'),
    lastUpdate: document.getElementById('lastUpdate'),
    itemCount: document.getElementById('itemCount'),
    feedsList: document.getElementById('feedsList'),
    feedCount: document.getElementById('feedCount'),
    addFeedForm: document.getElementById('addFeedForm'),
    feedUrl: document.getElementById('feedUrl'),
    feedName: document.getElementById('feedName'),
    toastContainer: document.getElementById('toastContainer')
};

// State
let currentDigest = null;
let isLoading = false;

// ==================
// Tab Navigation
// ==================

document.querySelectorAll('.nav-tab').forEach(tab => {
    tab.addEventListener('click', () => {
        document.querySelectorAll('.nav-tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        tab.classList.add('active');
        document.getElementById('tab-' + tab.dataset.tab).classList.add('active');
    });
});

// ==================
// Toast Notifications
// ==================

function showToast(message, type = 'success', duration = 3000) {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icon = type === 'success' ? '✓' : type === 'error' ? '✕' : 'ℹ';
    toast.innerHTML = `<span>${icon}</span><span>${escapeHtml(message)}</span>`;
    
    elements.toastContainer.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 200);
    }, duration);
}

// ==================
// Status Display
// ==================

function setStatus(type, message) {
    if (!type) {
        elements.status.className = 'status-toast';
        elements.status.style.display = 'none';
        elements.status.innerHTML = '';
        return;
    }
    
    const icon = type === 'loading' ? '⏳' : type === 'error' ? '✕' : '✓';
    elements.status.className = `status-toast ${type}`;
    elements.status.innerHTML = `<span>${icon}</span><span>${escapeHtml(message)}</span>`;
    elements.status.style.display = 'flex';
}

// ==================
// Loading State
// ==================

function setLoading(loading) {
    isLoading = loading;
    elements.refreshBtn.disabled = loading;
    
    if (loading) {
        elements.refreshBtn.classList.add('loading');
    } else {
        elements.refreshBtn.classList.remove('loading');
    }
}

// ==================
// Date Formatting
// ==================

function formatDate(date) {
    return new Intl.DateTimeFormat('en-US', {
        dateStyle: 'medium',
        timeStyle: 'short'
    }).format(date);
}

function formatRelative(date) {
    const now = new Date();
    const diff = now - date;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    
    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return formatDate(date);
}

// ==================
// HTML Escape
// ==================

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ==================
// Digest Parsing
// ==================

function parseDigestToCards(content) {
    if (!content) return [];
    
    // Split by markdown headers (## or ###)
    const sections = content.split(/^(?:#{2,3})\s+/m).filter(Boolean);
    const cards = [];
    
    for (const section of sections) {
        const lines = section.trim().split('\n');
        const title = lines[0].trim();
        const body = lines.slice(1).join('\n').trim();
        
        if (title && body) {
            cards.push({
                title: cleanTitle(title),
                content: parseContent(body)
            });
        }
    }
    
    // If no sections found, treat entire content as one card
    if (cards.length === 0 && content.trim()) {
        cards.push({
            title: 'Daily Digest',
            content: parseContent(content)
        });
    }
    
    return cards;
}

function cleanTitle(title) {
    // Remove markdown formatting
    return title.replace(/[*_`#]/g, '').trim();
}

function parseContent(text) {
    if (!text) return '';
    
    let html = text
        // Escape HTML
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        // Bold
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        // Italic (used for source attribution)
        .replace(/\*(.+?)\*/g, '<em>$1</em>')
        // Source attribution in parentheses - make it subtle
        .replace(/\(([^)]+)\)/g, '<em>($1)</em>')
        // List items
        .replace(/^[•\-\*]\s+(.+)$/gm, '<li>$1</li>')
        // Wrap consecutive list items in ul
        .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
        // Paragraphs (lines not already wrapped)
        .replace(/^(?!<[luo]l|<li|<em|<strong)(.+)$/gm, '<p>$1</p>');
    
    return html;
}

// ==================
// Digest Display
// ==================

function displayDigest(data) {
    currentDigest = data;
    
    if (!data || !data.content) {
        elements.digest.innerHTML = `
            <div class="placeholder-state">
                <div class="placeholder-icon">📭</div>
                <h3>No digest available</h3>
                <p>Click "Refresh" to fetch and summarize the latest news</p>
            </div>
        `;
        elements.lastUpdate.textContent = '';
        elements.itemCount.textContent = '';
        return;
    }
    
    const cards = parseDigestToCards(data.content);
    
    if (cards.length === 0) {
        elements.digest.innerHTML = `
            <div class="placeholder-state">
                <div class="placeholder-icon">🤔</div>
                <h3>Couldn't parse digest</h3>
                <p>Try refreshing again</p>
            </div>
        `;
        return;
    }
    
    // Build digest cards HTML
    let html = cards.map((card, i) => `
        <article class="digest-card">
            <div class="card-header">
                <h3 class="card-title">${escapeHtml(card.title)}</h3>
                ${i === 0 ? '<span class="card-badge">Latest</span>' : ''}
            </div>
            <div class="card-content">
                ${card.content}
            </div>
        </article>
    `).join('');
    
    // Add sources section if articles exist
    if (data.articles && data.articles.length > 0) {
        html += `
            <section class="sources-section">
                <h2 class="sources-title">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
                        <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
                    </svg>
                    Sources
                </h2>
                <div class="sources-grid">
                    ${data.articles.map(article => `
                        <a href="${escapeHtml(article.link)}" target="_blank" rel="noopener noreferrer" class="source-link">
                            <span class="source-title">${escapeHtml(article.title)}</span>
                            <span class="source-name">${escapeHtml(article.source)}</span>
                        </a>
                    `).join('')}
                </div>
            </section>
        `;
    }
    
    elements.digest.innerHTML = html;
    
    // Update metadata
    if (data.timestamp) {
        const date = new Date(data.timestamp);
        elements.lastUpdate.textContent = formatRelative(date);
        elements.lastUpdate.title = formatDate(date);
    }
    
    if (data.itemCount) {
        elements.itemCount.textContent = `${data.itemCount} items`;
    }
}

function showSkeleton() {
    elements.digest.innerHTML = `
        <div class="skeleton-loader">
            <div class="skeleton-card">
                <div class="skeleton skeleton-title"></div>
                <div class="skeleton skeleton-text"></div>
                <div class="skeleton skeleton-text"></div>
                <div class="skeleton skeleton-text short"></div>
            </div>
            <div class="skeleton-card">
                <div class="skeleton skeleton-title"></div>
                <div class="skeleton skeleton-text"></div>
                <div class="skeleton skeleton-text"></div>
            </div>
            <div class="skeleton-card">
                <div class="skeleton skeleton-title"></div>
                <div class="skeleton skeleton-text"></div>
            </div>
        </div>
    `;
}

// ==================
// API Calls
// ==================

async function fetchDigest() {
    try {
        const res = await fetch(API_BASE + '/api/digest');
        if (!res.ok) throw new Error('Failed to fetch digest');
        return await res.json();
    } catch (err) {
        console.error('Error fetching digest:', err);
        return null;
    }
}

async function refreshDigest() {
    if (isLoading) return;
    
    setLoading(true);
    setStatus('loading', 'Fetching feeds and generating summary...');
    
    try {
        const res = await fetch(API_BASE + '/api/refresh', { method: 'POST' });
        
        if (!res.ok) {
            const err = await res.text();
            throw new Error(err || 'Failed to refresh');
        }
        
        const data = await res.json();
        displayDigest(data);
        setStatus('');
        showToast('Digest updated successfully', 'success');
    } catch (err) {
        setStatus('error', err.message);
        showToast('Failed to refresh digest', 'error');
    } finally {
        setLoading(false);
    }
}

async function fetchFeeds() {
    try {
        const res = await fetch(API_BASE + '/api/feeds');
        if (!res.ok) throw new Error('Failed to fetch feeds');
        return await res.json();
    } catch (err) {
        console.error('Error fetching feeds:', err);
        return [];
    }
}

// ==================
// Feeds Display
// ==================

function displayFeeds(feeds) {
    const activeCount = feeds.filter(f => f.enabled).length;
    elements.feedCount.textContent = `${activeCount} of ${feeds.length} active`;
    
    if (!feeds || feeds.length === 0) {
        elements.feedsList.innerHTML = `
            <div class="feeds-empty">
                <p>No feeds configured yet</p>
                <p>Add an RSS feed above to get started</p>
            </div>
        `;
        return;
    }
    
    elements.feedsList.innerHTML = feeds.map(feed => `
        <div class="feed-item ${feed.enabled ? '' : 'disabled'}">
            <div class="feed-toggle ${feed.enabled ? 'active' : ''}" 
                 data-url="${escapeHtml(feed.url)}"
                 onclick="toggleFeed(this)"></div>
            <div class="feed-info">
                <div class="feed-name">${escapeHtml(feed.name)}</div>
                <div class="feed-url">${escapeHtml(feed.url)}</div>
            </div>
            <button class="feed-remove" onclick="removeFeed('${escapeHtml(feed.url)}')" title="Remove feed" aria-label="Remove feed">
                ×
            </button>
        </div>
    `).join('');
}

// ==================
// Feed Actions
// ==================

async function toggleFeed(el) {
    const url = el.dataset.url;
    const isActive = el.classList.contains('active');
    const newEnabled = !isActive;
    
    // Optimistic update
    el.classList.toggle('active', newEnabled);
    el.closest('.feed-item').classList.toggle('disabled', !newEnabled);
    
    try {
        const res = await fetch(API_BASE + '/api/feeds/toggle', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url, enabled: newEnabled })
        });
        
        if (!res.ok) throw new Error('Failed to toggle feed');
        
        // Update feed count
        const feeds = await fetchFeeds();
        const activeCount = feeds.filter(f => f.enabled).length;
        elements.feedCount.textContent = `${activeCount} of ${feeds.length} active`;
        
        showToast(newEnabled ? 'Feed enabled' : 'Feed disabled', 'success');
    } catch (err) {
        // Revert on error
        el.classList.toggle('active', isActive);
        el.closest('.feed-item').classList.toggle('disabled', !isActive);
        showToast('Failed to toggle feed', 'error');
    }
}

async function removeFeed(url) {
    if (!confirm('Remove this feed?')) return;
    
    try {
        const res = await fetch(API_BASE + '/api/feeds/remove', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url })
        });
        
        if (!res.ok) throw new Error('Failed to remove feed');
        
        const feeds = await fetchFeeds();
        displayFeeds(feeds);
        showToast('Feed removed', 'success');
    } catch (err) {
        showToast('Failed to remove feed', 'error');
    }
}

async function addFeed(e) {
    e.preventDefault();
    
    const url = elements.feedUrl.value.trim();
    const name = elements.feedName.value.trim() || url;
    
    if (!url) return;
    
    try {
        const res = await fetch(API_BASE + '/api/feeds/add', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url, name })
        });
        
        if (!res.ok) throw new Error('Failed to add feed');
        
        elements.feedUrl.value = '';
        elements.feedName.value = '';
        
        const feeds = await fetchFeeds();
        displayFeeds(feeds);
        showToast('Feed added successfully', 'success');
    } catch (err) {
        showToast('Failed to add feed: ' + err.message, 'error');
    }
}

// ==================
// Copy Digest
// ==================

async function copyDigest() {
    if (!currentDigest || !currentDigest.content) {
        showToast('No digest to copy', 'error');
        return;
    }
    
    try {
        await navigator.clipboard.writeText(currentDigest.content);
        showToast('Digest copied to clipboard', 'success');
    } catch (err) {
        showToast('Failed to copy', 'error');
    }
}

// ==================
// Event Listeners
// ==================

elements.refreshBtn.addEventListener('click', refreshDigest);
elements.addFeedForm.addEventListener('submit', addFeed);
elements.copyBtn.addEventListener('click', copyDigest);

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
    // Ctrl/Cmd + R to refresh (when not in input)
    if ((e.ctrlKey || e.metaKey) && e.key === 'r' && document.activeElement.tagName !== 'INPUT') {
        e.preventDefault();
        refreshDigest();
    }
});

// ==================
// Initial Load
// ==================

async function init() {
    // Load digest
    const digest = await fetchDigest();
    if (digest) {
        displayDigest(digest);
    } else {
        showSkeleton();
    }
    
    // Load feeds
    const feeds = await fetchFeeds();
    displayFeeds(feeds);
}

init();
