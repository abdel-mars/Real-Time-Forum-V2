import { showPage } from "./main.js";

export async function NewPost(categories = [], user) {
  // Load CSS
  if (!user) {
    const res = await fetch("/api/me");
    user = await res.json();
  }
  if (!document.querySelector('link[href*="new_post.css"]')) {
    const link = document.createElement("link");
    link.id = "Css";
    link.rel = "stylesheet";
    link.href = "/css/new_post.css";
    document.head.appendChild(link);
  }

  const mainContent = document.querySelector(".main-content");
  const mobileHeader = document.querySelector(".mobileHeader");
  const pagination = document.querySelector(".pagination");
  if (pagination) pagination.remove();
  const footerContainer = document.querySelector(".footer");

  if (mainContent) mainContent.innerHTML = "";

  if (footerContainer) footerContainer.innerHTML = "";

  function createEl(tag, className = "", html = "") {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (html) el.innerHTML = html;
    return el;
  }

  const existingHeader = document.querySelector("header.header");
  if (existingHeader) existingHeader.remove();
  if (mobileHeader) {
    mobileHeader.innerHTML = "";
    mobileHeader.style.display = "";
  }
  if (mobileHeader) {
    mobileHeader.innerHTML = `
      <img id="Logo" src="/svg/logo-4um.svg" alt="4um" class="brand-logo" width="140" height="auto">
      <div class="mobileHeader_profile">
        <nav class="profile-nav">
          <ul>
            <li class="profile-dropdown">
              <div class="header__profile">
                <div class="initial">${user.Initial || ""}</div>
                <div class="username">${user.Username || ""}</div>
              </div>
              <ul>
                <li><a id ="Out"> <img src="/svg/log-out.svg" /> Logout</a></li>
              </ul>
            </li>
          </ul>
        </nav>
      </div>`;
  }

  // Error container
  const errorMessage = createEl("div", "error-container");
  errorMessage.style.display = "none";

  // Main container
  const container = createEl("div", "container");
  
  // Single container for all form elements
  const formContainer = createEl("div", "form-container");
  formContainer.innerHTML = `
    <h1>Create a Post</h1>

    <div class="formfield">
      <input id="title" type="text" name="title" placeholder="Title..." minlength="10" maxlength="70" required />
    </div>

    <div class="formfield">
      <textarea id="content" name="content" placeholder="Insert content here..." minlength="1" maxlength="10000" required></textarea>
    </div>

    <div class="dropdown-post">
      <div class="dropdown-toggle" id="categories-toggle">Select Categories</div>
      <div class="dropdown-menu" id="categories-menu">
        ${categories.map(cat => `<label><input type="checkbox" value="${cat}"> ${cat}</label>`).join('')}
        
      </div>
    </div>

    <button type="submit" class="submit-btn">
      <span class="span-post">Post</span>
    </button>
  `;

  // Form
  const form = createEl("form", "postForm");
  form.setAttribute("method", "POST");
  form.setAttribute("action", "/newPost");

  // Assemble the form
  form.appendChild(formContainer);

  mainContent.appendChild(errorMessage);
  container.appendChild(form);

  // Footer
  const footer = createEl("footer");
  footer.innerHTML = `
      <div class="footer-container">
        <p>&copy; 2026 4um — Built with passion</p>
      </div>
    `;
  if (footerContainer) footerContainer.appendChild(footer);
  if (mainContent) mainContent.appendChild(container);

  // Dropdown functionality
  const dropdownPost = formContainer.querySelector('.dropdown-post');
  const dropdownToggle = formContainer.querySelector('#categories-toggle');
  const dropdownMenu = formContainer.querySelector('#categories-menu');
  const checkboxes = dropdownMenu.querySelectorAll('input[type="checkbox"]');

  const storageKey = "newpost_selected_categories_v1";

  dropdownToggle.addEventListener('click', () => {
    dropdownPost.classList.toggle('open');
  });

  // Prevent clicks inside the menu from bubbling to document (avoids accidental close)
  dropdownMenu.addEventListener('click', (e) => {
    e.stopPropagation();
  });

  // Close dropdown when clicking outside
  document.addEventListener('click', (e) => {
    if (!dropdownPost.contains(e.target)) {
      dropdownPost.classList.remove('open');
    }
  });

  // Restore selections from localStorage
  function restoreSelectedCategories() {
    try {
      const saved = localStorage.getItem(storageKey);
      if (!saved) return;
      const arr = JSON.parse(saved);
      if (!Array.isArray(arr)) return;
      Array.from(checkboxes).forEach(cb => {
        cb.checked = arr.includes(cb.value);
      });
    } catch (err) {
      console.warn("restoreSelectedCategories failed", err);
    }
  }

  // Update toggle text and persist selection
  function updateSelectedCategories() {
    const selected = Array.from(checkboxes).filter(cb => cb.checked).map(cb => cb.value);
    dropdownToggle.textContent = selected.length > 0 ? selected.join(', ') : 'Select Categories';
    try {
      localStorage.setItem(storageKey, JSON.stringify(selected));
    } catch (err) {
      console.warn("Could not save selected categories", err);
    }
  }

  // Ensure checkbox clicks don't bubble and update text immediately
  checkboxes.forEach(cb => {
    cb.addEventListener('click', (e) => {
      e.stopPropagation();
      updateSelectedCategories();
    });
  });

  // restore + initial render
  restoreSelectedCategories();
  updateSelectedCategories();

  // Handle form submit with fetch
  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const title = form.querySelector("#title").value.trim();
    const content = form.querySelector("#content").value.trim();
    const selectedCategories = Array.from(checkboxes).filter(cb => cb.checked).map(cb => cb.value);

    if (!title || !content) {
      errorMessage.innerText = "Title and content cannot be empty.";
      errorMessage.style.display = "block";
      return;
    }

    try {
      const formData = new FormData();
      formData.append("title", title);
      formData.append("content", content);
      selectedCategories.forEach((c) => formData.append("Categories", c));

      const res = await fetch("/newPost", {
        method: "POST",
        body: formData,
        headers: {
          "X-Requested-With": "fetch",
        },
      });
      const data = await res.json();
      if (!res.ok) {
        errorMessage.innerText = data.error || "Error submitting post";
        errorMessage.style.display = "block";
        return;
      }
      // Redirect To Home When The User Created The Post !!
      errorMessage.style.display = "none";
      const Stat = await fetch("/api/me");
      const user = await Stat.json();
      // Load and remove CSS
      const css = document.getElementById("css");
      if (css) css.remove();

      await showPage("home", user);
      window.history.replaceState({}, "", "/home");
      sessionStorage.setItem("currentPage", "home");
      // LocalStorage.setItem("")
      // HomeBuild(user)
      form.reset();
    } catch (err) {
      errorMessage.innerText = "Network error: " + err.message;
      errorMessage.style.display = "block";
    }
  });
}