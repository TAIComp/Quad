// File: static/js/navigationBar.js

/* =========================
   Navigation Bar Interactions
========================= */
function initNavigationBar() {
    const mobileMenuBtn = document.querySelector('.mobile-menu-btn');
    const navElements = document.querySelector('.nav-elements');
    const navLinks = document.querySelectorAll('.nav-link');
    let lastScroll = 0;

    if (mobileMenuBtn && navElements) {
        // Mobile menu toggle
        mobileMenuBtn.addEventListener('click', function() {
            this.classList.toggle('active');
            navElements.classList.toggle('active');
        });
    }

    if (navLinks.length > 0) {
        // Active link handling
        navLinks.forEach(link => {
            link.addEventListener('click', function() {
                navLinks.forEach(l => l.classList.remove('active'));
                this.classList.add('active');

                // Close mobile menu after clicking a link
                if (navElements.classList.contains('active')) {
                    navElements.classList.remove('active');
                    if (mobileMenuBtn) {
                        mobileMenuBtn.classList.remove('active');
                    }
                }
            });
        });
    }

    // Scroll handling for navbar (hide on scroll down, show on scroll up)
    window.addEventListener('scroll', function() {
        const navbar = document.querySelector('.navbar');
        const currentScroll = window.pageYOffset;

        if (currentScroll <= 0) {
            navbar.classList.remove('scroll-up');
            return;
        }

        if (currentScroll > lastScroll && !navbar.classList.contains('scroll-down')) {
            navbar.classList.remove('scroll-up');
            navbar.classList.add('scroll-down');
        } else if (currentScroll < lastScroll && navbar.classList.contains('scroll-down')) {
            navbar.classList.remove('scroll-down');
            navbar.classList.add('scroll-up');
        }
        lastScroll = currentScroll;
    });
}
