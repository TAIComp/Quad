// File: static/js/consultRoom.js

export function initConsultRoom() {
    const missionContainer = document.getElementById('missionContainer');
    const consultContainer = document.getElementById('consultContainer');
    const missionVideo = document.getElementById('missionVideo');
    const skipButton = document.getElementById('skipButton');
    const status = document.getElementById('status');

    if (!missionContainer || !consultContainer || !missionVideo || !skipButton || !status) {
        console.log('Required elements not found. Skipping consultRoom initialization.');
        return;
    }

    // Add fade-transition class to containers
    missionContainer.classList.add('fade-transition', 'visible');
    consultContainer.classList.add('fade-transition');

    // Initially set status to wait for mission
    status.textContent = 'Please watch the introduction video...';

    // Function to show consult interface and initialize audio
    function showConsult() {
        // Fade out mission container
        missionContainer.classList.remove('visible');
        
        // Wait for fade out animation to complete
        setTimeout(() => {
            // Stop and unload the video
            missionVideo.pause();
            missionVideo.currentTime = 0;
            missionVideo.src = "";
            
            // Hide mission container
            missionContainer.style.display = 'none';
            
            // Show and fade in consult container
            consultContainer.style.display = 'block';
            // Force reflow
            consultContainer.offsetHeight;
            consultContainer.classList.add('visible');
            
            // Update status to waiting for input
            status.textContent = 'Waiting for your question...';

            // Initialize audio context only after user interaction
            if (typeof initializeAudio === 'function') {
                initializeAudio();
            }
        }, 500); // Match this with the CSS transition duration
    }

    // Handle video end
    missionVideo.addEventListener('ended', showConsult);

    // Handle skip button
    skipButton.addEventListener('click', function(e) {
        e.preventDefault();
        showConsult();
    });
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    initConsultRoom();
});
