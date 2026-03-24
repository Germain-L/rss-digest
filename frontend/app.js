const API_BASE = '';

const elements = {
    refreshBtn: document.getElementById('refreshBtn'),
    btnText: document.querySelector('.btn-text'),
    btnLoading: document.querySelector('.btn-loading'),
    status: document.getElementById('status'),
    digest: document.getElementById('digest'),
    lastUpdate: document.getElementById('lastUpdate')
};

// Parse simple markdown-like text to HTML
function parseContent(text) {
    if (!text) return '<p>No content</p>';

    let html = text
        // Escape HTML
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        // Headers
        .replace(/^### (.+)$/gm, '<h3>$1</h3>')
        .replace(/^## (.+)$/gm, '<h3>$1</h3>')
        .replace(/^# (.+)$/gm, '<h3>$1</h3>')
        // Bold
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        // List items
        .replace(/^\* (.+)$/gm, '<li>$1</li>')
        // Wrap consecutive li elements in ul
        .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
        // Paragraphs (lines not already wrapped)
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

// Event listeners
elements.refreshBtn.addEventListener('click', refreshDigest);

// Load existing digest on page load
fetchDigest().then(displayDigest);
