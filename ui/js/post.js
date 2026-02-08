export async function POST(id, user, page = 1) {

  if (!document.querySelector('link[href*="post.css"]')) {
    const link = document.createElement('link');
    link.id = "cssPOST"
    link.rel = 'stylesheet';
    link.href = '/css/post.css';
    document.head.appendChild(link);
  }

  const main = document.querySelector('.main-content');
  const mobileHeader = document.querySelector('.mobileHeader');
  const footerEl = document.querySelector('.footer');
  const container = document.querySelector(".pagination");
  container.style.display = "none"
  const existingHeader = document.querySelector('header.header');
  if (existingHeader) existingHeader.remove();
  if (mobileHeader) {
    mobileHeader.innerHTML = "";
    mobileHeader.style.display = "";
  }
  if (mobileHeader) {
    mobileHeader.innerHTML = `
<img id="Logo" src="/assets/logo-4um.svg" alt="4um" class="brand-logo" width="140" height="auto"><div class="mobileHeader_profile">
        <nav class="profile-nav">
          <ul>
            <li class="profile-dropdown">
              <div class="header__profile">
                <div class="initial">${user.Initial || ''}</div>
                <div class="username">${user.Username || ''}</div>
              </div>
              <ul>
            </li>
                <li><a id ="Out"> <img src="/svg/log-out.svg" /> Logout</a></li>
              </ul>
            </li>
          </ul>
        </nav>
      </div>`;
  }

  // Show Loading In Main Content
  if (main) main.innerHTML = `<p style="text-align:center; margin:2rem;">Loading post...</p>`;

  try {
    const res = await fetch(`/post?Id=${id}&page=${page}`, { method: "GET" });
    if (!res.ok) throw new Error("Failed to fetch post");
    const data = await res.json();

    const Pagination = {
      "curent": data.currentPage,
      "hasNext": data.hasNext,
      "hasprev": data.hasPrev,
      "hasnext": data.nextPage
    }

    Create({
      post: data.post,
      comments: data.comments || [],
      user,
      auth: data.user || {},
      footerEl, Pagination
    });
  } catch (err) {
    console.error("Error fetching single post:", err);
    if (main) main.innerHTML = `<p style="text-align:center; color:red;">Failed to load post.</p>`;
  }
}

