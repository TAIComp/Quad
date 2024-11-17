// File: static/js/voiceInteraction.js

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
            formData.append('audio', audioBlob, 'recording.webm');

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
export function initVoiceInteraction() {
    if (document.querySelector('.consult-container')) {
        new VoiceInteraction();
    }
}
