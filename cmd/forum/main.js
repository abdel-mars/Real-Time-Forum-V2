window.onload = () => {
  const content = document.querySelector(".main-content");
  // Inject hero section
  content.innerHTML = `
    <!-- Hero Section -->
    <section class="hero">
      <h1 class="hero__title">Welcome to the Forum</h1>
      <p class="hero__description">
        Join the conversation. Share your thoughts. Discover amazing discussions from our vibrant community.
      </p>
      <div class="hero__cta">
        <a href="/login" class="btn-secondary" id="btn1">Login</a>
        <a href="/register" class="btn-secondary" id="btn2">Sign Up</a>
      </div>
    </section>
    <!-- Posts Section -->
    <section class="main-content__posts">
      <!-- Your posts will be injected here later -->
      <p>Posts will appear here...</p>
    </section>
    <!-- Pagination Section -->
    <nav class="pagination" aria-label="Pagination navigation">
      <span class="pagination__link pagination__link--prev" style="opacity: .3;">
        <img src="/static/svg/left.svg" alt="previous page"> Previous Page
      </span>
      <span class="pagination__current" aria-current="page">Page 1</span>
      <span class="pagination__link pagination__l.likebtn,
ink--next" style="opacity: .3;">
        Next Page <img src="/static/svg/right.svg" alt="next page">
      </span>
    </nav>
  `;
  // Optional: add event listeners to buttons if you want SPA navigation
  document.getElementById("btn1").addEventListener("click", (e) => {
    e.preventDefault();
    alert("Go to Login page (SPA style)");
  });
  document.getElementById("btn2").addEventListener("click", (e) => {
    e.preventDefault();
    alert("Go to Register page (SPA style)");
  });
};