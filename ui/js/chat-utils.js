export function throttle(func, delay) {
    let timeoutId;
    let lastExecTime = 0;
    return function (...args) {
        const currentTime = Date.now();
        if (currentTime - lastExecTime > delay) {
            func.apply(this, args);
            lastExecTime = currentTime;
        } else {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => {
                func.apply(this, args);
                lastExecTime = Date.now();
            }, delay - (currentTime - lastExecTime));
        }
    };
}

// Format message date
export function formatMessageDate(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    const messageDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    
    const timeStr = date.toLocaleTimeString('en-US', { 
        hour: 'numeric', 
        minute: '2-digit',
        hour12: true 
    });
    
    if (messageDate.getTime() === today.getTime()) {
        return `Today at ${timeStr}`;
    } else if (messageDate.getTime() === yesterday.getTime()) {
        return `Yesterday at ${timeStr}`;
    } else if (date.getFullYear() === now.getFullYear()) {
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric' 
        }) + ` at ${timeStr}`;
    } else {
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric',
            year: 'numeric'
        }) + ` at ${timeStr}`;
    }
}

// Initialize global state variables
export function initializeGlobalState() {
    // Store active chat state
    window.activeChatUsername = null;
    window.activeChatUserId = null;
    window.chatWebSocket = null;
    window.currentUserId = null;
    window.currentUsername = null;
    window.lastReceivedMessage = null;
    window.typingTimer = null;
    window.isTyping = false;
    window.userOrderOverrides = new Map();
    window.unreadCounts = new Map();
    window.chatReconnectAttempts = 0;
    // Debug flag to enable verbose logging during development
    if (typeof window.CHAT_DEBUG === 'undefined') window.CHAT_DEBUG = false;
}

// Call initialization
initializeGlobalState();