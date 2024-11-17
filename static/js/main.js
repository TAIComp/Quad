// File: static/js/main.js

import { initFadeOutTransitions } from './fadeOutTransitions.js';
import { initVoiceInteraction } from './voiceInteraction.js';

document.addEventListener('DOMContentLoaded', () => {
    try {
        initFadeOutTransitions();
    } catch (error) {
        console.error('Error initializing fade transitions:', error);
    }

    try {
        initVoiceInteraction();
    } catch (error) {
        console.error('Error initializing voice interaction:', error);
    }
});
