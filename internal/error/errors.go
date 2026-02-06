// package error
package forumerror

import (
	repo "forum/internal/repository"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Error struct {
	StatusCode       int
	StatusText       string
	ErrorMessage     string
	ErrorTitle       string
	ErrorDescription string
}

func NotFoundError(w http.ResponseWriter, r *http.Request) {
	notFoundError := Error{
		StatusCode:       http.StatusNotFound,
		StatusText:       "Not Found",
		ErrorMessage:     "The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.",
		ErrorTitle:       "Oops! Page Not Found",
		ErrorDescription: "We couldn't find the page you were looking for. Please check the URL for any mistakes or go back to the homepage.",
	}

	w.WriteHeader(notFoundError.StatusCode)
	tmplErr := repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "error.html", notFoundError)
	if tmplErr != nil {
		fmt.Fprintf(w, "%s", notFoundError.StatusText)
	}
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed := Error{
		StatusCode:       http.StatusMethodNotAllowed,
		StatusText:       "Method not allowed",
		ErrorMessage:     "The HTTP method used for this request is not allowed on this resource. Please use a different method.",
		ErrorTitle:       "Oops! method not allowed",
		ErrorDescription: "Please send a POST request to this endpoint with the required data for processing. Other HTTP methods are not allowed.",
	}

	w.WriteHeader(methodNotAllowed.StatusCode)
	tmplErr := repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "error.html", methodNotAllowed)
	if tmplErr != nil {
		fmt.Fprintf(w, "%s", methodNotAllowed.StatusText)
	}
}

func InternalServerError(w http.ResponseWriter, r *http.Request, forumErr error) {
	internalServerError := Error{
		StatusCode:       http.StatusInternalServerError,
		StatusText:       "Internal Server Error",
		ErrorMessage:     "An unexpected error occurred while processing your request.",
		ErrorTitle:       "Oops! Internal Server Error",
		ErrorDescription: "Something went wrong on our end. Please try again later or contact support if the issue persists.",
	}

	logFile, err := os.OpenFile(
		repo.INTERNAL_SERVER_ERROR_LOG_PATH,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Printf("Failed to open error log file: %v", err)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.Printf("[%s] %s %s\nError: %v",
			time.Now().Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			forumErr.Error(),
		)
	}

	w.WriteHeader(internalServerError.StatusCode)

	tmplErr := repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "error.html", internalServerError)
	if tmplErr != nil {
		fmt.Fprintf(w, "%s", internalServerError.StatusText)
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	badRequest := Error{
		StatusCode:       http.StatusBadRequest,
		StatusText:       "Bad Request",
		ErrorMessage:     "The Server recieved a Bad request!!",
		ErrorTitle:       "Oops! Bad Request",
		ErrorDescription: "Please make sure you respect the input limits",
	}

	w.WriteHeader(badRequest.StatusCode)
	tmplErr := repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "error.html", badRequest)
	if tmplErr != nil {
		fmt.Fprintf(w, "%s", badRequest.StatusText)
	}
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	unauthorizedError := Error{
		StatusCode:       http.StatusUnauthorized,
		StatusText:       "Unauthorized",
		ErrorMessage:     "You are not authorized to access this resource. Authentication is required.",
		ErrorTitle:       "Access Denied",
		ErrorDescription: "Please log in with valid credentials to access this endpoint. If you believe this is a mistake, contact support.",
	}

	w.WriteHeader(unauthorizedError.StatusCode)
	tmplErr := repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "error.html", unauthorizedError)
	if tmplErr != nil {
		fmt.Fprintf(w, "%s", unauthorizedError.StatusText)
	}
}

func TooManyRequests(w http.ResponseWriter, r *http.Request, option string) {
	postLimitError := Error{
		StatusCode:       http.StatusTooManyRequests,
		StatusText:       "Too Many Requests",
		ErrorMessage:     "You have reached the maximum number of " + option + " allowed per day.",
		ErrorTitle:       option + " Limit Reached",
		ErrorDescription: "You’ve hit your daily " + option + " limit. Try again tomorrow or contact support if this seems incorrect.",
	}

	w.WriteHeader(postLimitError.StatusCode)
	tmplErr := repo.GLOBAL_TEMPLATE.ExecuteTemplate(w, "error.html", postLimitError)
	if tmplErr != nil {
		fmt.Fprintf(w, "%s", postLimitError.StatusText)
	}
}
