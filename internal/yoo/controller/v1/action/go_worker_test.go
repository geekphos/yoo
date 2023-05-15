package action

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var rows = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // 任务列表

func TestWorkers(t *testing.T) {
	var wg sync.WaitGroup

	userCount := 2
	ch := make(chan int, 2)
	closeCh := make(chan struct{})
	for i := 0; i < userCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range ch {
				if d == 2 {
					close(closeCh)
				}
				fmt.Printf("go func: %d, time: %d\n", d, time.Now().Unix())
				time.Sleep(time.Second * time.Duration(3))
			}
		}()
	}

	done := false

	for i := 0; i < 10 && !done; i++ {
		fmt.Println("hello ", i)
		select {
		case ch <- rows[i]:
			// 默认
		case <-closeCh:
			fmt.Print("退出循环")
			done = true
			// 如何退出该 for 循环
		}
		//time.Sleep(time.Second)
	}

	close(ch)
	wg.Wait()
}
