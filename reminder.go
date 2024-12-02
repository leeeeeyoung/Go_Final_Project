package main

import (
	"fmt"
	"net/smtp"
	"time"
)

func SendEmail(to, subject, body string) error {
	from := "son60712@gmail.com"
	password := "fbdg yrrm aclv pouv" //應用程式密碼

	// Gmail SMTP 設定
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 建立郵件訊息
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	return err
}

func ReminderService() {
	for {
		now := time.Now()
		var memos []Memo

		// 查詢所有有提醒時間的備忘錄，並載入用戶資料
		db.Preload("User").Where("reminder_time IS NOT NULL AND completed = ?", false).Find(&memos)

		for _, memo := range memos {
			if memo.ReminderTime != nil {
				diff := memo.ReminderTime.Sub(now)
				if diff <= time.Minute && diff > 0 {
					err := SendReminder(memo)
					if err != nil {
						fmt.Printf("無法發送提醒: %s - %v\n", memo.Title, err)
					}
				}
			}
		}

		time.Sleep(time.Minute)
	}
}

func SendReminder(memo Memo) error {
	if memo.UserID == 0 || memo.User.Email == "" {
		return fmt.Errorf("用戶無電子郵件地址或未關聯用戶")
	}

	subject := fmt.Sprintf("提醒: %s", memo.Title)
	body := fmt.Sprintf("內容: %s\n提醒時間: %s", memo.Content, memo.ReminderTime.Format(time.RFC1123))
	return SendEmail(memo.User.Email, subject, body)
}
