import {
  setupChatWebSocket,
  handleTyping,
  sendMessage,
} from "./chat-websocket.js";
import { throttle } from "./chat-utils.js";
import { updateTotalUnreadBadge } from "./chat-users.js";

// UI state
export let messageOffset = 0;
let lastMessageDate = null;

// Maximum number of messages to keep visible in the chat window
const MAX_VISIBLE_MESSAGES = 10;

// Start a chat with a specific user
export async function startChatWithUser(username) {
  console.log('[Chat UI] startChatWithUser called with username:', username);
  console.log('[Chat UI] Current window.currentUsername:', window.currentUsername);
  console.log('[Chat UI] Previous window.activeChatUsername:', window.activeChatUsername);
  
  // FIRST: Close old WebSocket and clear old state
  if (window.chatWebSocket) {
    try {
      console.log('[Chat UI] Closing old WebSocket before switching to new user');
      window.chatWebSocket.close();
    } catch (e) {
      console.error('Error closing existing WebSocket:', e);
    }
    window.chatWebSocket = null;
  }
  
  // Clear message container and reset offset BEFORE changing active user
  const container = document.getElementById("chatMessagesContainer");
  if (container) {
    container.innerHTML = "";
  }
  messageOffset = 0;
  lastMessageDate = null;
  
  // Clear typing indicator when switching chats
  hideTypingIndicator();
  
  window.activeChatUsername = username;
  console.log('[Chat UI] Updated window.activeChatUsername to:', window.activeChatUsername);
  // Clear unread count for this user immediately
  if (!window.unreadCounts) window.unreadCounts = new Map();
  const oldCount = window.unreadCounts.get(username) || 0;
  window.unreadCounts.set(username, 0);
  // Update UI immediately
  const usersList = document.getElementById("chatUsersList");
  if (usersList) {
    const userItem = [...usersList.querySelectorAll(".user-item")].find(
      (item) => item.dataset.username === username
    );
    if (userItem) {
      const badge = userItem.querySelector(".unread-badge");
      if (badge) badge.remove();
      userItem.classList.remove("has-unread");
    }
  }
  updateTotalUnreadBadge();
  try {
    // Get current user info
    const currentUserRes = await fetch("/api/me");
    const currentUser = await currentUserRes.json();
    window.currentUsername = currentUser.Username || currentUser.username;
    window.currentUserId =
      currentUser.userId || currentUser.Id || currentUser.id || currentUser.ID;

    // Get target user ID
    const targetUserRes = await fetch(
      `/api/user-by-username?username=${encodeURIComponent(username)}`
    );
    if (!targetUserRes.ok) {
      console.error("Failed to get user info");
      return;
    }
    const targetUser = await targetUserRes.json();
    window.activeChatUserId = targetUser.id;

    // Update UI
    const activeChatUser = document.getElementById("activeChatUser");
    const chatInput = document.getElementById("chatMessageInput");
    const sendBtn = document.getElementById("chatSendBtn");

    if (activeChatUser) {
      activeChatUser.innerHTML = `
                <span>Chat with ${username}</span>
                <button class="close-chat-btn" onclick="closeChat()">✕</button>
            `;
    }

    if (chatInput) {
      chatInput.disabled = false;
      chatInput.placeholder = `Message ${username}...`;
      chatInput.value = "";
    }

    if (sendBtn) {
      sendBtn.disabled = false;
    }

    // Load chat history
    messageOffset = 0;
    await loadChatHistory(window.activeChatUserId, 10, 0);

    // Set up WebSocket for real-time messages (NEW connection for new user)
    setupChatWebSocket(window.currentUsername, username);

    // Set up send message handler
    setupSendMessageHandler();
  } catch (error) {
    console.error("Error starting chat:", error);
  }
}

