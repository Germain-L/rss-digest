const API_BASE = '';

const elements = {
    refreshBtn: document.getElementById('refreshBtn'),
    btnText: document.querySelector('.btn-text'),
    btnLoading: document.querySelector('.btn-loading'),
    status: document.getElementById('status'),
    digest: document.getElementById('digest'),
    lastUpdate: document.getElementById('lastUpdate'),
    feedsList: document.getElementById('feedsList'),
    addFeedForm: document.getElementById('addFeedForm'),
    feedUrl: document.getElementById('feedUrl'),
    feedName: document.getElementById('feedName')
};

// Tab switching
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        tab.classList.add('active');
        document.getElementById('tab-' + tab.dataset.tab).classList.add('active');
    });
});

// Parse simple markdown-like text to HTML
function parseContent(text) {
    if (!text) return '<p>No content</p>';

    let html = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/^### (.+)$/gm, '<h3>$1</h3>')
        .replace(/^## (.+)$/gm, '<h3>$1</h3>')
        .replace(/^# (.+)$/gm, '<h3>$1</h3>')
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/^\* (.+)$/gm, '<li>$1</li>')
        .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
        .replace(/^([^<\n].+)$/gm, '<p>$1</p>');

    return html;
}

function formatDate(date) {
    return new Intl.DateTimeFormat('en-US', {
        dateStyle: 'medium',
        timeStyle: 'short'
    }).format(date);
}

function setStatus(type, message) {
    elements.status.className = 'status ' + type;
    elements.status.textContent = message;
    elements.status.style.display = type ? 'block' : 'none';
}

function setLoading(loading) {
    elements.refreshBtn.disabled = loading;
    elements.btnText.style.display = loading ? 'none' : 'inline';
    elements.btnLoading.style.display = loading ? 'inline' : 'none';
}

// Digest functions
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
    } catch (err) {
        setStatus('error', 'Error: ' + err.message);
    } finally {
        setLoading(false);
    }
}

function displayDigest(data) {
    if (!data || !data.content) {
        elements.digest.innerHTML = '<div class="placeholder"><p>No digest available</p></div>';
        elements.lastUpdate.textContent = '';
        return;
    }

    elements.digest.innerHTML = parseContent(data.content);

    if (data.timestamp) {
        const date = new Date(data.timestamp);
        elements.lastUpdate.textContent = 'Last updated: ' + formatDate(date);
    }
}

// Feeds functions
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

function displayFeeds(feeds) {
    if (!feeds || feeds.length === 0) {
        elements.feedsList.innerHTML = '<p class="loading">No feeds configured</p>';
        return;
    }

    elements.feedsList.innerHTML = feeds.map(feed => `
        <div class="feed-item ${feed.enabled ? '' : 'disabled'}">
            <div class="feed-toggle ${feed.enabled ? 'active' : ''}" 
                 data-url="${feed.url}" 
                 data-enabled="${feed.enabled}"
                 onclick="toggleFeed(this)"></div>
            <div class="feed-info">
                <div class="feed-name">${escapeHtml(feed.name)}</div>
                <div class="feed-url">${escapeHtml(feed.url)}</div>
            </div>
            <button class="feed-remove" onclick="removeFeed('${escapeHtml(feed.url)}')" title="Remove feed">×</button>
        </div>
    `).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

async function toggleFeed(el) {
    const url = el.dataset.url;
    const enabled = el.dataset.enabled === 'true';
    const newEnabled = !enabled;

    try {
        const res = await fetch(API_BASE + '/api/feeds/toggle', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url, enabled: newEnabled })
        });

        if (!res.ok) throw new Error('Failed to toggle feed');

        el.dataset.enabled = newEnabled;
        el.classList.toggle('active', newEnabled);
        el.closest('.feed-item').classList.toggle('disabled', !newEnabled);
    } catch (err) {
        console.error('Error toggling feed:', err);
        alert('Failed to toggle feed');
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
    } catch (err) {
        console.error('Error removing feed:', err);
        alert('Failed to remove feed');
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
    } catch (err) {
        console.error('Error adding feed:', err);
        alert('Failed to add feed: ' + err.message);
    }
}

// Event listeners
elements.refreshBtn.addEventListener('click', refreshDigest);
elements.addFeedForm.addEventListener('submit', addFeed);

// Initial load
fetchDigest().then(displayDigest);
fetchFeeds().then(displayFeeds);
