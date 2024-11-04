// static/app.js

// 格式化日期时间函数
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

// 当前选择的排序方式
let currentSortBy = "time";

// 获取备忘录列表
async function fetchMemos() {
    const response = await fetch(`/api/memos?sort_by=${currentSortBy}`);
    if (response.ok) {
        const memos = await response.json();

        const memoList = document.getElementById('memo-list');
        memoList.innerHTML = '';
        memos.forEach(memo => {
            // 处理提醒时间
            let reminderTimeText = '未設定';
            let reminderTimeClass = '';
            if (memo.reminder_time) {
                const reminderTime = new Date(memo.reminder_time);
                reminderTimeText = formatDateTimeDisplay(reminderTime);

                // 检查是否小于1天
                const now = new Date();
                const diffTime = reminderTime - now;
                const diffDays = diffTime / (1000 * 60 * 60 * 24);
                if (diffDays < 1) {
                    reminderTimeClass = 'text-danger';
                }
            }

            // 处理完成状态
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

            // 渲染备忘录
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

// 初始化拖放功能
function initializeDragAndDrop() {
    const memoList = document.getElementById('memo-list');
    if (memoList.getAttribute('data-sortable-initialized') === 'true') {
        // 已初始化，避免重复初始化
        return;
    }

    Sortable.create(memoList, {
        animation: 150,
        handle: '.card', // 可以选择拖动的区域，例如整个卡片
        onEnd: function (evt) {
            // 获取新的排序顺序
            const sortedItems = Array.from(memoList.children).map(child => {
                // 从按钮的 onclick 属性中提取 memo.id
                const completeButton = child.querySelector('.btn-outline-success, .btn-outline-warning');
                const onclickAttr = completeButton.getAttribute('onclick');
                const idMatch = onclickAttr.match(/toggleCompleteMemo\((\d+)\)/);
                return idMatch ? parseInt(idMatch[1]) : null;
            }).filter(id => id !== null);

            // 构建排序数据
            const sortData = sortedItems.map((id, index) => {
                return {
                    id: id,
                    sort_order: index + 1
                };
            });

            // 发送排序数据到后端
            updateMemosSort(sortData);

            // 自动切换排序选单为“自定义排序”
            const sortSelect = document.getElementById('sort-options');
            sortSelect.value = 'custom';
            currentSortBy = 'custom';
        }
    });

    memoList.setAttribute('data-sortable-initialized', 'true');
}

// 发送排序数据到后端
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

// 新增备忘录
document.getElementById('create-memo')?.addEventListener('click', () => {
    openMemoModal();
});

// 监听排序选单的变化
document.getElementById('sort-options').addEventListener('change', (e) => {
    currentSortBy = e.target.value;
    fetchMemos();
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

// 新增函数：切换完成状态
async function toggleCompleteMemo(id) {
    const response = await fetch(`/api/memos/${id}/complete`, {
        method: 'POST'
    });
    if (response.ok) {
        const data = await response.json();
        fetchMemos();
    } else {
        alert('更新完成状态失败');
    }
}

// 初始化加载备忘录列表
fetchMemos();
