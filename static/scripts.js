// =========================
// scripts.js
// =========================

document.addEventListener('DOMContentLoaded', () => {
    // Initialize all functionalities
    initNavigationBar();
    initLogin();
    initConsultRoom();
    initMissionVideo();
    initFadeOutTransitions();
    initParticleAnimations();
    initVoiceInteraction();
});

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

/* =========================
   Particle (Blob) Animations
========================= */
function initParticleAnimations() {
    console.log('Initializing particle animations...');
    
    class Node {
        constructor(x, y) {
            this.x = x;
            this.y = y;
            this.vx = (Math.random() - 0.5) * 0.5;
            this.vy = (Math.random() - 0.5) * 0.5;
            this.radius = 3;
        }

        update(width, height) {
            this.x += this.vx;
            this.y += this.vy;

            // Wrap around the edges instead of bouncing
            if (this.x < 0) this.x = width;
            if (this.x > width) this.x = 0;
            if (this.y < 0) this.y = height;
            if (this.y > height) this.y = 0;
        }
    }

    const canvas = document.getElementById('nodeCanvas');
    if (!canvas) {
        console.error('Canvas element with ID "nodeCanvas" not found.');
        return; // Exit if canvas is not present
    }
    const ctx = canvas.getContext('2d');
    const nodes = [];
    const maxDistance = 200; // Increased maxDistance for isInView
    const numNodes = 40; // Reduced numNodes

    // Define a virtual space that's larger than the viewport
    const virtualSpace = {
        width: 2000,  // Fixed virtual width
        height: 1500  // Fixed virtual height
    };

    function initCanvas() {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;

        console.log('Canvas initialized with width:', canvas.width, 'and height:', canvas.height);

        // Only create nodes if they don't exist yet
        if (nodes.length === 0) {
            for (let i = 0; i < numNodes; i++) {
                nodes.push(new Node(
                    Math.random() * virtualSpace.width,
                    Math.random() * virtualSpace.height
                ));
            }
            console.log('Created', numNodes, 'particles.');
        }
    }

    function drawNodes() {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        // Calculate the visible area offset (centering the virtual space)
        const offsetX = Math.max(0, (virtualSpace.width - canvas.width) / 2);
        const offsetY = Math.max(0, (virtualSpace.height - canvas.height) / 2);
        
        // Draw connections between particles
        ctx.beginPath();
        for (let i = 0; i < nodes.length; i++) {
            for (let j = i + 1; j < nodes.length; j++) {
                const dx = nodes[i].x - nodes[j].x;
                const dy = nodes[i].y - nodes[j].y;
                const distance = Math.sqrt(dx * dx + dy * dy);
                
                if (distance < maxDistance) {
                    const opacity = (1 - distance / maxDistance) * 0.5;
                    ctx.strokeStyle = `rgba(100, 100, 100, ${opacity})`;
                    
                    // Calculate positions relative to the visible area
                    const x1 = nodes[i].x - offsetX;
                    const y1 = nodes[i].y - offsetY;
                    const x2 = nodes[j].x - offsetX;
                    const y2 = nodes[j].y - offsetY;
                    
                    // Draw the line only if at least one end is in view
                    if (isInView(x1, y1) || isInView(x2, y2)) {
                        ctx.moveTo(x1, y1);
                        ctx.lineTo(x2, y2);
                    }
                }
            }
        }
        ctx.stroke();

        // Draw each particle
        nodes.forEach(node => {
            const x = node.x - offsetX;
            const y = node.y - offsetY;
            
            if (isInView(x, y)) {
                ctx.beginPath();
                ctx.fillStyle = 'rgba(100, 100, 100, 0.8)';
                ctx.arc(x, y, node.radius, 0, Math.PI * 2);
                ctx.fill();
            }
            
            // Update particle position for the next frame
            node.update(virtualSpace.width, virtualSpace.height);
        });

        requestAnimationFrame(drawNodes);
    }

    function isInView(x, y) {
        return x >= -maxDistance && 
               x <= canvas.width + maxDistance && 
               y >= -maxDistance && 
               y <= canvas.height + maxDistance;
    }

    // Initialize and start animation
    initCanvas();
    drawNodes();

    // Handle window resize to maintain responsiveness
    window.addEventListener('resize', () => {
        console.log('Window resized.');
        initCanvas();
    });
}

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

