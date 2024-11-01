document.getElementById('add-memo').addEventListener('click', function() {
    const memoText = document.getElementById('memo-text').value;
    const memoDate = document.getElementById('memo-date').value;
    const memoType = document.getElementById('memo-type').value;

    if (memoText === '' || memoDate === '') {
      alert('Please enter memo details and a due date');
      return;
    }

    fetch('/add-memo', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: memoText, dueDate: memoDate, type: memoType })
    }).then(response => response.json())
      .then(data => {
        if (data.success) {
          addMemoToList(data.id, memoText, memoDate, memoType);
          document.getElementById('memo-text').value = '';
          document.getElementById('memo-date').value = '';
        }
      });
    });

    function addMemoToList(id, text, dueDate, type) {
    const memoList = document.getElementById('memo-list');
    const li = document.createElement('li');
    li.setAttribute('draggable', 'true');
    li.setAttribute('data-type', type);
    li.innerHTML = `
      <input type="checkbox" class="complete-checkbox">
      <span class="memo-text">${text}</span>
      <span class="memo-date">${dueDate}</span>
      <span class="memo-type">${type}</span>
      <button class="edit">Edit</button>
      <button class="delete">Delete</button>
    `;

    memoList.appendChild(li);

    // 支持拖動排序
    li.addEventListener('dragstart', function() {
      li.classList.add('dragging');
    });

    li.addEventListener('dragend', function() {
      li.classList.remove('dragging');
    });

    memoList.addEventListener('dragover', function(e) {
      e.preventDefault();
      const draggingItem = document.querySelector('.dragging');
      const afterElement = getDragAfterElement(memoList, e.clientY);
      if (afterElement == null) {
        memoList.appendChild(draggingItem);
      } else {
        memoList.insertBefore(draggingItem, afterElement);
      }
    });
  
    // 完成事項勾選
    li.querySelector('.complete-checkbox').addEventListener('change', function() {
      if (this.checked) {
        li.classList.add('completed');
      } else {
        li.classList.remove('completed');
      }
    });
  
    // 編輯與刪除功能
    li.querySelector('.edit').addEventListener('click', function() {
      openEditPopup(id, text, dueDate, type, li);
    });
  
    li.querySelector('.delete').addEventListener('click', function() {
      fetch(`/delete-memo/${id}`, { method: 'DELETE' })
        .then(response => response.json())
        .then(data => {
          if (data.success) {
            memoList.removeChild(li);
          }
        });
    });
  }
  
  // 用於拖動排序的輔助函數
  function getDragAfterElement(container, y) {
    const draggableElements = [...container.querySelectorAll('li:not(.dragging)')];
    return draggableElements.reduce((closest, child) => {
      const box = child.getBoundingClientRect();
      const offset = y - box.top - box.height / 2;
      if (offset < 0 && offset > closest.offset) {
        return { offset: offset, element: child };
      } else {
        return closest;
      }
    }, { offset: Number.NEGATIVE_INFINITY }).element;
  }
  
  function openEditPopup(id, text, dueDate, type, listItem) {
    const popup = document.getElementById('edit-popup');
    popup.classList.remove('hidden');
    document.getElementById('edit-memo-text').value = text;
    document.getElementById('edit-memo-date').value = dueDate;
    document.getElementById('edit-memo-type').value = type;
  
    document.getElementById('save-edit').onclick = function() {
      const newText = document.getElementById('edit-memo-text').value;
      const newDate = document.getElementById('edit-memo-date').value;
      const newType = document.getElementById('edit-memo-type').value;
  
      fetch(`/edit-memo/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: newText, dueDate: newDate, type: newType })
      }).then(response => response.json())
        .then(data => {
          if (data.success) {
            listItem.querySelector('.memo-text').textContent = newText;
            listItem.querySelector('.memo-date').textContent = newDate;
            listItem.querySelector('.memo-type').textContent = newType;
            popup.classList.add('hidden');
          }
        });
    };
  
    document.getElementById('cancel-edit').onclick = function() {
      popup.classList.add('hidden');
    };
  }
  // 修改代辦事項類型後，更新 Memo 的顏色
function updateMemoType(id, newType) {
  const memoItem = document.querySelector(`[data-id='${id}']`);
  // memoItem.classList.remove('Urgent', 'Important', 'Normal');
  // memoItem.classList.add(newType);
  memoItem.setAttribute('data-type', newType);
  console.log(memoItem)
}

// 保存代辦事項的編輯，並動態更新 Memo 外框顏色
document.getElementById('save-edit').addEventListener('click', () => {
  const memoId = document.getElementById('memoId').value;
  const newType = document.querySelector('input[name="type"]:checked').value;
  console.log(newType)
  updateMemoType(memoId, newType); // 動態更新顏色
});