// Load chat history
export async function loadChatHistory(otherUserId, limit = 10, offset = 0) {
  try {
    const res = await fetch(
      `/api/chat-messages?other_user_id=${otherUserId}&limit=${limit}&offset=${offset}`
    );
    if (!res.ok) return;

    const data = await res.json();
    console.log("Loaded chat history 10 :", data);
    const messages = data.messages || [];
    const container = document.getElementById("chatMessagesContainer");
    if (!container) return;

    // Reset on first load
    if (offset === 0) {
      container.innerHTML = "";
      messageOffset = 0;
      window.allMessagesLoaded = false;
    }

    if (messages.length < limit) {
      window.allMessagesLoaded = true;

    }

    // Show messages
    if (offset === 0) {
      // Initial load - normal order
      messages.forEach((m) => {
        const isSent = String(m.sender_id) === String(window.currentUserId);
        displayMessage(m, isSent);
      });
      scrollToBottom();
    } else {
      // Loading older messages - insert at top
      const scrollHeight = container.scrollHeight;
      // Insert older messages at the top (avoid duplicating the current first message)
      messages.reverse().forEach((m) => {
        const isSent = String(m.sender_id) === String(window.currentUserId);
        // Basic dedupe: if the top-most message already matches this one, skip it
        const firstMsgEl = container.querySelector('.message-row');
        if (firstMsgEl) {
          const firstText = firstMsgEl.querySelector('.message-text')?.textContent || '';
          const firstSender = firstMsgEl.querySelector('.message-sender')?.textContent || '';
          const incomingSender = isSent ? window.currentUsername : window.activeChatUsername;
          if (firstText === (m.message || '') && firstSender === incomingSender) {
            // skip duplicate
            return;
          }
        }

        const el = createMessageElement(m, isSent);
        container.insertBefore(el, container.firstChild);
      });
      container.scrollTop = container.scrollHeight - scrollHeight;
    }

    // Setup scroll listener only once
    if (!container.dataset.scrollHandler) {
      container.addEventListener(
        "scroll",
        throttle(() => {
          if (container.scrollTop === 0 && !window.allMessagesLoaded) {
            loadChatHistory(otherUserId, limit, messageOffset);
          }
        }, 500)
      );
      container.dataset.scrollHandler = "true";
    }

    // Update offset
    messageOffset += messages.length;
  } catch (e) {
    console.error("Error loading chat history:", e);
  }
}

// Display a message
export function displayMessage(message, isSent) {
  console.log(message);
  const container = document.getElementById("chatMessagesContainer");
  if (!container) return;

  // Deduplicate: if last message in the container has same sender and text, skip
  const lastMsgEl = [...container.querySelectorAll('.message-row')].pop();
  const incomingText = message.message || '';
  const incomingSender = isSent ? window.currentUsername : window.activeChatUsername;
  if (lastMsgEl) {
    const lastText = lastMsgEl.querySelector('.message-text')?.textContent || '';
    const lastSender = lastMsgEl.querySelector('.message-sender')?.textContent || '';
    const lastTime = lastMsgEl.querySelector('.message-time')?.textContent || '';
    // Compare text and sender; also allow slight timestamp differences
    if (lastText === incomingText && lastSender === incomingSender) {
      return; // skip duplicate
    }
  }

  maybeAddDateSeparator(container, message.created_at);

  const el = createMessageElement(message, isSent);
  container.appendChild(el);

  // Trim messages to keep only last MAX_VISIBLE_MESSAGES to avoid UI overflow
  trimVisibleMessages(container, MAX_VISIBLE_MESSAGES);
}

function trimVisibleMessages(container, max) {
  if (!container) return;
  const rows = Array.from(container.querySelectorAll('.message-row'));
  if (rows.length <= max) return;
  const toRemove = rows.length - max;
  for (let i = 0; i < toRemove; i++) {
    const el = rows[i];
    if (el && el.parentNode) el.parentNode.removeChild(el);
  }
}

export function createMessageElement(message, isSent) {
  const wrapper = document.createElement("div");
  wrapper.className = `message-row ${isSent ? "sent" : "received"}`;

  const bubble = document.createElement("div");
  bubble.className = `message-bubble ${
    isSent ? "bubble-sent" : "bubble-received"
  }`;

  // Show sender name (you or the other user)
  const name = document.createElement("div");
  name.className = "message-sender";
  name.textContent = isSent
    ? window.currentUsername
    : window.activeChatUsername;

  // Message text
  const text = document.createElement("div");
  text.className = "message-text";
  text.textContent = message.message;
  //    console.log('Creating message element for:', message.ID || message.id || message.message_id, 'Sent by:', isSent ? 'self' : 'other');
  // Full date + time
  const time = document.createElement("div");
  time.className = "message-time";
  // Handle both ISO 8601 format (string) and milliseconds (number)
  let d;
  if (typeof message.created_at === "string" && message.created_at.includes("T")) {
    // ISO 8601 format (e.g., "2025-01-15T10:30:45Z")
    d = new Date(message.created_at);
  } else {
    // Unix milliseconds (legacy/WebSocket format)
    d = new Date(parseInt(message.created_at));
  }
  time.textContent = d.toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
  console.log("Message time:", time.textContent);

  bubble.appendChild(name);
  bubble.appendChild(text);
  bubble.appendChild(time);
  wrapper.appendChild(bubble);

  return wrapper;
}

