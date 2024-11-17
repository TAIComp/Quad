// File: static/js/fadeOutTransitions.js

/* =========================
   Fade Out Navigation Transitions
========================= */
function initFadeOutTransitions() {
    function fadeOutAndNavigate(event) {
        const link = event.currentTarget;
        if (link.tagName.toLowerCase() !== 'a') return;

        event.preventDefault();
        document.body.classList.add('fade-out');
        setTimeout(() => {
            window.location.href = link.href;
        }, 1000); // Duration should match the CSS animation duration
    }

    // Attach to all internal links with class 'internal-link'
    document.querySelectorAll('a.internal-link').forEach(link => {
        link.addEventListener('click', fadeOutAndNavigate);
    });
}

// Add this export statement at the end of the file
export { initFadeOutTransitions };
