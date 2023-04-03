package clipboard

import (
	"context"
	"fmt"
	"golang.design/x/clipboard"
	"testing"
)

func TestCpoy(t *testing.T) {
	cxt := context.Background()
	defer cxt.Done()
	textCh := clipboard.Watch(cxt, clipboard.FmtText)
	//cpf.Write(clipboard.FmtText, []byte("测试一下"))
	select {
	case ttt := <-textCh:
		fmt.Println(string(ttt))
	}
}
