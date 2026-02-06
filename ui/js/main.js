import { setupGlobalListeners } from "./addlistners.js"
import { loginbuilding } from "./event.js"
import { homeBuild } from "./buildhome.js";
import { registerBuilding } from "./register.js";
import { POST } from "./post.js";
import { NewPost } from "./newpost.js";


let currentPage = "login";

// Will Need To Handle The Message Error !! 
export function showErrorPage({
  title = "Page not found",
  message = "This page not found",
  statusCode = "404",
  statusText = "Page not found"
} = {}) {
  
if (!document.querySelector('link[href*="error-page.css"]')) {
  const errorCSS = document.createElement('link');
  errorCSS.rel = 'stylesheet';
  errorCSS.href = 'static/css/error-page.css'; // Update path as needed
  document.head.appendChild(errorCSS);
}

const main = document.querySelector(".main-content");

  if (!main) return;
  main.innerHTML = "";
  main.innerHTML = `
      <div class="error-container">
        <h1 class="error-title">${title}</h1>
        <p class="error-message">${message}</p>
        <div class="error-details">
          ${statusCode ? `<p>Status Code: ${statusCode}</p>` : ""}
          ${statusText ? `<p>Status Text: ${statusText}</p>` : ""}
        </div>
        <button id="Logo" class="home-button">Go Home</button>
      </div>
    `;

  // Update browser history to match the error page
  if (title.includes("403")) {
    window.history.replaceState({}, "", "/unauthorized");
  } else if (title.includes("404")) {
    window.history.replaceState({}, "", "/404");
  } else if (title.includes("400")) {
    window.history.replaceState({}, "", "/BadRequest");
  } else if (title.includes("500")) {
    window.history.replaceState({}, "", "/servererror");
  } else if (title.includes("429")) {
    window.history.replaceState({}, "", "/TooManyRequests");
  }

  // Reattach global listeners
  setupGlobalListeners();
}

export async function showPage(page, user, id, currentPage) {

  currentPage = page;
  

  switch (page) {
    case "unauthorized":
      showErrorPage({
        title: "403 - Not Authorized",
        message: "This endpoint is only accessible from inside the app.",
        statusCode: 403,
        statusText: "Forbidden"
      });
      break;

    case "badrequest":
      showErrorPage({
        title: "400 - Bad Request",
        message: "The request was invalid or cannot be processed.",
        statusCode: 400,
        statusText: "Bad Request"
      });
      break;

    case "servererror":
      showErrorPage({
        title: "500 - Internal Server Error",
        message: "Something went wrong on our side. Please try again later.",
        statusCode: 500,
        statusText: "Internal Server Error"
      });
      break;

    case "toomanyrequests":
      showErrorPage({
        title: "429 - Too Many Requests",
        message: "You have reached the maximum number of comments/posts allowed for today. Please try again later.",
        statusCode: 429,
        statusText: "Too Many Requests"
      });
      break;

    default:
      console.log("Page not handled in showPage:", page);
  }

  if (page === "login") {
    window.history.replaceState({}, "", "/login");
    await loginbuilding(user);
  } else if (page === "register") {
    window.history.replaceState({}, "", "/register");
    await registerBuilding(user);
  } else if (page === "home") {
    const res = await fetch("/api/me");
    const user = await res.json();
    window.history.replaceState({}, "", "/home");
    await homeBuild(user);
  } else if (page === "createpost") {
    await NewPost(Object.keys(user.fileds), user);
  } else if (page === "Post") {
    if (id !== undefined) {
      sessionStorage.setItem("currentPostId", id);
    } else {
      id = sessionStorage.getItem("currentPostId");
    }
    if (id) {

      POST(id, user);
    } else {
      console.error("No post ID available");
      await showPage("home", user);
    }
  }

}

// Function to start a private chat with a specific user
export async function startPrivateChat(targetUsername) {
  const res = await fetch("/api/me");
  const user = await res.json();
  if (!user.authenticated) {
    alert("You must be logged in to chat");
    return;
  }
  await showPage("chat", user, null, null, targetUsername);
  sessionStorage.setItem("currentPage", "chat");
}

window.onload = async () => {

  let user = { authenticated: false };
  try {
    const res = await fetch("/api/me");
    if (res.ok) user = await res.json();
  } catch (err) {
    console.error("Error fetching user:", err);
  }
  const path = window.location.pathname;
  const validPages = {
    "/": "home",
    "/home": "home",
    "/login": "login",
    "/register": "register",
    "/createpost": "createpost",
    "/Post": "Post",
    "/unauthorized": "unauthorized",
    "/BadRequest": "badrequest",
    "/servererror": "servererror",
    "/TooManyRequests": "toomanyrequests"
  };

  currentPage = validPages[path];

  if (!currentPage) {
    showErrorPage({
      title: "404 - Page not found",
      message: "This page does not exist."
    });
    return;
  }
  if (user.authenticated) {
    if (currentPage === "login" || currentPage === "register") {
      currentPage = "home";
    }
  } else {
    const authRequiredPages = ["home", "newpost", "Post"];
    if (authRequiredPages.includes(currentPage)) {
      currentPage = "login";
    }
  }
  await showPage(currentPage, user);
  setupGlobalListeners(user);

};
