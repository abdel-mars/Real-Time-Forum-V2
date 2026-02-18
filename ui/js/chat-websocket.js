import { displayMessage, scrollToBottom, showTypingIndicator, hideTypingIndicator, messageOffset } from "./chat-ui.js"
import { updateUserListOrder, addUnreadMessage, updateTotalUnreadBadge, showNotification } from "./chat-users.js"

// WebSocket state
let chatReconnectAttempts = 0;
const MAX_CHAT_RECONNECT_ATTEMPTS = 5;

// Set up WebSocket for chat
export function setupChatWebSocket(user1, user2) {
    try {
        console.log('[WebSocket] setupChatWebSocket called with user1:', user1, 'user2:', user2);
        console.log('[WebSocket] Current window.activeChatUsername:', window.activeChatUsername);
        console.log('[WebSocket] Current window.currentUsername:', window.currentUsername);

        // Close existing chat WebSocket if any
        if (window.chatWebSocket) {
            try {
                console.log('[WebSocket] Closing existing WebSocket');
                window.chatWebSocket.close();
            } catch (e) {
                console.error('Error closing existing WebSocket:', e);
            }
            window.chatWebSocket = null;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.hostname;
        const port = window.location.port ? `:${window.location.port}` : '';
        const wsUrl = `${protocol}//${host}${port}/room?user1=${encodeURIComponent(user1)}&user2=${encodeURIComponent(user2)}&username=${encodeURIComponent(user1)}`;

        console.log('[WebSocket] Full URL:', wsUrl);
        console.log('[WebSocket] Connecting to room: user1=' + user1 + ', user2=' + user2);

        try {
            window.chatWebSocket = new WebSocket(wsUrl);
        } catch (error) {
            console.error('[WebSocket] Failed to create WebSocket:', error);
            return;
        }
        // Generate a unique token for this connection to avoid stale handlers
        const socketToken = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
        window.currentChatSocketToken = socketToken;

        window.chatWebSocket.addEventListener('open', () => {
            console.log('[WebSocket] Connection opened');
            chatReconnectAttempts = 0;
        });

        window.chatWebSocket.addEventListener('message', (event) => {
            // Ignore events from stale sockets
            if (socketToken !== window.currentChatSocketToken) {
                console.log('[WebSocket] Ignoring message from stale socket');
                return;
            }
            try {
                console.log('[WebSocket] Message received:', event.data);
                const data = JSON.parse(event.data);

                if (data.type === 'message') {
                    const isSent = data.name === window.currentUsername;
                    console.log("Received message:", data);
                    console.log("Is sent:", isSent);

                    try {
                        hideTypingIndicator();

                        // Always include the ID from the server when available
                        const message = {
                            id: data.id || data.ID,
                            message: data.message,
                            sender_id: isSent ? window.currentUserId : window.activeChatUserId,
                            receiver_id: isSent ? window.activeChatUserId : window.currentUserId,
                            created_at: data.created_at || new Date().toISOString()
                        };

                        // Display message for both sent and received messages
                        // For received messages, only display if from active chat
                        if (!isSent && data.name === window.activeChatUsername) {
                            displayMessage(message, isSent);
                            scrollToBottom();
                        } else if (isSent) {
                            // For sent messages, always update the user list order
                            // The local display is handled in sendMessage()
                            updateUserListOrder(window.activeChatUsername, data.message);
                        }

                        if (!isSent) {
                            addUnreadMessage(data.name);
                            showNotification(data.name);
                            updateUserListOrder(data.name, data.message);
                        }

                        updateTotalUnreadBadge();
                    } catch (msgError) {
                        console.error('[WebSocket] Error processing message type:', msgError);
                    }
                } else if (data.type === 'typing') {
                    try {
                        if (data.name !== window.currentUsername && data.name === window.activeChatUsername) {
                            // Clear any existing safety timeout
                            if (window.typingSafetyTimeout) {
                                clearTimeout(window.typingSafetyTimeout);
                                window.typingSafetyTimeout = null;
                            }

                            showTypingIndicator();

                            // Set a safety timeout to hide the indicator if we don't get another 'typing' 
                            // or 'stop_typing' event within 5 seconds (providing a buffer over the 1.5s heartbeat)
                            window.typingSafetyTimeout = setTimeout(() => {
                                hideTypingIndicator();
                                console.log('[WebSocket] Hiding typing indicator due to safety timeout');
                            }, 5000);
                        }
                    } catch (typingError) {
                        console.error('[WebSocket] Error showing typing indicator:', typingError);
                    }
                } else if (data.type === 'stop_typing') {
                    try {
                        if (data.name !== window.currentUsername && data.name === window.activeChatUsername) {
                            // Clear safety timeout as we have an explicit stop
                            if (window.typingSafetyTimeout) {
                                clearTimeout(window.typingSafetyTimeout);
                                window.typingSafetyTimeout = null;
                            }
                            hideTypingIndicator();
                        }
                    } catch (stopTypingError) {
                        console.error('[WebSocket] Error hiding typing indicator:', stopTypingError);
                    }
                }
            } catch (error) {
                console.error('[WebSocket] Error processing message:', error, event.data);
            }
        });

        window.chatWebSocket.addEventListener('error', (error) => {
            console.error('[WebSocket] Error:', error);
        });

        window.chatWebSocket.addEventListener('close', (event) => {
            console.log(`[WebSocket] Connection closed. Code: ${event.code}, Reason: ${event.reason}, Clean: ${event.wasClean}`);
            window.chatWebSocket = null;

            // Clear typing indicator on disconnect to prevent zombie typing state
            hideTypingIndicator();

            // Only reconnect if this socket is still the active one for the same peer
            if (
                socketToken === window.currentChatSocketToken &&
                window.activeChatUsername === user2 &&
                window.currentUsername === user1 &&
                chatReconnectAttempts < MAX_CHAT_RECONNECT_ATTEMPTS
            ) {
                chatReconnectAttempts++;
                const delay = Math.min(1000 * (2 ** (chatReconnectAttempts - 1)), 5000);
                console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${chatReconnectAttempts})`);

                setTimeout(() => {
                    // Re-check the token and active user before reconnecting
                    if (
                        socketToken === window.currentChatSocketToken &&
                        window.activeChatUsername === user2 &&
                        window.currentUsername === user1
                    ) {
                        setupChatWebSocket(user1, user2);
                    }
                }, delay);
            } else if (chatReconnectAttempts >= MAX_CHAT_RECONNECT_ATTEMPTS) {
                alert('Lost connection to chat. Please refresh the page.');
            }
        });
    } catch (setupError) {
        console.error('[WebSocket] Fatal error in setupChatWebSocket:', setupError);
    }
}

// Typing indicator functions
export function handleTyping() {
    if (!window.chatWebSocket || window.chatWebSocket.readyState !== WebSocket.OPEN) {
        return;
    }

    const now = Date.now();
    const heartbeatInterval = 1500; // Send "typing" frequently (1.5s) to keep it alive

    // Send typing status if:
    // 1. We are starting a new typing session (!window.isTyping)
    // 2. We are resuming typing after a pause where stop_typing sent (!window.isTyping)
    // 3. We are typing continuously and need a heartbeat (now - last > interval)
    if (!window.isTyping || !window.lastTypingSentTime || (now - window.lastTypingSentTime > heartbeatInterval)) {
        window.isTyping = true;
        try {
            window.chatWebSocket.send(JSON.stringify({
                type: 'typing',
                name: window.currentUsername
            }));
            window.lastTypingSentTime = now;
            console.log(`[WebSocket] Sent typing heartbeat at ${now}`);
        } catch (e) {
            console.error('[WebSocket] Error sending typing status:', e);
            return;
        }
    }

    // Reset the "stop typing" timer - keep typing active
    clearTimeout(window.typingTimer);
    window.typingTimer = setTimeout(() => {
        if (window.chatWebSocket && window.chatWebSocket.readyState === WebSocket.OPEN) {
            window.isTyping = false;
            // Reset last sent time so next key press triggers immediate send
            window.lastTypingSentTime = 0;
            try {
                window.chatWebSocket.send(JSON.stringify({
                    type: 'stop_typing',
                    name: window.currentUsername
                }));
                console.log(`[WebSocket] Sent stop_typing at ${Date.now()}`);
            } catch (e) {
                console.error('[WebSocket] Error sending stop_typing:', e);
            }
        }
    }, 1000); // 1 second of inactivity triggers stop
}

// Send message
export function sendMessage() {
    const chatInput = document.getElementById('chatMessageInput');
    if (!chatInput) return;

    const message = chatInput.value.trim();
    if (!message || !window.activeChatUsername) {
        console.log('[sendMessage] Cannot send: message empty or activeChatUsername not set');
        return;
    }

    console.log('[sendMessage] Sending message to:', window.activeChatUsername);
    console.log('[sendMessage] Current user:', window.currentUsername);
    console.log('[sendMessage] WebSocket state:', window.chatWebSocket?.readyState);
    console.log('[sendMessage] Input value BEFORE clear:', chatInput.value);

    // Immediately stop typing and clear input to prevent conflicts
    window.isTyping = false;
    clearTimeout(window.typingTimer);

    // Force clear input
    chatInput.value = '';

    // Double check clear
    console.log('[sendMessage] Input value AFTER clear:', chatInput.value);

    // Use requestAnimationFrame to ensure UI update sticks if there's some weird race
    requestAnimationFrame(() => {
        if (document.getElementById('chatMessageInput')) {
            document.getElementById('chatMessageInput').value = '';
        }
    });

    if (!window.chatWebSocket || window.chatWebSocket.readyState !== WebSocket.OPEN) {
        console.error('[sendMessage] WebSocket not open! State:', window.chatWebSocket?.readyState);
        chatInput.value = message;
        return;
    }

    const messageData = {
        type: 'message',
        name: window.currentUsername,
        message: message,
        sender_id: window.currentUserId
    };

    console.log('[sendMessage] Message data:', messageData);

    // FIRST: Try to send via WebSocket
    try {
        window.chatWebSocket.send(JSON.stringify(messageData));
    } catch (error) {
        console.error('[sendMessage] Failed to send via WebSocket:', error);
        // Only restore input if the actual network send failed
        chatInput.value = message;
        return;
    }

    // SECOND: Update UI (if this fails, do NOT restore input since message was sent)
    try {
        // Display the message locally for the sender immediately
        // Use a temporary negative ID to mark this as a local unsaved message
        // The actual ID will come from the server response
        const localMessage = {
            id: -Date.now(), // Temporary negative ID to identify this as a local message
            message: message,
            sender_id: window.currentUserId,
            receiver_id: window.activeChatUserId,
            created_at: new Date().toISOString()
        };
        displayMessage(localMessage, true);
        scrollToBottom();
        updateUserListOrder(window.activeChatUsername, message);
        messageOffset++;
    } catch (error) {
        console.error('[sendMessage] Error updating UI after send:', error);
        // Message was sent successfully, so we don't restore the input
        // just log the UI error
    }
}

