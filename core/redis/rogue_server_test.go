package redis

import (
	"fmt"
	"testing"
)

// 注：写入内容长度过短会导致redis崩溃，这应该和redis的处理有关系，<=8会崩溃
func TestName(t *testing.T) {
	data := "12345678"
	rogueServer := CreateRogueserver(6366, []byte(data))
	fmt.Println("length: ", len([]byte(data)))
	done := make(chan struct{})
	rogueServer.Handle(done)
	<-done
}
