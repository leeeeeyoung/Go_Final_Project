// 新增格式化函数
function formatDateTimeLocal(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0'); // 月份从0开始
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
}

function formatDateTimeDisplay(date) {
    // 格式化日期时间为 YYYY-MM-DD HH:MM
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0'); // 月份从0开始
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}`;
}

async function fetchMemos() {
    const response = await fetch('/api/memos');
    if (response.ok) {
        const memos = await response.json();
        const memoList = document.getElementById('memo-list');
        memoList.innerHTML = '';
        memos.forEach(memo => {
            const memoItem = document.createElement('div');
            memoItem.className = 'col-md-4 memo-item';
            if (memo.type === 'important') {
                memoItem.classList.add('important');
            }
            let reminderTimeText = '未設定';
            let reminderTimeClass = '';
            if (memo.reminder_time) {
                const reminderTime = new Date(memo.reminder_time);
                reminderTimeText = formatDateTimeDisplay(reminderTime);

                // 檢查是否小於1天
                const now = new Date();
                const diffTime = reminderTime - now;
                const diffDays = diffTime / (1000 * 60 * 60 * 24);
                if (diffDays < 1) {
                    reminderTimeClass = 'text-danger';
                }
            }

            memoItem.innerHTML = `
                <div class="card">
                    <div class="card-body">
                        <h5 class="card-title">${memo.title}</h5>
                        <p class="card-text">${memo.content}</p>
                        <p class="card-text">
                            <small class="${reminderTimeClass}">提醒時間：${reminderTimeText}</small>
                        </p>
                        <button class="btn btn-sm btn-outline-primary" onclick="editMemo(${memo.id})">編輯</button>
                        <button class="btn btn-sm btn-outline-danger" onclick="deleteMemo(${memo.id})">刪除</button>
                    </div>
                </div>
            `;
            memoList.appendChild(memoItem);
        });
    } else {
        window.location.href = '/login';
    }
}

document.getElementById('logout')?.addEventListener('click', async () => {
    await fetch('/api/logout', { method: 'POST' });
    window.location.href = '/login';
});

document.getElementById('create-memo')?.addEventListener('click', () => {
    openMemoModal();
});

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
        if (memo.id) {
            updateMemo(memo.id);
        } else {
            createMemo();
        }
        closeMemoModal();
    };
}

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

async function editMemo(id) {
    const response = await fetch(`/api/memos`);
    const memos = await response.json();
    const memo = memos.find(m => m.id === id);
    openMemoModal(memo);
}

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

async function deleteMemo(id) {
    await fetch(`/api/memos/${id}`, {
        method: 'DELETE'
    });
    fetchMemos();
}

fetchMemos();
