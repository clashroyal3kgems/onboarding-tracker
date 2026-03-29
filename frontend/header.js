(function() {
    function initHeader() {
        const headerContainer = document.getElementById("main-header");
        if (!headerContainer) return;

        const role = localStorage.getItem("user_role");
        const userId = localStorage.getItem("user_id");

        // Защита: если нет ID — на логин (кроме самой страницы логина)
        if (!userId && !window.location.pathname.includes("login.html")) {
            window.location.href = "/login.html";
            return;
        }

        // Исправленные пути согласно твоей структуре
        let navLinks = '';
        if (role === 'admin') {
            navLinks = `
                <a href="/materials.html">Материалы</a>
                <a href="/onboarding-plans.html">Планы</a>
            `;
        } else if (role === 'mentor') {
            navLinks = `<a href="/onboarding-plans.html">Подопечные</a>`;
        } else {
            navLinks = `<a href="/onboarding-my.html">Мой план</a>`;
        }

        // Вставляем HTML и минималистичные стили
        headerContainer.innerHTML = `
            <style>
                .nav-wrapper {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    padding: 0.8rem 2rem;
                    background: #ffffff;
                    border-bottom: 1px solid #eaeaea;
                    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
                }
                .nav-left { display: flex; gap: 20px; align-items: center; }
                .nav-left strong { color: #007bff; letter-spacing: 1px; }
                .nav-left a { 
                    text-decoration: none; 
                    color: #666; 
                    font-size: 0.9rem; 
                    transition: color 0.2s;
                }
                .nav-left a:hover { color: #000; }
                .nav-right { display: flex; align-items: center; gap: 15px; font-size: 0.85rem; color: #888; }
                .logout-link { 
                    color: #dc3545; 
                    text-decoration: none; 
                    cursor: pointer; 
                    font-weight: 500;
                    border: 1px solid #dc3545;
                    padding: 4px 12px;
                    border-radius: 4px;
                    transition: all 0.2s;
                }
                .logout-link:hover { background: #dc3545; color: white; }
            </style>
            <nav class="nav-wrapper">
                <div class="nav-left">
                    <strong>ONBOARD</strong>
                    ${navLinks}
                </div>
                <div class="nav-right">
                    <span>${role.toUpperCase()} #${userId}</span>
                    <a id="logout-btn" class="logout-link">Выход</a>
                </div>
            </nav>
        `;

        document.getElementById("logout-btn").onclick = (e) => {
            e.preventDefault();
            localStorage.clear();
            window.location.href = "/login.html";
        };
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", initHeader);
    } else {
        initHeader();
    }
})();