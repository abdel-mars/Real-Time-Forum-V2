package utils

import (
	repo "forum/internal/repository"
	"fmt"
	"regexp"
	"time"
)

func InitRegex() {
	var err error
	repo.EmailExp, err = regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]+`)
	if err != nil {
		fmt.Println(err)
		return
	}

	repo.UsernameExp, err = regexp.Compile(`^[a-zA-Z]+[a-zA-Z0-9._]+`)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ValidUsername(s string) bool {
	if len(s) < repo.USERNAME_MIN_LEN || len(s) > repo.USERNAME_MAX_LEN {
		return false
	}
	return repo.UsernameExp.MatchString(s)
}

func ValidEmail(s string) bool {
	if len(s) < repo.EMAIL_MIN_LEN || len(s) > repo.EMAIL_MAX_LEN {
		return false
	}
	return repo.EmailExp.MatchString(s)
}

func ValidPassword(s string) bool {
	return len(s) >= repo.PASSWORD_MIN_LEN && len(s) <= repo.PASSWORD_MAX_LEN
}

func ValidComment(comment string) bool {
	return len(comment) >= repo.COMMENT_MIN_LEN && len(comment) <= repo.COMMENT_MAX_LEN
}

func ValidPost(content string) bool {
	return len(content) >= repo.POST_MIN_LEN && len(content) <= repo.POST_MAX_LEN
}

func ValidPostTitle(title string) bool {
	return len(title) >= repo.TITLE_MIN_LEN && len(title) <= repo.TITLE_MAX_LEN
}

func Contain(query string) bool {
	_, exists := repo.IT_MAJOR_FIELDS[query]
	return exists
}

func SqlDateFormater(date string) string {
	today := time.Now().Format("2006-01-02")
	if len(date) != 20 {
		return "date error"
	}
	tmp := date[:10]
	if tmp == today {
		tmp = "today"
		tmp = date[11:16]
	}
	return tmp
}