// Set up send message handler
function setupSendMessageHandler() {
  const input = document.getElementById("chatMessageInput");
  const btn = document.getElementById("chatSendBtn");

  // Remove old listeners first
  btn.replaceWith(btn.cloneNode(true));
  input.replaceWith(input.cloneNode(true));

  const newBtn = document.getElementById("chatSendBtn");
  const newInput = document.getElementById("chatMessageInput");

  newBtn.onclick = sendMessage;
  newInput.onkeypress = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  newInput.oninput = () => {
    handleTyping();
  };
}

export function showTypingIndicator() {
  let indicator = document.getElementById("typingIndicator");
  if (!indicator) {
    const container = document.getElementById("chatMessagesContainer");
    if (!container) return;

    indicator = document.createElement("div");
    indicator.id = "typingIndicator";
    indicator.className = "typing-indicator";
    indicator.innerHTML = `
            <div class="message-row received">
                <div class="message-bubble bubble-received">
                    <div class="message-sender">${window.activeChatUsername}</div>
                    <div class="typing-dots">
                        <span></span><span></span><span></span>
                    </div>
                </div>
            </div>
        `;
    container.appendChild(indicator);
    scrollToBottom();
  }
}

export function hideTypingIndicator() {
  const indicator = document.getElementById("typingIndicator");
  if (indicator) {
    indicator.remove();
  }
}

// Scroll to bottom of messages
export function scrollToBottom() {
  const messagesContainer = document.getElementById("chatMessagesContainer");
  if (messagesContainer) {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
  }
}

// Close chat
window.closeChat = function () {
  window.activeChatUsername = null;
  window.activeChatUserId = null;
  messageOffset = 0;
  window.chatReconnectAttempts = 0;

  // Clear typing state
  clearTimeout(window.typingTimer);
  window.isTyping = false;
  hideTypingIndicator();

  // Close WebSocket connection
  if (window.chatWebSocket) {
    try {
      window.chatWebSocket.close();
    } catch (e) {
      console.error("Error closing WebSocket:", e);
    }
    window.chatWebSocket = null;
  }

  const activeChatUser = document.getElementById("activeChatUser");
  const chatInput = document.getElementById("chatMessageInput");
  const sendBtn = document.getElementById("chatSendBtn");
  const messagesContainer = document.getElementById("chatMessagesContainer");

  if (activeChatUser) {
    activeChatUser.textContent = "Select a user to chat";
  }

  if (chatInput) {
    chatInput.disabled = true;
    chatInput.placeholder = "Type a message...";
    chatInput.value = "";
  }

  if (sendBtn) {
    sendBtn.disabled = true;
  }

  if (messagesContainer) {
    messagesContainer.innerHTML = `
            <div class="no-chat-selected">
                <p>Select a user from the list to start chatting</p>
            </div>
        `;
  }
};

function maybeAddDateSeparator(container, createdAt) {
  // Handle both ISO 8601 format (string) and milliseconds (number)
  let msgDate;
  if (typeof createdAt === "string" && createdAt.includes("T")) {
    // ISO 8601 format
    msgDate = new Date(createdAt);
  } else {
    // Unix milliseconds
    msgDate = new Date(parseInt(createdAt));
  }
  const dateKey = msgDate.toDateString();

  if (lastMessageDate !== dateKey) {
    lastMessageDate = dateKey;

    const separator = document.createElement("div");
    separator.className = "date-separator";

    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (msgDate >= today) {
      separator.textContent = "Today";
    } else if (msgDate >= yesterday) {
      separator.textContent = "Yesterday";
    } else {
      separator.textContent = msgDate.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year:
          msgDate.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
      });
    }

    container.appendChild(separator);
  }
}

function clearUnreadMessages(username) {
  if (!window.unreadCounts) window.unreadCounts = new Map();
  window.unreadCounts.set(username, 0);
  updateUserBadge(username);
  updateTotalUnreadBadge();
}

// Make chat window draggable - initialize after DOM loads
window.addEventListener("load", initChatDrag);

