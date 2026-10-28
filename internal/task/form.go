package task

import (
	"strings"
	"unicode/utf8"
)

type taskForm struct {
	UserID int64
	Title string
	Completed bool
	Errors   map[string]string
}

func (f *taskForm) validate() bool {
    f.Errors = make(map[string]string)

    title := strings.TrimSpace(f.Title)
    if title == "" {
        f.addError("title", "Title cannot be empty.")
    } else if utf8.RuneCountInString(title) < 5 {
        f.addError("title", "Must be at least 5 characters.")
    } else if utf8.RuneCountInString(title) > 60 {
        f.addError("title", "Cannot exceed 60 characters.")
    }

    return len(f.Errors) == 0
}

func (f *taskForm) addError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[field] = message
}
