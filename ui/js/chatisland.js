// CHAT ISLAND FUNCTIONALITY - COMMENTED OUT
// All chat island functionality has been split into separate modules:
// - chat-core.js: Main initialization and lifecycle
// - chat-websocket.js: WebSocket connections and messaging
// - chat-ui.js: UI management and message display
// - chat-users.js: User list and status management
// - chat-utils.js: Utility functions and helpers

// Import from the new modular files
import { creatchatisland as createChatIsland, showChatIsland, hideChatIsland, destroyChatIsland } from './chat-core.js';

// Re-export the functions to maintain compatibility
export { createChatIsland as creatchatisland, showChatIsland, hideChatIsland, destroyChatIsland };
  