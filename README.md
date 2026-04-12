# ThreadsMarketing

因為小弟我只會開發不懂行銷，所以打算用我能理解的流程去測試與打造一個全自動的 Threads 行銷發文 Agent<br>
會自己去爬 Hacker News、RSS 找最新資料，丟給 Claude 寫貼文，並自動發到 Threads 上並持續追蹤每篇貼文的互動數據<br>
哪篇貼文爆了，從中分析為什麼爆，並固定時間分析本日發文成果檢討冷門貼文可能原因，並把學到的東西寫回 DNA 文檔持續成長<br>
有沒有用不知道，就是用來學著做行銷

## 2026/04/13
- 啟動時自動驗證 Threads token 有效性，失敗時清除 keychain 並提示重新登入
- `rewriteLog` 改用 `app.QueueUpdateDraw` 包裝，避免 goroutine 競態

## 2026/04/12
- 建立 TUI 介面（tview），含 dashboard 與指令輸入區
- 新增互動式 login 流程（App ID → App Secret → Short-lived Token → 交換 Long-lived Token）
- 新增 Threads API client（`internal/threads/auth.go`）：token 交換與驗證
- Token 與 user_id 儲存至 macOS Keychain（fallback 檔案）
