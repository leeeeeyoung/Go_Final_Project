function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    const hamburger = document.querySelector('.hamburger');
    
    if (sidebar.style.width === '250px') {
        sidebar.style.width = '0';
        hamburger.classList.remove('open');
    } else {
        sidebar.style.width = '250px';
        hamburger.classList.add('open');
    }
}

document.addEventListener("DOMContentLoaded", function() {
    fetch('/api/users')
        .then(response => response.json())
        .then(users => {
            const userList = document.getElementById('user-list');
            users.forEach(user => {
                const listItem = document.createElement('li');
                listItem.innerHTML = `${user.username}<br>(${user.email})<br>`;
                userList.appendChild(listItem);
            });
        })
        .catch(error => console.error('Error fetching users:', error));
});