// WebSocket connection for real-time user updates
let userWebSocket = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000;  

// Load all users via REST API
export async function loadAllUsers(){
    const apirout = '/api/all-users'
    try{
        const response = await fetch(apirout)
        if(!response.ok){
            throw new Error(`Response status: ${response.status}`);
        }
        const result = await response.json();
        return result;
    } catch (error) {
        console.error('Error loading all users:', error.message);
        return null;
    }
}

// Load online users via REST API
export async function loadOnlineUsers(){
    const apirout = '/api/online-users'
    try{
        const response = await fetch(apirout)
        if(!response.ok){
            throw new Error(`Response status: ${response.status}`);
        }
        const result = await response.json();
        return result;
    } catch (error) {
        console.error('Error loading online users:', error.message);
        return null;
    }
}

// Get current user info
export async function getCurrentUser(){
    try {
        const response = await fetch('/api/me');
        if(!response.ok){
            throw new Error(`Response status: ${response.status}`);
        }
        const result = await response.json();
        return result;
    } catch (error) {
        console.error('Error loading current user:', error.message);
        return null;
    }
}

// Rate limiting for API calls
let lastApiCall = 0;
const MIN_API_INTERVAL = 1000; 
// Load users with online status (excluding current user)
export async function loadUsersWithStatus(){
    try {
        // Rate limiting check
        const now = Date.now();
        const timeSinceLastCall = now - lastApiCall;
        if (timeSinceLastCall < MIN_API_INTERVAL) {
            const waitTime = MIN_API_INTERVAL - timeSinceLastCall;
            await new Promise(resolve => setTimeout(resolve, waitTime));
        }
        lastApiCall = Date.now();
        // Fetch current user first
        const currentUserData = await getCurrentUser();
        if (!currentUserData) return [];
        // Delay between requests
        await new Promise(resolve => setTimeout(resolve, 200));
        // Fetch all users
        const allUsersData = await loadAllUsers();
        if (!allUsersData) return [];
        
        // Delay between requests
        await new Promise(resolve => setTimeout(resolve, 200));
        
        // Fetch online users
        const onlineUsersData = await loadOnlineUsers();
        
        // Extract user arrays
        const currentUsername = currentUserData?.Username || currentUserData?.username || '';
        const allUsers = allUsersData?.users || [];
        const onlineUsers = onlineUsersData?.users || [];
        
        // Create a Set for faster lookup
        const onlineUsersSet = new Set(onlineUsers);
        
        // Map all users with their online status, excluding current user
        const usersWithStatus = allUsers
            .filter(username => username !== currentUsername)
            .map(username => ({
                username: username,
                isOnline: onlineUsersSet.has(username)
            }));
        
        return usersWithStatus;
    } catch (error) {
        console.error('Error loading users with status:', error);
        return [];
    }
}

// Initialize WebSocket connection for real-time user updates
export function initUserWebSocket(onUsersUpdate) {
    // Close existing connection if any
    if (userWebSocket) {
        userWebSocket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = window.location.port || (protocol === 'wss:' ? '443' : '80');
    const wsUrl = `${protocol}//${host}:${port}/room?room=users_lobby`;    
    try {
        userWebSocket = new WebSocket(wsUrl);
        
        userWebSocket.addEventListener('open', () => {
            reconnectAttempts = 0;
            
            // Request initial user list with status
            loadUsersWithStatus().then(users => {
                if (users && onUsersUpdate) {
                    onUsersUpdate(users);
                }
            }).catch(err => console.error('Error loading initial users:', err));
        });
        
        let lastRefresh = 0;
        const REFRESH_THROTTLE = 5000; // 5 seconds minimum between refreshes
        
        userWebSocket.addEventListener('message', (event) => {
            try {
                const data = JSON.parse(event.data);
                // Throttle user list refreshes to prevent rate limiting
                const now = Date.now();
                if (now - lastRefresh < REFRESH_THROTTLE) {
                    return;
                }
                // Handle different message types
                if (data.type === 'users_update' || data.type === 'user_joined' || data.type === 'user_left') {
                    lastRefresh = now;
                    // Add delay before refresh to prevent rate limiting
                    setTimeout(() => {
                        loadUsersWithStatus().then(users => {
                            if (users && onUsersUpdate) {
                                onUsersUpdate(users);
                            }
                        }).catch(err => {
                            console.log('Rate limited, skipping refresh');
                        });
                    }, 500);
                }
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        });
        
        userWebSocket.addEventListener('error', (error) => {
            console.error(' User WebSocket error:', error);
        });
        
        userWebSocket.addEventListener('close', (event) => {
            // Attempt to reconnect
            if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                reconnectAttempts++;
                setTimeout(() => {
                    initUserWebSocket(onUsersUpdate);
                }, RECONNECT_DELAY);
            } else {
                console.log(' Max reconnection attempts reached');
            }
        });
        
    } catch (error) {
        console.error('Failed to create WebSocket:', error);
    }
    
    return userWebSocket;
}

// Close WebSocket connection
export function closeUserWebSocket() {
    if (userWebSocket) {
        userWebSocket.close();
        userWebSocket = null;
    }
}

// Send a Message Through WebSocket !! \\ 
export function sendUserWebSocketMessage(message) {
    if (userWebSocket && userWebSocket.readyState === WebSocket.OPEN) {
        userWebSocket.send(JSON.stringify(message));
        return true;
    } else {
        console.warn('WebSocket is not connected');
        return false;
    }
}

export function getUserWebSocketStatus() {
    if (!userWebSocket) return 'disconnected';
    switch (userWebSocket.readyState) {
        case WebSocket.CONNECTING:
            return 'connecting';
        case WebSocket.OPEN:
            return 'connected';
        case WebSocket.CLOSING:
            return 'closing';
        case WebSocket.CLOSED:
            return 'disconnected';
        default:
            return 'unknown';
    }
}