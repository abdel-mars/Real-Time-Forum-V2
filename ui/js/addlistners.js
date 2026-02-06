import { Logoutfunc } from "./event.js";
import { fetchPosts } from "./setevrntlisnter.js";
import { showPage } from "./main.js"


export function setupGlobalListeners(User) {
  // Monitor cookie changes and reload page if changed
  let lastCookie = document.cookie;
  setInterval(() => {
    console.log("Current cookie:", document.cookie);
    if (document.cookie !== lastCookie) {
      lastCookie = document.cookie;
      window.location.reload();
    }
  }, 1000); // Check every second

  // Listen for logout signals from other tabs (localStorage + BroadcastChannel)
  window.addEventListener('storage', (e) => {
    if (!e) return;
    if (e.key === 'logout') {
      // Another tab logged out — redirect this tab to login
      try {
        // Use showPage if available to render SPA login view
        showPage('login');
      } catch (err) {
        // Fallback: full reload to /login
        window.location.href = '/login';
      }
    }
  });

  // BroadcastChannel listener for browsers that support it (faster than storage)
  try {
    const bc = new BroadcastChannel('auth');
    bc.addEventListener('message', (ev) => {
      if (ev && ev.data && ev.data.type === 'logout') {
        try {
          showPage('login');
        } catch (err) {
          window.location.href = '/login';
        }
      }
    });
  } catch (e) {
    // BroadcastChannel not supported — localStorage storage event will suffice
  }

  document.addEventListener("click", async (event) => {
    let target = event.target;
    const filterLink = target.closest("a[data-filter]");
    if (filterLink) {
      event.preventDefault();
      const filter = decodeURIComponent(filterLink.dataset.filter);
      const res = await fetch("/api/me");
      const user = await res.json();
      
      await showPage("home", user);     
      await fetchPosts(filter, 1);
      return;
    }
    const POSTID = event.target.closest("a.readmore");
    if (POSTID) {
      const res = await fetch("/api/me");
      const user = await res.json();
      const postId = POSTID.dataset.id; // 
      await showPage("Post", user, postId);
      window.history.replaceState({}, "", "/Post");
      sessionStorage.setItem("currentPage", "Post");
      
    }
    if (!target) return;
    const id = target.id;
    if (id === "Login") {
      const res = await fetch("/api/me");
      const user = await res.json();
      await showPage("login", user);
    }
    else if (id === "Register") {
      const res = await fetch("/api/me");
      const user = await res.json();
      event.preventDefault();
      await showPage("register", user);
    }
    // --- <Logout> ---
    else if (id === "Out") {
      await Logoutfunc();
    }
    else if (id === "Logo") {
      const res = await fetch("/api/me");
      const user = await res.json();
      window.history.replaceState({}, "", "/home");
      if (!user.authenticated) {
        showPage("login", user)
        return
      }
      // I Will Get The Css Of Any Css <!-!>
      const css0 = document.getElementById("cssPOST")
      const css1 = document.getElementById("css")
      if (css1) {
        css1.remove()
      } else if (css0) {
        css0.remove()
      }
      await showPage("home", user)
    }
    else if (id === "Newpost") {
      const res = await fetch("/api/me");
      const user = await res.json();
      window.history.replaceState({}, "", "/createpost");
     await showPage("createpost", user);
    }
    const userItem = target.closest('.online-user');
    if (userItem && userItem.dataset.username) {
      event.preventDefault();
      const username = userItem.dataset.username;
      // Import And Call SetupPrivateChat <!-!> !
      try {
        const { setupPrivateChat } = await import('./chatui.js');
        const { currentSocket, setIntentionalSocketClose, setCurrentSocket } = await import('./chatState.js');

        if (currentSocket) {
          setIntentionalSocketClose(true);
          currentSocket.close();
          setCurrentSocket(null);
        }
        await setupPrivateChat(username);
      } catch (err) {
        console.error('Error setting up private chat:', err);
      }
    }
  });
}
