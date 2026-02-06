import { getUserWebSocketStatus } from "./users.js"
import { startChatWithUser } from "./chat-ui.js"

// Initialize global state
if (!window.userOrderOverrides) window.userOrderOverrides = new Map();
if (!window.unreadCounts) window.unreadCounts = new Map();

// Test function to verify notifications work
window.testNotification = function(username) {
    addUnreadMessage(username);
};
// Test function to simulate message from another user
window.simulateMessage = function(fromUser, message) {
    // Only add notification if not from active chat
    if (fromUser !== window.activeChatUsername) {
        addUnreadMessage(fromUser);
        znotif(fromUser);
    }
    updateTotalUnreadBadge();
};

// Quick test function
window.testNotificationPopup = function() {
    znotif('TestUser');
};

// Update the users list in the UI
export function updateUsersList(data) {
    const usersList = document.getElementById('chatUsersList');
    const usersCount = document.getElementById('onlineUsersCount');
    if (!usersList) return;

    let users = [];
    if (Array.isArray(data)) {
        users = data;
    } else if (data && Array.isArray(data.users)) {
        users = data.users;
    } else if (data && data.Users && Array.isArray(data.Users)) {
        users = data.Users;
    }

    // Count online users
    const onlineCount = users.filter(u => u.isOnline).length;
    if (usersCount) usersCount.textContent = onlineCount || 0;

    usersList.innerHTML = '';

    if (users.length === 0) {
        usersList.innerHTML = `<div class="user-item"><span>No users found</span></div>`;
        return;
    }

    // Sort: prioritize recent activity overrides, then by last_message_at, then alphabetically
    users.sort((a, b) => {
        const aOverride = window.userOrderOverrides.get(a.username);
        const bOverride = window.userOrderOverrides.get(b.username);
        
        // If both have recent overrides, sort by override timestamp
        if (aOverride && bOverride) {
            return bOverride.timestamp - aOverride.timestamp;
        }
        // Recent override always comes first
        if (aOverride && !bOverride) return -1;
        if (!aOverride && bOverride) return 1;
        
        // Fall back to original sorting logic
        const aHasMsg = a.last_message_at && a.last_message_at !== '';
        const bHasMsg = b.last_message_at && b.last_message_at !== '';
        
        if (aHasMsg && bHasMsg) {
            return new Date(b.last_message_at) - new Date(a.last_message_at);
        }
        if (aHasMsg && !bHasMsg) return -1;
        if (!aHasMsg && bHasMsg) return 1;
        return (a.username || '').localeCompare(b.username || '');
    });

    users.forEach(user => {
        const username = user.username || 'Unknown';
        const userId = user.id || '';
        const initial = username.charAt(0).toUpperCase();
        const statusClass = user.isOnline ? 'online' : 'offline';

        const item = document.createElement('div');
        item.className = 'user-item';
        item.dataset.userId = userId;
        item.dataset.username = username;

        // Check for recent activity override first
        const override = window.userOrderOverrides.get(user.username);
        const lastMessage = override ? override.lastMessage : user.last_message;
        const lastMessagePreview = lastMessage ? 
            `<div class="user-last-message">${lastMessage.length > 30 ? lastMessage.substring(0, 30) + '...' : lastMessage}</div>` : '';
        
        const unreadCount = window.unreadCounts.get(username) || 0;
        
        // Add notification class if there are unread messages
        if (unreadCount > 0) {
            item.classList.add('has-unread');
        }
        
        item.innerHTML = `
            <div class="initial"><span>${initial}</span></div>
            <div class="user-info">
                <span class="user-name">${username}</span>
                <span class="user-status ${statusClass}">●</span>
                ${lastMessagePreview}
            </div>
        `;
        
        // Add unread badge after creating the element
        if (unreadCount > 0) {
            const userNameEl = item.querySelector('.user-name');
            const badge = document.createElement('span');
            badge.className = 'unread-badge';
            badge.textContent = unreadCount;
            userNameEl.appendChild(badge);
        }
        item.addEventListener('click', () => startChatWithUser(username));
        usersList.appendChild(item);
    });
}

// Update WebSocket status indicator
export function updateWebSocketStatus() {
    const statusIndicator = document.getElementById('wsStatus');
    if (!statusIndicator) return;
    
    const status = getUserWebSocketStatus();
    
    switch (status) {
        case 'connected':
            statusIndicator.textContent = '🟢';
            statusIndicator.title = 'Connected';
            break;
        case 'connecting':
            statusIndicator.textContent = '🟡';
            statusIndicator.title = 'Connecting...';
            break;
        case 'disconnected':
            statusIndicator.textContent = '🔴';
            statusIndicator.title = 'Disconnected';
            break;
        default:
            statusIndicator.textContent = '⚪';
            statusIndicator.title = 'Unknown';
    }
}

