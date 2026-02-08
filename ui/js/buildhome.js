 
import { fetchPosts } from "./setevrntlisnter.js";
import { creatchatisland, showChatIsland } from "./chatisland.js"

export async function homeBuild(user) {
  // Remove login/register classes to show chat on home
  document.body.classList.remove('login-page', 'register-page');

  // Create floating chat island if user is authenticated
  if (user.authenticated) {
    creatchatisland()
    // Show the chat island after a short delay to ensure it's rendered
    setTimeout(() => {
      showChatIsland();
    }, 500);
  }

  const HOWA = document.getElementById('chat-container')
  if (HOWA) {
    HOWA.remove()
  }
  // check Authenticatec evey time to be sure the user it's authenticated 
  const ii = document.getElementById("Css")
  if (ii) {
    ii.remove()
  }
  //const rm = document.getElementById
  const mobileHeader = document.querySelector('.mobileHeader');
  if (mobileHeader) {
    mobileHeader.innerHTML = "";
    mobileHeader.style.display = "none";
  }
  const mainContent = document.querySelector('.main-content');
  if (mainContent) {
    mainContent.style.display = ""
  }
  if (mainContent && !mainContent.querySelector('.hero')) {
    mainContent.innerHTML = `
            <section class="hero">
            </section>
            <section class="main-content__posts">
                <p>post will be here </p>
            </section>
            <nav class="pagination" aria-label="Pagination navigation"></nav>
        `;
  }
  // < END OF ADDITION !!!>
  function createEl(tag, className = '', html = '') {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (html) el.innerHTML = html;
    return el;
  }
  // Remove existing header to avoid duplicates
  const existingHeader = document.querySelector('header.header');
  if (existingHeader) existingHeader.remove();
  // Build the header dynamically
  const body = document.body;
  const header = createEl("header", "header");

  // Logo handler !!
  const logoDiv = createEl("div", "header__logo", `
    <a id = "Logo" class="header__logo-link">
      <img id = "Logo" src="/svg/logo-4um.svg" alt="4um" class="brand-logo" width="140" height="auto">
    </a>
  `);

  header.appendChild(logoDiv);
  // User actions container
  const actionsDiv = createEl("div", "user_actions--header");

  // Menu button for small screens
  const menuButton = createEl("button", "header__menu-button", `
    <img src="/svg/settings.svg" alt="Menu" class="menu-icon">
  `);
  menuButton.setAttribute('aria-label', 'Toggle menu');
  actionsDiv.appendChild(menuButton);

  // Menu container
  const menuContainer = createEl("div", "header__menu-container");

  const nav = createEl("nav", "profile-nav");
  const ul = createEl("ul");
  const liDropdown = createEl("li", "profile-dropdown");
  const filterDiv = createEl("div", "header__filter-label");
  filterDiv.innerHTML = ` <img src="/svg/incon_filter.svg" alt="Filter by Categories"> Filter `;
  liDropdown.appendChild(filterDiv);
  // Make the filter label accessible and interactive
  filterDiv.setAttribute('role', 'button');
  filterDiv.setAttribute('tabindex', '0');
  filterDiv.setAttribute('aria-haspopup', 'true');
  filterDiv.setAttribute('aria-expanded', 'false');

  // Toggle dropdown open/close
  filterDiv.addEventListener('click', (e) => {
    e.stopPropagation();
    const isOpen = liDropdown.classList.toggle('open');
    filterDiv.setAttribute('aria-expanded', String(isOpen));
  });

  filterDiv.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      filterDiv.click();
    } else if (e.key === 'Escape') {
      liDropdown.classList.remove('open');
      filterDiv.setAttribute('aria-expanded', 'false');
      filterDiv.focus();
    }
  });

  // Delegate clicks on filter links (will work once filterUl is appended)
  liDropdown.addEventListener('click', async (e) => {
    const anchor = e.target.closest('a[data-filter]');
    if (!anchor) return;
    e.preventDefault();
    const filter = decodeURIComponent(anchor.dataset.filter || '');
    // Fetch posts for the selected filter
    await fetchPosts(filter);
    // Close dropdown after selection
    liDropdown.classList.remove('open');
    filterDiv.setAttribute('aria-expanded', 'false');
  });

  // Close dropdown when clicking outside or pressing Escape anywhere
  document.addEventListener('click', (e) => {
    if (!liDropdown.contains(e.target)) {
      liDropdown.classList.remove('open');
      filterDiv.setAttribute('aria-expanded', 'false');
    }
  });

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      liDropdown.classList.remove('open');
      filterDiv.setAttribute('aria-expanded', 'false');
    }
  });
  const filterUl = createEl("ul");
  if (user.fileds) {
    Object.keys(user.fileds).forEach(key => {
      const li = createEl("li");
      li.innerHTML = `<a href="#" data-filter="${encodeURIComponent(key)}">${key}</a>`;
      filterUl.appendChild(li);
    });
  }

  if (user.authenticated) {
    filterUl.innerHTML += `<li><a href="#" data-filter="Owned">Owned</a></li>`;
    filterUl.innerHTML += `<li><a href="#" data-filter="Likes">Liked</a></li>`;
  }

  liDropdown.appendChild(filterUl);
  ul.appendChild(liDropdown);
  nav.appendChild(ul);
  menuContainer.appendChild(nav);

  // Create Post Button !!
  const postDiv = createEl("div", "nes_post--header", `
    <a href="#" id = "Newpost" aria-label="Crea newpost" >
      <img class="SvgImg" src="/svg/pencil-linee.svg" alt="New post icon">
      Create Post
    </a>
  `);

  menuContainer.appendChild(postDiv);

