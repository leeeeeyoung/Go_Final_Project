// static/auth.js

// 處理登入表單提交
document.getElementById('login-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            username: document.getElementById('username').value,
            password: document.getElementById('password').value
        })
    });
    if (response.ok) {
        window.location.href = '/index';
    } else {
        alert('登入失敗');
    }
});

// 處理註冊表單提交
document.getElementById('register-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const response = await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            username: document.getElementById('username').value,
            password: document.getElementById('password').value,
            email: document.getElementById('email').value
        })
    });
    if (response.ok) {
        window.location.href = '/login';
    } else {
        alert('註冊失敗');
    }
});
