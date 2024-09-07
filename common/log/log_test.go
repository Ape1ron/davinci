package log

import (
	"testing"
)

func TestLogger(t *testing.T) {

	defer Close()
	Output([][]string{{"title"}, {"data"}})
	Output("new output")
	Info("new info\nqq")
	Warn("new warn")
	Error("new error")

}
