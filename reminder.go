package main

import (
	"fmt"
	"time"
)

func ReminderService() {
	for {
		now := time.Now()
		for _, memo := range memos {
			if memo.ReminderTime.IsZero() {
				continue
			}
			if memo.ReminderTime.Sub(now) <= time.Minute && memo.ReminderTime.Sub(now) > 0 {
				SendReminder(memo)
			}
		}
		time.Sleep(time.Minute)
	}
}

func SendReminder(memo Memo) {
	// 實現通知功能，例如打印到控制台
	fmt.Printf("提醒：%s - %s\n", memo.Title, memo.Content)
}
