package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Memo struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	DueDate   string `json:"dueDate"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"` // 新增類型屬性
}

var memos []Memo
var idCounter int

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/add-memo", addMemoHandler)
	http.HandleFunc("/edit-memo/", editMemoHandler)
	http.HandleFunc("/delete-memo/", deleteMemoHandler)
	http.ListenAndServe(":8080", nil)
}

func addMemoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var newMemo Memo
		err := json.NewDecoder(r.Body).Decode(&newMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		idCounter++
		newMemo.ID = idCounter
		newMemo.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		memos = append(memos, newMemo)

		response := map[string]interface{}{
			"success":   true,
			"id":        newMemo.ID,
			"timestamp": newMemo.Timestamp,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func editMemoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		idStr := r.URL.Path[len("/edit-memo/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var updatedMemo Memo
		err = json.NewDecoder(r.Body).Decode(&updatedMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for i := range memos {
			if memos[i].ID == id {
				memos[i].Text = updatedMemo.Text
				memos[i].DueDate = updatedMemo.DueDate
				memos[i].Type = updatedMemo.Type // 更新類型屬性
				break
			}
		}

		response := map[string]interface{}{
			"success": true,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func deleteMemoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		idStr := r.URL.Path[len("/delete-memo/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for i := range memos {
			if memos[i].ID == id {
				memos = append(memos[:i], memos[i+1:]...)
				break
			}
		}

		response := map[string]interface{}{
			"success": true,
		}
		json.NewEncoder(w).Encode(response)
	}
}