export function updateUserListOrder(username, lastMessage) {
    const usersList = document.getElementById('chatUsersList');
    if (!usersList) return;
    
    // Track this user's recent activity
    window.userOrderOverrides.set(username, {
        lastMessage: lastMessage,
        timestamp: Date.now()
    });
    
    const userItem = Array.from(usersList.children).find(item => 
        item.dataset.username === username
    );
    
    if (userItem) {
        // Update last message preview
        const lastMessageEl = userItem.querySelector('.user-last-message');
        const preview = lastMessage.length > 30 ? lastMessage.substring(0, 30) + '...' : lastMessage;
        
        if (lastMessageEl) {
            lastMessageEl.textContent = preview;
        } else {
            const userInfo = userItem.querySelector('.user-info');
            const newLastMessage = document.createElement('div');
            newLastMessage.className = 'user-last-message';
            newLastMessage.textContent = preview;
            userInfo.appendChild(newLastMessage);
        }
        
        // Move to top
        usersList.insertBefore(userItem, usersList.firstChild);
    }
}

export function addUnreadMessage(senderName) {
    // Initialize unreadCounts if not exists
    if (!window.unreadCounts) window.unreadCounts = new Map();
    // Skip if currently chatting with this person
    if (senderName === window.activeChatUsername) {
        return;
    }
    const current = window.unreadCounts.get(senderName) || 0;
    const newCount = current + 1;
    window.unreadCounts.set(senderName, newCount);
    // Immediately update UI
    updateUserBadge(senderName);
    updateTotalUnreadBadge();
}

function updateUserBadge(username) {
    const usersList = document.getElementById('chatUsersList');
    if (!usersList) return;

    const userItem = [...usersList.querySelectorAll('.user-item')]
        .find(item => item.dataset.username === username);
    if (!userItem) return;

    const userNameEl = userItem.querySelector('.user-name');
    if (!userNameEl) return;

    let badge = userNameEl.querySelector('.unread-badge');
    const count = window.unreadCounts?.get(username) || 0;

    if (count > 0) {
        if (!badge) {
            badge = document.createElement('span');
            badge.className = 'unread-badge';
            userNameEl.appendChild(badge);
        }
        badge.textContent = count;
        userItem.classList.add('has-unread');
    } else {
        if (badge) badge.remove();
        userItem.classList.remove('has-unread');
    }
}

export function updateTotalUnreadBadge() {
    const totalBadge = document.getElementById('floatingUnreadBadge');
    if (!totalBadge) return;

    let totalUnread = 0;
    if (window.unreadCounts) {
        for (const count of window.unreadCounts.values()) {
            totalUnread += count;
        }
    }
    if (totalUnread > 0) {
        totalBadge.textContent = totalUnread > 9 ? '+9' : totalUnread;
        totalBadge.style.display = 'inline';
    } else {
        totalBadge.style.display = 'none';
    }
}


export function showNotification(username) {
    // Only keep badge notifications - all popup and browser notifications commented out
    
    // if (!window.chatWebSocket || window.chatWebSocket.readyState !== WebSocket.OPEN) {
    //     return;
    // }
    
    // Create popup - commented out
    // const popup = document.createElement('div');
    // popup.style.cssText = 'position:fixed;top:20px;right:20px;background:white;padding:15px;border-radius:8px;box-shadow:0 4px 12px rgba(0,0,0,0.3);z-index:99999;border-left:4px solid #2563eb';
    // popup.innerHTML = `<div style="display:flex;align-items:center;gap:10px"><span style="font-size:20px">💬</span><div><strong>${username}</strong><br><small>New message</small></div><button onclick="this.parentElement.parentElement.remove()" style="background:none;border:none;font-size:18px;cursor:pointer">×</button></div>`;
    // document.body.appendChild(popup);
    // setTimeout(() => popup.remove(), 4000);
    
    // Desktop notification - commented out
    // if (Notification.permission === 'granted') {
    //     new Notification(`New message from ${username}`);
    // } else if (Notification.permission !== 'denied') {
    //     Notification.requestPermission();
    // }
}

window.testNotification = () => showNotification('TestUser');