/* =========================
   Consult Room Functionality
========================= */
function initConsultRoom() {
    // Ensure this script runs only on consult_room.html
    if (!window.location.pathname.endsWith('consult.html')) return;

    const conversationDiv = document.getElementById('conversation');
    const statusDiv = document.getElementById('status');
    const endButton = document.getElementById('end-button');

    let mediaRecorder;
    let audioChunks = [];
    let isRecording = false;
    let isPlaying = false;
    let username = null; // Will be set after authentication

    // Function to append messages to the conversation div
    function appendMessage(message, sender) {
        const messageDiv = document.createElement('div');
        messageDiv.classList.add('message', sender);
        messageDiv.textContent = message;
        conversationDiv.appendChild(messageDiv);
        conversationDiv.scrollTop = conversationDiv.scrollHeight;
    }

    // Function to initialize microphone access and start the loop
    async function init() {
        try {
            // Fetch authenticated user information
            const response = await fetch('/api/me', {
                method: 'GET',
                credentials: 'include', // include cookies
            });

            if (response.ok) {
                const data = await response.json();
                username = data.username;
            } else {
                throw new Error('User not authenticated.');
            }

            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            mediaRecorder = new MediaRecorder(stream, { mimeType: 'audio/webm; codecs=opus' });

            mediaRecorder.ondataavailable = event => {
                audioChunks.push(event.data);
                if (mediaRecorder.state === 'inactive') {
                    const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
                    audioChunks = [];
                    sendAudio(audioBlob);
                }
            };

            mediaRecorder.onstart = () => {
                isRecording = true;
                updateStatus('Listening...');
            };

            mediaRecorder.onstop = () => {
                isRecording = false;
                updateStatus('Processing...');
            };

            mediaRecorder.onerror = event => {
                console.error('MediaRecorder error:', event.error);
                updateStatus('Error occurred during recording.');
            };

            // Start the first recording
            startRecording();
        } catch (error) {
            console.error('Error accessing microphone:', error);
            updateStatus('Microphone access denied or user not authenticated.');
        }
    }

    // Function to update status
    function updateStatus(message) {
        if (statusDiv) {
            statusDiv.textContent = message;
        }
    }

    // Function to start recording
    function startRecording() {
        if (mediaRecorder && mediaRecorder.state !== 'recording') {
            mediaRecorder.start();
        }
    }

    // Function to stop recording
    function stopRecording() {
        if (mediaRecorder && mediaRecorder.state === 'recording') {
            mediaRecorder.stop();
        }
    }

    // Function to send audio to the backend
    async function sendAudio(blob) {
        const formData = new FormData();
        formData.append('audio', blob, 'recording.webm');

        try {
            const response = await fetch(`/api/get-response`, {
                method: 'POST',
                body: formData,
                credentials: 'include' // Include cookies for authentication
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to get response from server.');
            }

            const data = await response.json();
            appendMessage(data.userInput, 'user');
            appendMessage(data.aiResponse, 'ai');

            if (data.audioBase64) {
                playAudio(data.audioBase64);
            } else {
                // If no audio, resume listening after a short delay
                setTimeout(() => {
                    if (!isPlaying) { // Prevent overlapping recordings
                        startRecording();
                    }
                }, 1000);
            }
        } catch (error) {
            console.error('Error:', error);
            updateStatus(`Error: ${error.message}`);
            // Optionally, you can decide to retry or end the conversation
        }
    }

    // Function to play AI response audio
    function playAudio(base64Audio) {
        isPlaying = true;
        updateStatus('Playing AI response...');

        const audio = new Audio(`data:audio/mp3;base64,${base64Audio}`);
        audio.play().then(() => {
            isPlaying = false;
            updateStatus('Listening...');
            // Resume listening after playback
            startRecording();
        }).catch(err => {
            console.error('Audio playback error:', err);
            updateStatus('Error playing audio.');
            isPlaying = false;
            // Optionally, resume listening even if playback fails
            startRecording();
        });
    }

    // Event listener to end the conversation
    if (endButton) {
        endButton.addEventListener('click', () => {
            stopRecording();
            updateStatus('Conversation ended.');
            appendMessage('You have ended the conversation.', 'user');
        });
    }

    // Initialize the consultation on page load
    init();
}

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

/* =========================
   Voice Interaction Functionality
========================= */
class VoiceInteraction {
    constructor() {
        // State management
        this.states = {
            WAITING: 'waiting',
            LISTENING: 'listening',
            PROCESSING: 'processing',
            ANSWERING: 'answering'
        };
        this.currentState = this.states.WAITING;

        // DOM elements
        this.recordingIndicator = document.getElementById('recordingIndicator');
        this.conversation = document.getElementById('conversation');
        this.status = document.getElementById('status');
        this.responseAudio = document.getElementById('responseAudio');
        
        // Audio processing setup
        this.mediaRecorder = null;
        this.audioChunks = [];
        this.audioContext = null;
        this.meydaAnalyzer = null;
        this.stream = null;
        
        // Silence detection
        this.silenceStart = null;
        this.SILENCE_DURATION = 1000; // 1 second silence threshold
        this.consecutiveSilenceFrames = 0;
        this.REQUIRED_SILENCE_FRAMES = 50; // Adjust based on your frame rate
        
        // Voice activity detection
        this.voiceDetected = false;
        this.VAD_THRESHOLD = {
            RMS: 0.01,
            ZCR: 50,
            FLATNESS: 0.5
        };

        this.init();
        this.setupAudioResponseHandling();
    }

    async init() {
        try {
            this.stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            await this.setupAudioProcessing(this.stream);
            this.setState(this.states.WAITING);
        } catch (error) {
            console.error('Error initializing voice interaction:', error);
            this.updateStatus('Error accessing microphone. Please check permissions.');
        }
    }

    async setupAudioProcessing(stream) {
        this.audioContext = new AudioContext();
        const source = this.audioContext.createMediaStreamSource(stream);

        // Initialize Meyda with bound callback
        this.meydaAnalyzer = Meyda.createMeydaAnalyzer({
            audioContext: this.audioContext,
            source: source,
            bufferSize: 512,
            featureExtractors: ['rms', 'zcr', 'spectralFlatness'],
            callback: this.processAudioFeatures.bind(this)
        });

        // Setup MediaRecorder
        this.mediaRecorder = new MediaRecorder(stream);
        this.mediaRecorder.ondataavailable = (event) => {
            if (event.data.size > 0) {
                this.audioChunks.push(event.data);
            }
        };
        this.mediaRecorder.onstop = () => this.processRecording();

        // Start continuous analysis
        this.meydaAnalyzer.start();
    }

    setupAudioResponseHandling() {
        if (this.responseAudio) {
            this.responseAudio.addEventListener('ended', () => {
                if (this.currentState === this.states.ANSWERING) {
                    this.setState(this.states.WAITING);
                }
            });

            this.responseAudio.addEventListener('error', () => {
                console.error('Audio playback error');
                this.setState(this.states.WAITING);
            });
        }
    }

    startRecording() {
        if (this.mediaRecorder && this.mediaRecorder.state === 'inactive') {
            this.audioChunks = [];
            this.mediaRecorder.start();
            this.setState(this.states.LISTENING);
            console.log('Started recording');
        }
    }

    stopRecording() {
        if (this.mediaRecorder && this.mediaRecorder.state === 'recording') {
            this.mediaRecorder.stop();
            this.setState(this.states.PROCESSING);
            console.log('Stopped recording');
        }
    }

    setState(newState) {
        const prevState = this.currentState;
        this.currentState = newState;
        
        // Reset state-specific variables
        if (newState === this.states.WAITING) {
            this.consecutiveSilenceFrames = 0;
            this.silenceStart = null;
            this.voiceDetected = false;
        }

        this.updateUIForState(prevState, newState);
    }

    updateUIForState(prevState, newState) {
        if (!this.recordingIndicator) return;

        const statusText = this.recordingIndicator.querySelector('.status-text');
        const pulseRing = this.recordingIndicator.querySelector('.pulse-ring');

        // Remove previous state class
        if (prevState) {
            this.recordingIndicator.classList.remove(prevState);
        }

        // Add new state class
        this.recordingIndicator.classList.add(newState);

        const stateConfig = {
            [this.states.WAITING]: {
                status: 'Waiting for speech...',
                text: 'Waiting...',
                class: 'waiting'
            },
            [this.states.LISTENING]: {
                status: 'Listening to your speech...',
                text: 'Listening...',
                class: 'listening'
            },
            [this.states.PROCESSING]: {
                status: 'Processing your input...',
                text: 'Processing...',
                class: 'processing'
            },
            [this.states.ANSWERING]: {
                status: 'AI is responding...',
                text: 'Answering...',
                class: 'answering'
            }
        };

        const config = stateConfig[newState];
        this.updateStatus(config.status);
        statusText.textContent = config.text;
        this.recordingIndicator.className = `recording-indicator ${config.class}`;
    }

    updateStatus(message) {
        if (this.status) {
            this.status.textContent = message;
        }
    }

    processAudioFeatures(features) {
        if (this.currentState !== this.states.WAITING && 
            this.currentState !== this.states.LISTENING) {
            return;
        }

        const isVoice = this.detectVoiceActivity(features);
        
        if (isVoice) {
            this.handleVoiceDetected();
        } else {
            this.handleSilenceDetected();
        }
    }

    detectVoiceActivity(features) {
        const isVoice = features.rms > this.VAD_THRESHOLD.RMS &&
                       features.zcr > this.VAD_THRESHOLD.ZCR &&
                       features.spectralFlatness < this.VAD_THRESHOLD.FLATNESS;

        // Add hysteresis to prevent rapid switching
        if (isVoice) {
            this.consecutiveSilenceFrames = 0;
            return true;
        } else {
            this.consecutiveSilenceFrames++;
            return false;
        }
    }

    handleVoiceDetected() {
        this.voiceDetected = true;
        this.silenceStart = null;
        
        if (this.currentState === this.states.WAITING) {
            this.startRecording();
        }
    }

    handleSilenceDetected() {
        if (!this.voiceDetected || this.currentState !== this.states.LISTENING) {
            return;
        }

        if (this.consecutiveSilenceFrames >= this.REQUIRED_SILENCE_FRAMES) {
            if (!this.silenceStart) {
                this.silenceStart = Date.now();
            } else if (Date.now() - this.silenceStart >= this.SILENCE_DURATION) {
                this.stopRecording();
            }
        }
    }

    async processRecording() {
        if (this.audioChunks.length === 0) {
            this.setState(this.states.WAITING);
            return;
        }

        try {
            this.setState(this.states.PROCESSING);
            const audioBlob = new Blob(this.audioChunks, { type: 'audio/webm' });
            const formData = new FormData();
            formData.append('audio', audioBlob);

            // First, get the transcription and AI response
            const response = await fetch('/api/get-response', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error('Failed to get AI response');
            }

            const data = await response.json();
            
            // Update UI with text response
            this.updateConversation(data.userInput, data.aiResponse);
            
            // Get audio response for the AI text
            const audioResponse = await this.getAudioResponse(data.aiResponse);
            
            // Set state to answering and play the audio
            this.setState(this.states.ANSWERING);
            await this.playAudioResponse(audioResponse);

        } catch (error) {
            console.error('Error processing recording:', error);
            this.updateStatus('Error processing audio');
            this.setState(this.states.WAITING);
        }
    }

    async getAudioResponse(text) {
        try {
            const response = await fetch('/api/text-to-speech', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ text })
            });

            if (!response.ok) {
                throw new Error('Failed to get audio response');
            }

            const audioData = await response.blob();
            return audioData;

        } catch (error) {
            console.error('Error getting audio response:', error);
            throw error;
        }
    }

    async playAudioResponse(audioBlob) {
        try {
            const audioUrl = URL.createObjectURL(audioBlob);
            
            // Create a promise that resolves when the audio finishes playing
            const playPromise = new Promise((resolve, reject) => {
                this.responseAudio.src = audioUrl;
                
                this.responseAudio.onended = () => {
                    URL.revokeObjectURL(audioUrl); // Clean up the URL
                    resolve();
                };
                
                this.responseAudio.onerror = (error) => {
                    URL.revokeObjectURL(audioUrl);
                    reject(error);
                };
            });

            // Start playing the audio
            await this.responseAudio.play();
            
            // Wait for the audio to finish playing
            await playPromise;
            
            // After audio finishes, transition back to waiting state
            this.setState(this.states.WAITING);

        } catch (error) {
            console.error('Error playing audio response:', error);
            this.setState(this.states.WAITING);
        }
    }

    updateConversation(userInput, aiResponse) {
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message-container';
        
        // Add timestamp
        const timestamp = new Date().toLocaleTimeString();
        
        messageDiv.innerHTML = `
            <div class="message-timestamp">${timestamp}</div>
            <div class="user-message">
                <div class="message-icon">👤</div>
                <div class="message-content">${userInput}</div>
            </div>
            <div class="ai-message">
                <div class="message-icon">🤖</div>
                <div class="message-content">${aiResponse}</div>
            </div>
        `;
        
        this.conversation.appendChild(messageDiv);
        this.conversation.scrollTop = this.conversation.scrollHeight;
    }

    cleanup() {
        if (this.meydaAnalyzer) {
            this.meydaAnalyzer.stop();
        }
        if (this.audioContext) {
            this.audioContext.close();
        }
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
        }
    }
}

// Initialize voice interaction
function initVoiceInteraction() {
    if (document.querySelector('.consult-container')) {
        new VoiceInteraction();
    }
}
