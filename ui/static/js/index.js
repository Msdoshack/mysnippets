function disableButton() {
  // Disable the submit button to prevent multiple submissions
  document.getElementById("submit-button").disabled = true;
  // Allow the form to be submitted
  return true;
}

function showCustomAlert(message) {
  const customAlert = document.getElementById("custom-alert");
  const alertMessage = document.getElementById("alert-message");

  alertMessage.textContent = message;
  customAlert.style.display = "block"; // Show the alert

  // Automatically close the alert after 5 seconds
  const autoCloseTimeout = setTimeout(function () {
    customAlert.style.display = "none"; // Hide the alert
  }, 4000); // 5000 milliseconds = 5 seconds
}

function resetPage() {
  document.getElementById("currentPage").value = 1;
}

document.addEventListener("DOMContentLoaded", function () {
  // Get the current URL
  const currentUrl = window.location.href;

  // Create a new URL object
  const url = new URL(currentUrl);

  // Use the searchParams property to get the query parameters
  const searchParams = url.searchParams;

  // Get the value of the 'language' parameter
  const language = searchParams.get("language");
  const title = searchParams.get("title");
  const hamburger = document.getElementById("hamburger");
  const navContent = document.getElementById("mobileNav");
  const mobileContent = document.getElementById("mobileContent");
  const delModal = document.getElementById("delModal");
  const delSnippetBtn = document.getElementById("delSnippetBtn");
  const cancelDelBtn = document.getElementById("cancelDelBtn");
  const searchResult = document.getElementById("searchResult");

  if (searchResult) {
    if (language) {
      searchResult.classList.add("show");
      searchResult.textContent += `result for: ${language}`;
    } else if (title) {
      searchResult.classList.add("show");
      searchResult.textContent += `result for: ${title}`;
    }
  }

  if (hamburger) {
    hamburger.addEventListener("click", function (e) {
      // navContent.classList.toggle("show");
      e.stopPropagation(); // Prevent click event from propagating to document
      navContent.classList.toggle("show");
    });
  }

  // Close nav menu when clicking outside of it
  document.addEventListener("click", function (event) {
    // Check if the click was outside the navContent and the hamburger
    if (!mobileContent.contains(event.target)) {
      navContent.classList.remove("show"); // Remove the 'show' class to hide the menu
    }
  });

  if (delSnippetBtn) {
    delSnippetBtn.addEventListener("click", function () {
      delModal.classList.add("show");
    });
  }

  if (cancelDelBtn) {
    cancelDelBtn.addEventListener("click", function () {
      delModal.classList.remove("show");
    });
  }

  const copyCodeBtn = document.getElementById("copyCodeBtn");
  const copiedIcon = document.getElementById("copiedCodeBtn");

  if (copyCodeBtn) {
    copyCodeBtn.addEventListener("click", function () {
      const snippetContent = document.getElementById("snippet").textContent;
      // Use the Clipboard API
      navigator.clipboard
        .writeText(snippetContent)
        .then(() => {
          copyCodeBtn.classList.remove("show");
          copiedIcon.classList.add("show");
        })
        .catch((err) => {
          console.error("Could not copy text: ", err);
        });
    });
  }
});
