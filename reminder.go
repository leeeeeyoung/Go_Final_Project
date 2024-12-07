package main

import (
	"fmt"
	"net/smtp"
	"time"
)

// SendEmail 發送電子郵件
// 參數: 接收者地址 (to)、主旨 (subject)、內容 (body)
func SendEmail(to, subject, body string) error {
	from := "son60712@gmail.com"
	password := "fbdg yrrm aclv pouv" //應用程式密碼

	// Gmail SMTP 設定
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 建立郵件訊息
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)

	// 使用 SMTP 驗證並發送郵件
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	return err
}

// ReminderService 定期檢查備忘錄，若需要提醒則發送郵件
func ReminderService() {
	for {
		now := time.Now()
		var memos []Memo

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

// SendReminder 發送備忘錄的提醒郵件
// 參數: 備忘錄資料 (memo)
func SendReminder(memo Memo) error {
	if memo.UserID == 0 || memo.User.Email == "" {
		return fmt.Errorf("用戶無電子郵件地址或未有關聯用戶")
	}

	subject := fmt.Sprintf("提醒: %s", memo.Title)
	body := fmt.Sprintf("內容: %s\n提醒時間: %s", memo.Content, memo.ReminderTime.Format(time.RFC1123))
	return SendEmail(memo.User.Email, subject, body)
}