function initChatDrag() {
  const chatWindow = document.querySelector(".chat-island");
  const chatHeader = chatWindow
    ? chatWindow.querySelector(".chat-island-header")
    : null;

  if (!chatWindow || !chatHeader) {
    console.log("Chat elements not found, retrying in 1s...");
    setTimeout(initChatDrag, 1000); // Retry after 1 second
    return;
  }

  let isDragging = false;
  let offsetX, offsetY;

  // Restore saved position if on desktop/tablet and position exists
  if (window.innerWidth >= 768) {
    const savedPosition = localStorage.getItem("chatWindowPosition");
    if (savedPosition) {
      try {
        const pos = JSON.parse(savedPosition);
        // Validate position is within current viewport bounds
        const maxLeft = window.innerWidth - chatWindow.offsetWidth;
        const maxTop = window.innerHeight - chatWindow.offsetHeight;
        const validLeft = Math.max(0, Math.min(maxLeft, pos.left));
        const validTop = Math.max(0, Math.min(maxTop, pos.top));

        // Add restoring class for smooth transition
        chatWindow.classList.add("restoring");
        chatWindow.style.left = validLeft + "px";
        chatWindow.style.top = validTop + "px";
        chatWindow.style.right = "auto";
        chatWindow.style.bottom = "auto";

        // Remove restoring class after transition completes
        setTimeout(() => {
          chatWindow.classList.remove("restoring");
        }, 400);
      } catch (e) {
        console.warn("Invalid saved chat position:", e);
      }
    }
  }

  chatWindow.addEventListener("mousedown", (e) => {
    // Exclude interactive elements from drag
    if (
      ["INPUT", "TEXTAREA", "BUTTON"].includes(e.target.tagName) ||
      e.target.closest(".chat-message") ||
      e.target.closest(".chat-users-list") ||
      e.target.closest(".chat-input-section")
    ) {
      return;
    }
    if (window.innerWidth < 768) return; // disable on mobile
    isDragging = true;
    const rect = chatWindow.getBoundingClientRect();
    offsetX = e.clientX - rect.left;
    offsetY = e.clientY - rect.top;
    chatWindow.classList.add("dragging");
    chatWindow.classList.add("draggable-active");
  });

  document.addEventListener("mousemove", (e) => {
    if (!isDragging) return;
    e.preventDefault();
    let left = e.clientX - offsetX;
    let top = e.clientY - offsetY;

    // Keep inside viewport bounds
    const maxLeft = window.innerWidth - chatWindow.offsetWidth;
    const maxTop = window.innerHeight - chatWindow.offsetHeight;
    left = Math.max(0, Math.min(maxLeft, left));
    top = Math.max(0, Math.min(maxTop, top));

    chatWindow.style.left = left + "px";
    chatWindow.style.top = top + "px";
    chatWindow.style.right = "auto";
    chatWindow.style.bottom = "auto";
  });

  document.addEventListener("mouseup", () => {
    if (!isDragging) return;
    isDragging = false;
    chatWindow.classList.remove("dragging");
    chatWindow.classList.remove("draggable-active");

    // Save position to localStorage
    if (window.innerWidth >= 768) {
      const rect = chatWindow.getBoundingClientRect();
      const position = {
        left: rect.left,
        top: rect.top,
      };
      localStorage.setItem("chatWindowPosition", JSON.stringify(position));
    }
  });

  // Reset position on window resize to mobile
  window.addEventListener("resize", () => {
    if (window.innerWidth < 768) {
      chatWindow.style.left = "";
      chatWindow.style.top = "";
      chatWindow.style.right = "20px";
      chatWindow.style.bottom = "20px";
      // Clear saved position on mobile
      localStorage.removeItem("chatWindowPosition");
    } else {
      // On resize to desktop/tablet, try to restore saved position
      const savedPosition = localStorage.getItem("chatWindowPosition");
      if (savedPosition) {
        try {
          const pos = JSON.parse(savedPosition);
          // Validate position is within current viewport bounds
          const maxLeft = window.innerWidth - chatWindow.offsetWidth;
          const maxTop = window.innerHeight - chatWindow.offsetHeight;
          const validLeft = Math.max(0, Math.min(maxLeft, pos.left));
          const validTop = Math.max(0, Math.min(maxTop, pos.top));

          chatWindow.style.left = validLeft + "px";
          chatWindow.style.top = validTop + "px";
          chatWindow.style.right = "auto";
          chatWindow.style.bottom = "auto";
        } catch (e) {
          console.warn("Invalid saved chat position on resize:", e);
        }
      }
    }
  });
}
