// File: static/js/login.js

/* =========================
   Login Functionality
========================= */
function initLogin() {
    const loginForm = document.getElementById('loginForm');
    if (!loginForm) return; // Exit if login form is not present

    loginForm.addEventListener('submit', async function(event) {
        event.preventDefault();
        
        const usernameInput = document.getElementById('username');
        const passwordInput = document.getElementById('password');
        const errorDiv = document.querySelector('.error-message');

        const username = usernameInput ? usernameInput.value.trim() : '';
        const password = passwordInput ? passwordInput.value.trim() : '';

        if (!username || !password) {
            if (errorDiv) {
                errorDiv.textContent = 'Please enter both username and password.';
            }
            return;
        }

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password }),
                credentials: 'include' // Ensure cookies are included
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Login failed.');
            }

            // Redirect to consult room upon successful login
            window.location.href = 'consult.html';
        } catch (error) {
            console.error('Login error:', error);
            if (errorDiv) {
                errorDiv.textContent = error.message;
            }
        }
    });
}

export { initLogin };
