package main

import (
	"fmt"
	"time"
)

func ReminderService() {
	for {
		now := time.Now()
		var memos []Memo
		db.Where("reminder_time IS NOT NULL").Find(&memos)
		for _, memo := range memos {
			if memo.ReminderTime != nil {
				diff := memo.ReminderTime.Sub(now)
				if diff <= time.Minute && diff > 0 {
					SendReminder(memo)
				}
			}
		}
		time.Sleep(time.Minute)
	}
}

func SendReminder(memo Memo) {
	// 实现通知功能，例如打印到控制台
	fmt.Printf("提醒：%s - %s\n", memo.Title, memo.Content)
}
