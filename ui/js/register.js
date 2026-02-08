import { showPage } from "./main.js";
import { destroyChatIsland } from "./chat-core.js";


export async function registerBuilding(user) {
  destroyChatIsland();
  if (!document.querySelector('link[href*="forms.css"]')) {
    const formsCSS = document.createElement('link');
    formsCSS.id = "cssl9dim";
    formsCSS.rel = 'stylesheet';
    formsCSS.href = '/css/forms.css';
    document.head.appendChild(formsCSS);
  }
  const mainContent = document.querySelector('.main-content');
  const mobileHeader = document.querySelector('.mobileHeader');
  const pagination = document.querySelector('.pagination');
  const footerContainer = document.querySelector('.footer');
  if (mainContent) mainContent.innerHTML = "";
  if (mobileHeader) mobileHeader.innerHTML = "";
  if (pagination) pagination.remove();
  if (footerContainer) footerContainer.innerHTML = "";
  function createEl(tag, className = "", html = "") {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (html) el.innerHTML = html;
    return el;
  }
  if (mobileHeader) {
    mobileHeader.innerHTML = `
      <h1 class="header-message">Welcome to 4UM — where every dev begins their journey</h1>
      <button class="signup-btn" id="Login">Login</button>
    `;
  }
  const container = createEl("div", "login-section register-page");
  const form = createEl("form", "login-card");
  form.action = "/register";
  form.method = "POST";

  form.innerHTML = `
    <img src="/assets/logo-4um.svg" alt="4UM" class="login-logo">
    <h2>Join the community!</h2>
    <input type="text" name="username" id="username" minlength="3" maxlength="32" placeholder="Nickname" required />
    <input type="number" name="age" id="age" min="13" max="120" placeholder="Age" required />
    <select name="gender" id="gender" required>
      <option value="" disabled selected>Select gender</option>
      <option value="male">Male</option>
      <option value="female">Female</option>
      <option value="other">Other</option>
    </select>
    <input type="text" name="first_name" id="first_name" minlength="2" maxlength="32" placeholder="First Name" required />
    <input type="text" name="last_name" id="last_name" minlength="2" maxlength="32" placeholder="Last Name" required />
    <input type="email" name="email" id="email" minlength="6" maxlength="254" placeholder="Your Email" required />
    <input type="password" name="password" id="password" minlength="8" maxlength="72" placeholder="Password" required />
    <input type="password" name="confirm_password" id="confirm_password" minlength="8" maxlength="72" placeholder="Confirm Password" required />
    <div id="errorMessage" class="error-container" style="display:none;"></div>
    <button type="submit" class="register-btn">Register</button>
    <p class="register-link">Already have an account? <a href="#" id="Login">Login here</a></p>
  `;
  
  container.appendChild(form);

  const footer = createEl("footer");
  footer.innerHTML = `
    <div class="footer-container">
      <p>&copy; 2026 4um — Built with passion</p>
    </div>
  `;

  if (mainContent) mainContent.appendChild(container);
  if (footerContainer) footerContainer.appendChild(footer);

  // Add class to hide chat on register page
  document.body.classList.add('register-page');

  // Add warning for gender field manipulation
  const genderSelect = document.getElementById("gender");
  const warningMessage = document.createElement("div");
  warningMessage.id = "genderWarning";
  warningMessage.style.cssText = "color: red; font-size: 14px; margin-top: 5px; display: none;";
  warningMessage.textContent = "Warning: Attempting to modify gender selection is not allowed. Please select a valid option.";
  genderSelect.parentNode.insertBefore(warningMessage, genderSelect.nextSibling);

  // Detect changes to gender select (including dev tool manipulation)
  const observer = new MutationObserver(() => {
    const value = genderSelect.value;
    if (value !== "male" && value !== "female" && value !== "other" && value !== "") {
      warningMessage.style.display = "block";
      genderSelect.value = ""; // Reset to default
    } else {
      warningMessage.style.display = "none";
    }
  });

  observer.observe(genderSelect, {
    attributes: true,
    attributeFilter: ['value']
  });

  // Also check on change event
  genderSelect.addEventListener("change", () => {
    const value = genderSelect.value;
    if (value !== "male" && value !== "female" && value !== "other" && value !== "") {
      warningMessage.style.display = "block";
      genderSelect.value = "";
    } else {
      warningMessage.style.display = "none";
    }
  });

  //form.appendChild(errorMessage)
  // form submit handler
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const errorMessage = document.getElementById("errorMessage");
    errorMessage.style.display = "none";
    errorMessage.textContent = "";
    const username = document.getElementById("username").value;
    const age = document.getElementById("age").value;
    const gender = document.getElementById("gender").value;
    const firstName = document.getElementById("first_name").value;
    const lastName = document.getElementById("last_name").value;
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;
    const confirmpassword = document.getElementById("confirm_password").value;
    if (password !== confirmpassword) {
      errorMessage.style.display = "block";
      errorMessage.textContent = "Passwords do not match.";
      return;
    }
    // Validate gender selection
    if (!gender || gender === "") {
      errorMessage.style.display = "block";
      errorMessage.textContent = "Please select a gender.";
      return;
    }
    try {
      const res = await fetch("/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username,
          age : Number(age),
          gender,
          first_name: firstName,
          last_name: lastName,
          email,
          password,
          confirmpassword
        })

      });
      const result = await res.json();
      if (res.ok && result.status === "ok") {
        errorMessage.style.display = "none";
        const userRes = await fetch("/api/me");
        const user = await userRes.json();
        await showPage("login", user);
        sessionStorage.setItem("currentPage", "login");
      } else {
        errorMessage.style.display = "block";
        errorMessage.textContent = result.message || "Registration failed";
      }
    } catch (err) {
      errorMessage.style.display = "block";
      errorMessage.textContent = "Server error. Try again later.";
    }
  });
}
