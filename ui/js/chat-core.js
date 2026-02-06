import { loadUsersWithStatus, initUserWebSocket, closeUserWebSocket} from "./users.js"
import { updateUsersList, updateWebSocketStatus } from "./chat-users.js"

// Store interval IDs for cleanup
let wsStatusInterval = null;
let userRefreshInterval = null;

export function creatchatisland() {
    // Initialize global notification state
    if (!window.unreadCounts) window.unreadCounts = new Map();
    if (!window.userOrderOverrides) window.userOrderOverrides = new Map();
    // Remove existing chat island if any
    const existingChatIsland = document.querySelector('.chat-island');
    if (existingChatIsland) {
        existingChatIsland.remove();
    }
    const chatSection = document.createElement("div");
    chatSection.className = 'chat-island';
    chatSection.id = 'chatIsland';
    chatSection.innerHTML = `
        <div class="chat-island-container">
            <!-- Header -->
            <div class="chat-island-header">
                <span class="ws-status" id="wsStatus" title="WebSocket Status">🔴</span>
                <button class="chat-minimize-btn">−</button>
            </div>

            <!-- Users Section -->
            <div class="chat-users-section">
                <div class="chat-users-header">
                    <span>Online Users</span>
                    <span class="users-count" id="onlineUsersCount">0</span>
                </div>
                <div class="chat-users-list" id="chatUsersList">
                    <!-- Users will be loaded here dynamically -->
                    <div class="user-item">
                        <div class="initial"><span>U</span></div>
                        <div class="user-info">
                            <span class="user-name">Loading...</span>
                            <span class="user-status online">Online</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Chat Messages Section -->
            <div class="chat-messages-section">
                <div class="chat-messages-header">
                </div>
                <div class="chat-messages-container" id="chatMessagesContainer">
                    <!-- Messages will appear here -->
                    <div class="no-chat-selected">
                        <p>Select a user to chat with!</p>
                    </div>
                </div>
                <div class="chat-input-section">
                    <input 
                        type="text" 
                        id="chatMessageInput" 
                        placeholder="Type a message..." 
                        class="chat-input"
                        disabled
                    />
                    <button id="chatSendBtn" class="chat-send-btn" disabled>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    `;

    document.body.appendChild(chatSection);

    // Create floating chat icon
    const floatingIcon = document.createElement('div');
    floatingIcon.className = 'chat-floating';
    floatingIcon.innerHTML = `
        <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        </svg>
        <span id="floatingUnreadBadge" class="floating-unread-badge" style="display:none;"></span>
    `;
    document.body.appendChild(floatingIcon);

    // Initially hide chat island and show floating icon
    chatSection.classList.add('closed');
    floatingIcon.style.display = 'flex';

    // Set up minimize button with toggle functionality (closes to floating icon with smooth animation)
    const minimizeBtn = document.querySelector('.chat-minimize-btn');
    if (minimizeBtn) {
        minimizeBtn.addEventListener('click', () => {
            const chatIsland = document.getElementById('chatIsland');
            if (chatIsland) {
                // Start closing animation from current position
                chatIsland.classList.remove('open');
                chatIsland.classList.add('closed');

                // After animation completes, show floating icon
                setTimeout(() => {
                    floatingIcon.style.display = 'flex';
                }, 400);
            }
        });
    }

    // Set up users section collapse/expand functionality
    const usersHeader = chatSection.querySelector('.chat-users-header');
    if (usersHeader) {
        usersHeader.style.cursor = 'pointer';
        usersHeader.addEventListener('click', () => {
            const usersSection = chatSection.querySelector('.chat-users-section');
            const container = chatSection.querySelector('.chat-island-container');
            if (usersSection && container) {
                usersSection.classList.toggle('collapsed');
                container.classList.toggle('users-collapsed');
            }
        });
    }

    // Floating icon click to open chat
    floatingIcon.addEventListener('click', () => {
        floatingIcon.style.display = 'none';
        chatSection.classList.remove('closed');
        chatSection.classList.add('open');
    });
    
    // Initialize WebSocket for real-time user updates
    initUserWebSocket(updateUsersList);
    
    // Load initial users list
    loadUsersWithStatus().then(users => {
        if (users) {
            updateUsersList(users);
        }
    });
    
    // Get current user info for chat
    fetch('/api/me').then(res => {
        if (!res.ok) throw new Error('Not authenticated');
        return res.text();
    }).then(text => {
        try {
            const user = JSON.parse(text);
            if (user.authenticated) {
                window.currentUsername = user.Username || user.username;
                window.currentUserId = user.userId || user.Id || user.id || user.ID;
                setupGlobalNotifications();
            }
        } catch (e) {}
    }).catch(err => {});

    function setupGlobalNotifications() {
        // Initialize message deduplication cache
        if (!window.recentMessages) {
            window.recentMessages = new Map();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.hostname;
        const port = window.location.port || (protocol === 'wss:' ? '443' : '80');
        const wsUrl = `${protocol}//${host}:${port}/notifications?username=${encodeURIComponent(window.currentUsername)}`;

        window.globalWS = new WebSocket(wsUrl);

        window.globalWS.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'message' && data.name !== window.currentUsername) {
                    // Create unique message key for deduplication
                    const messageKey = `${data.name}_${data.created_at || data.message}`;
                    const now = Date.now();

                    // Check if we've seen this message recently (within last 5 seconds)
                    if (window.recentMessages.has(messageKey)) {
                        const lastSeen = window.recentMessages.get(messageKey);
                        if (now - lastSeen < 5000) {
                            console.log('Duplicate message ignored:', messageKey);
                            return; // Skip processing duplicate
                        }
                    }

                    // Mark message as seen
                    window.recentMessages.set(messageKey, now);

                    // Clean up old entries (older than 10 seconds)
                    for (const [key, timestamp] of window.recentMessages.entries()) {
                        if (now - timestamp > 10000) {
                            window.recentMessages.delete(key);
                        }
                    }

                    import('./chat-users.js').then(module => {
                        module.addUnreadMessage(data.name);
                        module.showNotification(data.name);
                        module.updateUserListOrder(data.name, data.message);
                        module.updateTotalUnreadBadge();
                    });
                }
            } catch (e) {}
        };
    }

    // Update WebSocket status indicator periodically
    wsStatusInterval = setInterval(updateWebSocketStatus, 10000);
    
    // Refresh user list periodically (reduced frequency to avoid rate limiting)
    userRefreshInterval = setInterval(() => {
        loadUsersWithStatus().then(users => {
            if (users) {
                updateUsersList(users);
            }
        }).catch(err => {});
    }, 120000); // Increased to 2 minutes
}