// Authenticated One !!
  if (user.authenticated) {
    const profileNav = createEl("nav", "profile-nav");
    profileNav.innerHTML = `
      <ul>
        <li class="profile-dropdown">
          <div class="header__profile">
            <div class="initial">${user.Initial}</div>
            <div class="username">${user.Username}</div>
          </div>
          <ul>
            <li>
              <a href="#" id ="Out" class="header__profile-link">
                <img src="/svg/log-out.svg" class="header__profile-icon" data-action="logout"> Logout
              </a>
            </li>
          </ul>
        </li>
      </ul>
    `;
    menuContainer.appendChild(profileNav);
  } else {
    // Here There Is The The Logout Out Stat I Need to lisntnign to it !!
    const authLinks = createEl("div", "header__auth-links", `
      <a href="#" id = "Register" class="header__auth-link">Sign Up</a>
      <a href="#" id = "Login" class="header__auth-link">Login</a>
    `);
    menuContainer.appendChild(authLinks);
  }

  actionsDiv.appendChild(menuContainer);

  // Toggle menu on small screens
  menuButton.addEventListener('click', () => {
    menuContainer.classList.toggle('open');
  });
  // |< <==-|-==> >|
  header.appendChild(actionsDiv);
  body.insertBefore(header, body.firstChild);
  // build hero !! 
  // Get the posts for the home page with no filter to show all posts
  await fetchPosts("")
  buildHero(user)
  const foot = document.getElementById("333")
  if (foot) {
    foot.style.display = ""
  }
  foot.innerHTML = `
    <div class="footer__container">
      <p class="footer__text">&copy; 2026 4um — Built with passion</p>
    </div>
  `
}

function buildHero(user) {
  const hero = document.querySelector(".hero");
  hero.innerHTML = `
    <h1 class="hero__title">Welcome to the Forum</h1>
    <div class="hero__cta">
      ${user.authenticated
      ? ``
      : `<a href="#" id = "Login"  class="btn-secondary">Login</a>
           <a href="#" id = "Register"  class="btn-secondary">Sign Up</a>`}
    </div>
  `;
}
