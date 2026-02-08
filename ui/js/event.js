import { showPage } from "./main.js";
import{ destroyChatIsland} from "./chat-core.js"

export async function loginbuilding(user) {
  // Destroy chat island when entering login page
  destroyChatIsland();
  // Navigation Clear It Applicate Vanila Is Hill
  const HOWA = document.getElementById('chat-container')
  if (HOWA) {
    HOWA.remove()
  }
  const existingHomeHeader = document.querySelector('header.header');
  if (existingHomeHeader) existingHomeHeader.remove();
  const m = document.querySelector(".mobileHeader");
  if (m) {
    m.style.display = "";  
    m.innerHTML = ""; 
  }

  const p = document.querySelector(".pagination");
  if (p) {
  p.style.display = "none"
  }
  if (!document.querySelector('link[href*="forms.css"]')) {
    const formsCSS = document.createElement('link');
    formsCSS.id = "Csslogin"
    formsCSS.rel = 'stylesheet';
    formsCSS.href = '/css/forms.css';
    document.head.appendChild(formsCSS);
  }

  const mainContent = document.querySelector('.main-content');
  const mobileHeader = document.querySelector('.mobileHeader');
  const pagination = document.querySelector('.pagination');
  const footerContainer = document.querySelector('.footer');

  // Clear The Containers
  if (mainContent) mainContent.innerHTML = "" , mainContent.style.display = "";
  if (mobileHeader) mobileHeader.innerHTML = "";
  if (pagination) pagination.innerHTML = "" 
  if (footerContainer) footerContainer.innerHTML = "", footerContainer.style.display = "";

  // Helper To Create Elements
  function createEl(tag, className = "", html = "") {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (html) el.innerHTML = html;
    return el;
  }

  //  Header 
  const headerContent = `
    <h1 class="header-message">Welcome back to 4UM — where every dev meets here</h1>
    <button class="signup-btn" id="Register">Sign Up</button>
  `;

  // Append header content directly to mobileHeader (no nested <header>)
  if (mobileHeader) mobileHeader.innerHTML = headerContent;

  // Form 
  const container = createEl("div", "login-section");
  const form = createEl("form", "login-card");
  form.innerHTML = `
    <img src="/assets/logo-4um.svg" alt="4UM" class="login-logo">
    <h2>Welcome Back!</h2>
    <label>Username or Email</label>
    <input type="text" name="username" id="username" minlength="3" maxlength="32" placeholder="Enter your username" required />
    <label>Password</label>
    <input type="password" name="password" id="password" minlength="8" maxlength="72" placeholder="Enter your password" required />
    <div id="errorMessage" class="error-container" style="display:none;"></div>
    <button type="submit" class="login-btn">Login</button>
    <p class="register-link">Don’t have an account? <a href="#" id="Register">Register here</a></p>
  `;
  container.appendChild(form);
  //Addeventlistnerandpost();

  const footer = createEl("footer");
  footer.innerHTML = `
    <div class="footer-container">
      <p>&copy; 2025 4um — Built with passion</p>
    </div>
  `;

  if (mainContent) mainContent.appendChild(container);
  if (footerContainer) footerContainer.appendChild(footer);

  // Add class to hide chat on login page
  document.body.classList.add('login-page');

  // Handle form submit
  //const errorMessage = document.getElementById("errorMessage");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const errorMessage = e.target.querySelector("#errorMessage");
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;
    errorMessage.style.display = "none";
    errorMessage.textContent = "";

    try {
      const res = await fetch("/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
        credentials: "same-origin",
      });

      const result = await res.json();
      if (res.ok && result.status === "ok") {
        errorMessage.style.display = "none";
        if (mobileHeader) {
          mobileHeader.innerHTML = "";
          mobileHeader.style.display = "none";
        }
        const css = document.getElementById('Csslogin')
        const css1 = document.getElementById('css')
        if (css) {
        css.remove()
        }
        if (css1) {
          css1.remove()
        }

        await showPage("home");
        sessionStorage.setItem("currentPage", "home");
        //await homeBuild(user)
      } else {
        errorMessage.style.display = "block";
        errorMessage.textContent = result.message || "Login failed";
      }
    } catch (err) {
      errorMessage.style.display = "block";
      errorMessage.textContent = "Server error. Try again later.";
    }
  });
}

export async function Logoutfunc() {
  // Close all WebSocket connections before logout
  if (window.chatWebSocket) {
    window.chatWebSocket.close();
    window.chatWebSocket = null;
  }
  
  try {
    const response = await fetch("/logout", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
           "X-Requested-With": "fetch" 
      },
    });
    const data = await response.json();
    if (response.ok) {
      console.log(data.message);
      const res = await fetch("/api/me");
      const user = await res.json();
      // Signal other tabs that a logout occurred
      try {
        localStorage.setItem('logout', Date.now().toString());
      } catch (e) {
        console.warn('localStorage unavailable:', e);
      }

      // If BroadcastChannel is available, post a message too (faster, reliable)
      try {
        const bc = new BroadcastChannel('auth');
        bc.postMessage({ type: 'logout' });
        bc.close();
      } catch (e) {
        // ignore
      }

      await showPage("login", user);
      sessionStorage.setItem("currentPage", "login");
    } else {
      alert("Logout failed: " + data.error);
    }
  } catch (err) {
    alert("Something went wrong while logging out ..........");
  }
}