//  Create Post 
export function Create({ post, comments = [], user, footerEl, Pagination = null }) {

  const main = document.querySelector('.main-content');
  const mobileHeader = document.querySelector('.mobileHeader');

  if (main) main.innerHTML = '';

  //  Post HTML
  const postHTML = `
    <div class="container">
      <div class="post">
        <div class="top">
          <div class="post_head">
            <div class="user_container">
              <div class="initial"><span>${post.Initial}</span></div>
              <div class="info">
                <div class="publisher"><strong>${post.Publisher}</strong></div>
                <div class="date"><span>${post.Created_at}</span></div>
              </div>
            </div>
            <span class="divider"></span>
            <h2>${post.Title}</h2>
          </div>
        </div>
        <section><p>${post.Content}</p></section>
        <span class="divider"></span>
        <footer>
        <div class="post" id="post-${post.Id}">
          <section id="likes_dislikes">
            <form action="/like" method="POST">
              <input type="hidden" name="post_id" value="${post.Id}">
              <button type="submit" class="likebtn">
                <img src="/svg/${post.IsLikedByUser ? 'likee' : 'green_likee'}.svg" alt="like">
                ${post.Likes}
              </button>
            </form>
            <form action="/dislike" method="POST">
              <input type="hidden" name="post_id" value="${post.Id}">
              <button type="submit" class="dislikebtn">
                <img src="/svg/${post.IsDislikedByUser ? 'dislikee' : 'red_dislikee'}.svg" alt="dislike">
                ${post.Dislikes}
              </button>
            </form>
          </section>
          </div>
        </footer>
      </div>
    </div>
  `;

  const commentFormHTML = `
    <div class="containercomment">
      <form id="form_comment" action="#" method="POST">
        <input type="hidden" name="post_id" value="${post.Id}">
        <textarea id="comment" name="comment" placeholder="comment..." minlength="1" maxlength="1000" required></textarea>
        <button type="submit"><img src="/svg/send.svg" alt="comment"></button>
      </form>
    </div>
  `;

  const commentsHTML = comments.length
    ? comments.map(c => `
        <div class="container">
          <div class="commentcontainer" id="${c.CommentId}">
            <div class="commentTop">
              <div class="initial">${c.initial}</div>
              <div class="info">
                <div class="publisher"><strong>${c.username}</strong></div>
                <div class="content"><span>${c.content}</span></div>
              </div>
            </div>
          </div>
        </div>
      `).join('')
    : `<div id="Com" class="container"><p style="text-align:center;">No comments yet.</p></div>`;
  let paginationHTML = '';
  if (Pagination) {
    paginationHTML = `
  <div class="post-nav-container">
    <nav class="post-nav" aria-label="Pagination navigation">
      ${Pagination.hasprev ?
        `<button class="post-nav__btn post-nav__btn--prev post-nav__action" 
                     data-page="${Pagination.curent - 1}" 
                     data-post-id="${post.Id}" 
                     aria-label="Previous page">
               <img src="/svg/left.svg" alt="previous page"> Previous Page
             </button>` :
        `<span class="post-nav__btn post-nav__btn--prev post-nav__btn--disabled" aria-label="Previous page">
               <img src="/svg/left.svg" alt="previous page">
             </span>`
      }
      <span class="post-nav__status" aria-current="page">Page ${Pagination.curent}</span>
      ${Pagination.hasNext ?
        `<button class="post-nav__btn post-nav__btn--next post-nav__action" 
                     data-page="${Pagination.curent + 1}" 
                     data-post-id="${post.Id}" 
                     aria-label="Next page">
               Next Page <img src="/svg/right.svg" alt="next page">
             </button>` :
        `<span class="post-nav__btn post-nav__btn--next post-nav__btn--disabled" aria-label="Next page">
               <img src="/svg/right.svg" alt="next page">
             </span>`
      }
    </nav>
  </div>
`;

  }

  if (main) main.innerHTML = postHTML + commentFormHTML + `<div class="comments-list">${commentsHTML}</div>` + paginationHTML;

  const likeForm = document.querySelector('#likes_dislikes form[action="/like"]');
  const dislikeForm = document.querySelector('#likes_dislikes form[action="/dislike"]');

  if (likeForm) {
    likeForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData(likeForm);
      try {
        const res = await fetch('/like', {
          method: 'POST',
          body: formData
        });
        // if(res.status === 401) {}
        if (!res.ok) throw new Error('Failed to submit like');
        const result = await res.json();
        updateLikeDislikeUI(result.data);

      } catch (err) {
        console.error('Error submitting like:', err);
        alert('Failed to like post. Please try again.');
      }
    });
  }

  if (dislikeForm) {
    dislikeForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData(dislikeForm);
      try {
        const res = await fetch('/dislike', {
          method: 'POST',
          body: formData
        });
        if (!res.ok) throw new Error('Failed to submit dislike');
        const result = await res.json();
        updateLikeDislikeUI(result.data);

      } catch (err) {
        console.error('Error submitting dislike:', err);
        alert('Failed to dislike post. Please try again.');
      }
    });
  }

  function updateLikeDislikeUI(result) {
    const container = document.querySelector(`#post-${result.Id}`);
    if (!container) return;
    const likeBtn = container.querySelector('.likebtn');
    const dislikeBtn = container.querySelector('.dislikebtn');
    if (likeBtn && result.Likes !== undefined) {
      likeBtn.innerHTML = `
        <img src="/svg/${result.IsLikedByUser ? 'likee' : 'green_likee'}.svg" alt="like">
        ${result.Likes}
      `;
    }
    if (dislikeBtn && result.Dislikes !== undefined) {
      dislikeBtn.innerHTML = `
        <img src="/svg/${result.IsDislikedByUser ? 'dislikee' : 'red_dislikee'}.svg" alt="dislike">
        ${result.Dislikes}
      `;
    }
  }
  // Footer 
  if (footerEl) {
    footerEl.innerHTML = `<div class="footer-container">
      <p>&copy; 2026 4um — Built with passion</p>
    </div>`;
  }
  const paginationBtns = document.querySelectorAll('.post-nav__action');
  paginationBtns.forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.preventDefault();
      const page = parseInt(btn.dataset.page);
      const postId = btn.dataset.postId;
      if (main) {
        // styling loading 
        main.innerHTML = `<p style="text-align:center; margin:2rem;">Loading page ${page}...</p>`;
      }
      await POST(postId, user, page);
    });
  });

  const formComment = document.querySelector("#form_comment");

  if (formComment) {
    formComment.addEventListener("submit", async (e) => {
      e.preventDefault();

      let curent = Pagination.curent;
      let post_id = post.Id;
      const formData = new FormData(formComment);

      // Add current page to form data
      formData.append("currentPage", curent);

      try {
        const res = await fetch("/comment", { method: "POST", body: formData });
        if (!res.ok) throw new Error("Failed to submit comment");
        const newComment = await res.json();

        // If user is on another page, redirect to page 1
        if (newComment.no) {
          await POST(post_id, user, 1);
          return;
        }

        const commentsList = document.querySelector(".comments-list");
        const visible = commentsList.querySelectorAll(".commentcontainer");
        const max = 10;

        if (visible.length >= max) {
          // Refresh the entire page to show updated pagination
          await POST(post_id, user, curent);
          return;
        }

        // If less than 10 comments, add the new one dynamically
        if (commentsList && newComment) {
          const node = createCommentNode(newComment);
          commentsList.prepend(node);
        }

        // Remove "No comments yet" message if it exists
        const NOcoment = document.getElementById('Com');
        if (NOcoment) {
          NOcoment.remove();
        }

        formComment.reset();
      } catch (err) {
        console.error("Error submitting comment:", err);
        alert("Failed to submit comment. Please try again.");
      }
    });
  }
}

function escapeHtml(s) {
  if (!s) return "";
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function createCommentNode(c) {
  const id = c.comment_id ?? c.CommentId ?? c.commentId ?? "";
  const initial = c.initial ?? c.Initial ?? "";
  const username = c.username ?? c.Username ?? "";
  const content = c.content ?? c.Content ?? "";

  const wrapper = document.createElement("div");
  wrapper.className = "container";
  wrapper.innerHTML = `
    <div class="commentcontainer" data-comment-id="${escapeHtml(id)}">
      <div class="commentTop">
        <div class="initial">${escapeHtml(initial)}</div>
        <div class="info">
          <div class="publisher"><strong>${escapeHtml(username)}</strong></div>
          <div class="content"><span>${content}</span></div>
        </div>
      </div>
      <span class="divider"></span>
      <div class="commentBottom"></div>
    </div>
  `;
  return wrapper;
}