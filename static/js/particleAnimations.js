// File: static/js/particleAnimations.js

/* =========================
   Particle (Blob) Animations
========================= */
export function initParticleAnimations() {
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
        return;
    }

    const ctx = canvas.getContext('2d');
    const nodes = [];
    const maxDistance = 150; // Connection distance between particles
    const numNodes = 50; // Number of particles
    
    // Define a virtual space that's larger than the viewport
    const virtualSpace = {
        width: window.innerWidth,  // Match to window width
        height: window.innerHeight  // Match to window height
    };

    function initCanvas() {
        // Set canvas size to match window size
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;

        // Clear existing nodes
        nodes.length = 0;

        // Create new nodes
        for (let i = 0; i < numNodes; i++) {
            nodes.push(new Node(
                Math.random() * virtualSpace.width,
                Math.random() * virtualSpace.height
            ));
        }
    }

    function drawNodes() {
        // Clear the canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        // Draw connections between particles
        ctx.beginPath();
        ctx.strokeStyle = 'rgba(0, 0, 0, 0.08)'; // Changed to black with low opacity
        ctx.lineWidth = 1;

        for (let i = 0; i < nodes.length; i++) {
            for (let j = i + 1; j < nodes.length; j++) {
                const dx = nodes[i].x - nodes[j].x;
                const dy = nodes[i].y - nodes[j].y;
                const distance = Math.sqrt(dx * dx + dy * dy);
                
                if (distance < maxDistance) {
                    const opacity = (1 - distance / maxDistance) * 0.15; // Reduced opacity for softer lines
                    ctx.strokeStyle = `rgba(0, 0, 0, ${opacity})`; // Changed to black
                    ctx.beginPath();
                    ctx.moveTo(nodes[i].x, nodes[i].y);
                    ctx.lineTo(nodes[j].x, nodes[j].y);
                    ctx.stroke();
                }
            }
        }

        // Draw particles
        nodes.forEach(node => {
            ctx.beginPath();
            ctx.fillStyle = 'rgba(0, 0, 0, 0.6)'; // Changed to black with medium opacity
            ctx.arc(node.x, node.y, node.radius, 0, Math.PI * 2);
            ctx.fill();
            
            // Update particle position
            node.update(virtualSpace.width, virtualSpace.height);
        });

        // Continue animation
        requestAnimationFrame(drawNodes);
    }

    // Initialize canvas and start animation
    initCanvas();
    drawNodes();

    // Handle window resize
    window.addEventListener('resize', () => {
        virtualSpace.width = window.innerWidth;
        virtualSpace.height = window.innerHeight;
        initCanvas();
    });
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', initParticleAnimations);
