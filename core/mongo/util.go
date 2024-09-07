package mongo

import (
	"errors"
)

func findCloseBrackets(content string) (int, error) {
	stack := make([]rune, 0)

	if len(content) < 2 || content[0] != '(' {
		return -1, errors.New("content is not viald")
	}
	stack = append(stack, '(')

	for i := 1; i < len(content); i++ {
		c := content[i]
		top := stack[len(stack)-1]
		switch c {
		case '"':
			if top == '"' {
				stack = stack[:len(stack)-1]
			} else if top != '\'' {
				stack = append(stack, '"')
			}
		case '\'':
			if top == '\'' {
				stack = stack[:len(stack)-1]
			} else if top != '"' {
				stack = append(stack, '\'')
			}
		case '(':
			if top != '"' && top != '\'' {
				stack = append(stack, '(')
			}
		case ')':
			if top == '(' {
				stack = stack[:len(stack)-1]
			}
		case '\\':
			i++
		}
		if len(stack) == 0 {
			return i, nil
		}
	}
	return -1, errors.New("can't find")

}