// Show the chat island
export function showChatIsland() {
    const chatIsland = document.getElementById('chatIsland');
    if (chatIsland) {
        chatIsland.classList.add('visible');

    }
}

// Hide the chat island
export function hideChatIsland() {
    const chatIsland = document.getElementById('chatIsland');
    if (chatIsland) {
        chatIsland.classList.remove('visible');

    }
}

// Clean up function to close WebSocket when chat island is removed
export function destroyChatIsland() {
    // Clear intervals
    if (wsStatusInterval) {
        clearInterval(wsStatusInterval);
        wsStatusInterval = null;
    }
    if (userRefreshInterval) {
        clearInterval(userRefreshInterval);
        userRefreshInterval = null;
    }
    
    // Clear typing timer
    if (window.typingTimer) {
        clearTimeout(window.typingTimer);
        window.typingTimer = null;
    }
    
    // Close chat WebSocket
    if (window.chatWebSocket) {
        try {
            window.chatWebSocket.close();
        } catch (e) {}
        window.chatWebSocket = null;
    }
    
    // Close user WebSocket
    closeUserWebSocket();
    
    // Reset state
    window.activeChatUsername = null;
    window.activeChatUserId = null;
    window.chatReconnectAttempts = 0;
    
    // Remove chat section
    const chatSection = document.querySelector('.chat-island');
    if (chatSection) {
        chatSection.remove();
    }
}