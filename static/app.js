// static/app.js

// 格式化日期時間為本地格式（YYYY-MM-DDTHH:MM，用於表單）
function formatDateTimeLocal(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
}

// 格式化日期時間為顯示格式（YYYY-MM-DD HH:MM，用於顯示）
function formatDateTimeDisplay(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}`;
}

// 用於記錄當前的備忘錄排序方式
let currentSortBy = "time";

// 獲取備忘錄列表，並更新顯示
async function fetchMemos() {
    const response = await fetch(`/api/memos?sort_by=${currentSortBy}`);
    if (response.ok) {
        const memos = await response.json();

        const memoList = document.getElementById('memo-list');
        memoList.innerHTML = '';
        memos.forEach(memo => {
            // 生成提醒時間顯示
            let reminderTimeText = '未設定';
            let reminderTimeClass = '';
            if (memo.reminder_time) {
                const reminderTime = new Date(memo.reminder_time);
                reminderTimeText = formatDateTimeDisplay(reminderTime);

                // 如果時間小於一天，顯示警告樣式
                const now = new Date();
                const diffTime = reminderTime - now;
                const diffDays = diffTime / (1000 * 60 * 60 * 24);
                if (diffDays < 1) {
                    reminderTimeClass = 'text-danger';
                }
            }

            // 處理完成狀態
            let completedClass = '';
            let titleStyle = '';
            let completeButtonClass = 'btn-outline-success';
            let completeButtonText = '完成';
            if (memo.completed) {
                completedClass = 'completed';
                titleStyle = 'text-decoration: line-through; color: #6c757d;';
                completeButtonClass = 'btn-outline-warning';
                completeButtonText = '復原';
            }

            // 渲染備忘錄項目
            const memoItem = document.createElement('div');
            memoItem.className = 'col-md-4 memo-item';
            if (memo.type === 'important') {
                memoItem.classList.add('important');
            }
            if (memo.completed) {
                memoItem.classList.add('completed');
            }

            memoItem.innerHTML = `
                <div class="card">
                    <div class="card-body d-flex flex-column">
                        <h5 class="card-title" style="${titleStyle}">${memo.title}</h5>
                        <p class="card-text">${memo.content}</p>
                        <p class="card-text">
                            <small class="${reminderTimeClass}">提醒時間：${reminderTimeText}</small>
                        </p>
                        <div class="button-group mt-auto">
                            <button class="btn btn-sm btn-outline-primary mr-2" onclick="editMemo(${memo.id})">
                                <i class="fas fa-edit"></i> 編輯
                            </button>
                            <button class="btn btn-sm btn-outline-danger mr-2" onclick="deleteMemo(${memo.id})">
                                <i class="fas fa-trash-alt"></i> 刪除
                            </button>
                            <button class="btn btn-sm ${completeButtonClass}" onclick="toggleCompleteMemo(${memo.id})">
                                ${completeButtonText}
                            </button>
                        </div>
                    </div>
                </div>
            `;
            memoList.appendChild(memoItem);
        });

        // 初始化或更新拖放功能
        initializeDragAndDrop();
    } else {
        window.location.href = '/login';
    }
}

// 初始化備忘錄拖放功能
function initializeDragAndDrop() {
    const memoList = document.getElementById('memo-list');
    if (memoList.getAttribute('data-sortable-initialized') === 'true') {
        return;
    }

    Sortable.create(memoList, {
        animation: 150,
        handle: '.card',
        onEnd: function (evt) {
            // 取得新的排序順序
            const sortedItems = Array.from(memoList.children).map(child => {
                // 从按钮的 onclick 属性中提取 memo.id
                const completeButton = child.querySelector('.btn-outline-success, .btn-outline-warning');
                const onclickAttr = completeButton.getAttribute('onclick');
                const idMatch = onclickAttr.match(/toggleCompleteMemo\((\d+)\)/);
                return idMatch ? parseInt(idMatch[1]) : null;
            }).filter(id => id !== null);

            // 構建排序數據
            const sortData = sortedItems.map((id, index) => {
                return {
                    id: id,
                    sort_order: index + 1
                };
            });

            // 更新排序數據
            updateMemosSort(sortData);

            // 切換排序選單為自訂
            const sortSelect = document.getElementById('sort-options');
            sortSelect.value = 'custom';
            currentSortBy = 'custom';
        }
    });

    memoList.setAttribute('data-sortable-initialized', 'true');
}

// 發送排序數據到後端
async function updateMemosSort(sortData) {
    const response = await fetch('/api/memos/sort', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sortData)
    });

    if (response.ok) {
        console.log('排序更新成功');
    } else {
        alert('排序更新失败');
    }
}

// 登出功能
document.getElementById('logout')?.addEventListener('click', async () => {
    await fetch('/api/logout', { method: 'POST' });
    window.location.href = '/login';
});

// 新增備忘錄
document.getElementById('create-memo')?.addEventListener('click', () => {
    openMemoModal();
});

// 排序選單變更監聽器
document.getElementById('sort-options').addEventListener('change', (e) => {
    currentSortBy = e.target.value;
    fetchMemos();
});

// 開啟備忘錄彈窗
function openMemoModal(memo = {}) {
    const modal = document.getElementById('memo-modal');
    modal.classList.add('show');
    document.getElementById('modal-title').textContent = memo.id ? '編輯備忘錄' : '新增備忘錄';
    document.getElementById('memo-title').value = memo.title || '';
    document.getElementById('memo-content').value = memo.content || '';
    document.getElementById('memo-type').value = memo.type || 'normal';
    document.getElementById('memo-reminder-time').value = memo.reminder_time ? formatDateTimeLocal(new Date(memo.reminder_time)) : '';
    document.getElementById('memo-form').onsubmit = (e) => {
        e.preventDefault();
        
        // 檢查提醒時間
        const reminderTimeInput = document.getElementById('memo-reminder-time').value;
        const reminderTime = new Date(reminderTimeInput);
        const now = new Date();

        // 檢查提醒時間是否為未來時間
        if (isNaN(reminderTime.getTime()) || reminderTime <= now) {
            alert("提醒時間無效，必須設置為今天之後");
            return;
        }
        
        if (memo.id) {
            updateMemo(memo.id);
        } else {
            createMemo();
        }
        closeMemoModal();
    };
}

// 關閉備忘錄彈窗
function closeMemoModal() {
    const modal = document.getElementById('memo-modal');
    modal.classList.remove('show');
}

window.onclick = function(event) {
    const modal = document.getElementById('memo-modal');
    if (event.target == modal) {
        closeMemoModal();
    }
};

// 新增備忘錄功能
async function createMemo() {
    const title = document.getElementById('memo-title').value;
    const content = document.getElementById('memo-content').value;
    const type = document.getElementById('memo-type').value;
    const reminderTime = document.getElementById('memo-reminder-time').value;

    await fetch('/api/memos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            title,
            content,
            type,
            reminder_time_str: reminderTime
        })
    });
    fetchMemos();
}

// 編輯備忘錄功能
async function editMemo(id) {
    const response = await fetch(`/api/memos`);
    const memos = await response.json();
    const memo = memos.find(m => m.id === id);
    openMemoModal(memo);
}

// 更新備忘錄功能
async function updateMemo(id) {
    const title = document.getElementById('memo-title').value;
    const content = document.getElementById('memo-content').value;
    const type = document.getElementById('memo-type').value;
    const reminderTime = document.getElementById('memo-reminder-time').value;

    await fetch(`/api/memos/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            title,
            content,
            type,
            reminder_time_str: reminderTime
        })
    });
    fetchMemos();
}

// 刪除備忘錄功能
async function deleteMemo(id) {
    const confirmed = confirm("確定要刪除這個備忘錄嗎？這個操作無法撤回。");
    if (confirmed) {
        await fetch(`/api/memos/${id}`, {
            method: 'DELETE'
        });
        fetchMemos();
    }
}

// 切換完成狀態功能
async function toggleCompleteMemo(id) {
    const response = await fetch(`/api/memos/${id}/complete`, {
        method: 'POST'
    });
    if (response.ok) {
        const data = await response.json();
        fetchMemos();
    } else {
        alert('更新完成狀態失敗');
    }
}

// 初始化時載入備忘錄列表
fetchMemos();
