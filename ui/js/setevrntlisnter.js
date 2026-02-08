// evernt listners !!
function escapeHtml(str) {
  if (str === null || str === undefined) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function el(tag, className = "", inner = "") {
  const e = document.createElement(tag);
  if (className) e.className = className;
  if (inner) e.innerHTML = inner;
  return e;
}

export async function fetchPosts(filter = "", page=1) {

  try {
    //const res = await fetch(`/api/posts?filter=${encodeURIComponent(filter)}`);
    const res = await fetch(`/api/posts?filter=${encodeURIComponent(filter)}&page=${page}`);
    const data = await res.json();
    // Still Need To Fix It's No Posts When I Want Her Just Some Handlling
    // Normalize Posts Array From Several Possible Shapes <<==>>
    const pagination = {
      "CurrentPage" : data.CurrentPage,
      "HasNext" :       data.HasNext,
      "HasPrev" : data.HasPrev,
      "NextPage" : data.CurrentPage + 1,
      "PrevPage" : data.CurrentPage -1 ,
    }
    let postsList = normalizePosts(data) 
    const postsContainer = document.querySelector(".main-content__posts");
    if (!postsContainer) {
      console.warn("No .main-content__posts container found.");
      return;
    }
    postsContainer.innerHTML = "";

    if (!postsList || postsList.length === 0) {
      postsContainer.innerHTML = `
        <div class="container">
          <div style="display:flex;align-items:center;justify-content:center;">
            <p>No posts found.</p>
          </div>
        </div>
      `;
      return;
    }
    postsList.forEach(p => {
      const Id = p.Id ?? p.id ?? "";
      const Initial = escapeHtml(p.Initial ?? p.initial ?? "");
      const Publisher = escapeHtml(p.Publisher ?? p.publisher ?? "");
      const Created_at = escapeHtml(p.Created_at ?? p.created_at ?? "");
      const Title = escapeHtml(p.Title ?? p.title ?? "");
      const Content = p.Content ?? p.content ?? "";
      const Catigories = p.Catigories ?? p.catigories ?? p.categories ?? [];
      const HasCategories = Array.isArray(Catigories) && Catigories.length > 0;
      const CommentsCount = p.CommentsCount ?? p.commentsCount ?? 0;

      const containerDiv = el("div", "container");
      const postDiv = el("div", "post");

      // Top
      const topDiv = el("div", "top");
      const postHeadWrap = el("div", "post_head");
      if (Id) postHeadWrap.id = Id;

      // user_container
      const userContainer = el("div", "user_container");
      const initialDiv = el("div", "initial");
      initialDiv.innerHTML = `<span>${Initial}</span>`;
      userContainer.appendChild(initialDiv);

      const infoDiv = el("div", "info");
      infoDiv.appendChild(el("div", "publisher", `<strong>${Publisher}</strong>`));
      infoDiv.appendChild(el("div", "date", `<span>${Created_at}</span>`));
      userContainer.appendChild(infoDiv);

      postHeadWrap.appendChild(userContainer);
      postHeadWrap.appendChild(el("span", "divider"));
      postHeadWrap.appendChild(el("h2", "", Title));

      const section = el("section");
      const pContent = el("p", "clamp");
      pContent.textContent = Content;  
      section.appendChild(pContent);

      topDiv.appendChild(postHeadWrap);
      topDiv.appendChild(section);

      // Bottom
      const bottomDiv = el("div", "bottom");

      // categories
      const categoriesDiv = el("div", "categories");
      const catWrap = el("div", "catigories");
      if (HasCategories) {
        Catigories.forEach(cat => {
          catWrap.appendChild(el("span", "category", escapeHtml(cat)));
        });
      } else {
        catWrap.innerHTML = `<span class="category">Uncategorized</span>`;
      }
      
      categoriesDiv.appendChild(catWrap);
      bottomDiv.appendChild(categoriesDiv);
      bottomDiv.appendChild(el("span", "divider"));

      // postbot: only comment count link + ReadMore (no like/dislike)
      const postbot = el("div", "postbot");
      const smallActions = el("div", "likes_dislikes_comment");

      //i will switch this to be not clicked
      smallActions.innerHTML = `
        <div href="/post?Id=${Id}#comment">
          <img class="comment-icon" src="/svg/commentt.svg" alt="comment" /> ${CommentsCount}
        </div>
      `;

      const readMoreDiv = el("div", "ReadMore");
      readMoreDiv.innerHTML = `<a href="#" class="readmore" data-id="${Id}">Read More</a>`;

      postbot.appendChild(smallActions);
      postbot.appendChild(readMoreDiv);
      bottomDiv.appendChild(postbot);

      postDiv.appendChild(topDiv);
      postDiv.appendChild(bottomDiv);
      containerDiv.appendChild(postDiv);

      postsContainer.appendChild(containerDiv);
      renderPagination(pagination);
    });
  } catch (err) {
    console.error("Failed to fetch posts:", err);
  }
}

function normalizePosts(data) {
  // here i will protect them from null posts And Not Condition IT's applicate !!
  if (!data) return [];

  if (Array.isArray(data.Posts)) return data.Posts;
  if (data.Posts && Array.isArray(data.Posts.Posts)) return data.Posts.Posts;
  if (Array.isArray(data.posts)) return data.posts;
  if (data.Posts && Array.isArray(data.Posts.posts)) return data.Posts.posts;

  return [];   
}

function renderPagination(pagination) {

  const container = document.querySelector(".pagination");
  if (container) {
    container.style.display = ""
  }
  if (!container) return;
  // BUILD THE HTML PAGINATION
  container.innerHTML = `
    ${pagination.HasPrev
      ? `<a href="#" class="pagination__link pagination__link--prev" data-page="${pagination.PrevPage}">
           <img src="/svg/left.svg" alt="previous page"> Previous
         </a>`
      : `<span class="pagination__link pagination__link--prev" style="opacity:.3;">
           <img src="/svg/left.svg" alt="previous page">
         </span>`}
    <span class="pagination__current">Page ${pagination.CurrentPage}</span>
    ${pagination.HasNext
      ? `<a href="#" class="pagination__link pagination__link--next" data-page="${pagination.NextPage}">
           Next <img src="/svg/right.svg" alt="next page">
         </a>`
      : `<span class="pagination__link pagination__link--next" style="opacity:.3;">
           <img src="/svg/right.svg" alt="next page">
         </span>`}
  `;
  //  Attach click events to the active links 
  container.querySelectorAll("a[data-page]").forEach(link => {
    link.addEventListener("click", e => {
      e.preventDefault();
      const page = Number(link.dataset.page);
      fetchPosts("", page);
    });
  });
}

