// File: static/js/missionVideo.js

/* =========================
   Mission Video Functionality
========================= */
function initMissionVideo() {
    // Ensure this script runs only on mission-video.html
    if (!window.location.pathname.endsWith('mission-video.html')) return;

    const fullscreenOverlay = document.querySelector('.fullscreen-video-overlay');
    const closeVideoBtn = document.querySelector('.close-video');
    const clickMessage = document.querySelector('.click-message');
    const videoElement = fullscreenOverlay ? fullscreenOverlay.querySelector('video') : null;
    const expertButton = document.querySelector('.expert-button');

    if (!fullscreenOverlay || !videoElement) return; // Exit if elements are missing

    // Function to show fullscreen video overlay
    function showVideo() {
        fullscreenOverlay.classList.add('active');
    }

    // Function to hide fullscreen video overlay
    function hideVideo() {
        fullscreenOverlay.classList.remove('active');
    }

    // Event listener for clicking to play video
    fullscreenOverlay.addEventListener('click', function(event) {
        if (event.target !== videoElement && !closeVideoBtn.contains(event.target)) {
            videoElement.play();
            clickMessage.style.display = 'none';
            if (expertButton) {
                expertButton.classList.add('visible');
            }
        }
    });

    // Event listener for close video button
    if (closeVideoBtn) {
        closeVideoBtn.addEventListener('click', function(event) {
            event.stopPropagation(); // Prevent triggering the overlay click
            videoElement.pause();
            videoElement.currentTime = 0;
            hideVideo();
            clickMessage.style.display = 'block';
            if (expertButton) {
                expertButton.classList.remove('visible');
            }
        });
    }

    // Optional: Automatically show video overlay on page load
    // Uncomment the line below if you want the video to show automatically
    // showVideo();

    // Event listener for expert button
    if (expertButton) {
        expertButton.addEventListener('click', () => {
            window.location.href = 'consult.html';
        });
    }
}
