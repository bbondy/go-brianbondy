// Easter Egg: Stick Figure
(function() {
  let logoClickCount = 0;
  let stickFigure = null;
  let stickX = 100;
  let stickY = 100;
  let isWalking = false;
  let activeKeys = new Set();
  let lastDirectionX = 1; // default right
  let lastDirectionY = 0;
  // Enhanced mobile detection
  let isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || 
                (navigator.maxTouchPoints && navigator.maxTouchPoints > 2) ||
                ('ontouchstart' in window);
  
  // (debug logs removed)
  let mobileControls = null;
  let easterEggActivated = false; // Flag to prevent multiple activations
  
  // Game scoring system
  let gameScore = 0;
  let highScore = parseInt(localStorage.getItem('stickFigureHighScore') || '0');
  let scoreDisplay = null;
  
  // Danger mechanics
  let playerHealth = 3;
  let maxHealth = 3;
  let healthDisplay = null;
  let enemyBullets = [];
  let bouncingBalls = [];
  let ballSpawnTimer = null;
  let gameActive = true;
  let gameStartTime = 0;
  // Add variable declarations near other globals after heart logic
  let heartPickups = [];

  const themeToggleBtn = document.getElementById('theme-toggle');
  const logo = document.getElementById('site-logo'); // Keep both for backwards compatibility
  if (!themeToggleBtn && !logo) return;

  // Create stick figure SVG with animated legs
  function createStickFigure() {
    const div = document.createElement('div');
    div.className = 'stick-figure';
    div.innerHTML = `
      <svg viewBox="0 0 30 40" xmlns="http://www.w3.org/2000/svg">
        <!-- Head -->
        <circle cx="15" cy="6" r="4" fill="none" stroke="var(--text-color)" stroke-width="2"/>
        <!-- Body -->
        <line x1="15" y1="10" x2="15" y2="25" stroke="var(--text-color)" stroke-width="2"/>
        <!-- Arms -->
        <line x1="15" y1="16" x2="8" y2="12" stroke="var(--text-color)" stroke-width="2" class="arm-left"/>
        <line x1="15" y1="16" x2="22" y2="12" stroke="var(--text-color)" stroke-width="2" class="arm-right"/>
        <!-- Animated Legs -->
        <line x1="15" y1="25" x2="8" y2="36" stroke="var(--text-color)" stroke-width="2" class="leg-left"/>
        <line x1="15" y1="25" x2="22" y2="36" stroke="var(--text-color)" stroke-width="2" class="leg-right"/>
        <!-- Eyes -->
        <circle cx="13" cy="5" r="0.5" fill="var(--text-color)"/>
        <circle cx="17" cy="5" r="0.5" fill="var(--text-color)"/>
        <!-- Smile -->
        <path d="M 12 7 Q 15 9 18 7" fill="none" stroke="var(--text-color)" stroke-width="1"/>
      </svg>
    `;
    return div;
  }

  // Create trail effect
  function createTrail(x, y) {
    const trail = document.createElement('div');
    trail.className = 'trail-dot';
    trail.style.left = (x + 15) + 'px';
    trail.style.top = (y + 20) + 'px';
    document.body.appendChild(trail);
    setTimeout(() => trail.remove(), 1000);
  }

  // Create mobile controls
  function createMobileControls() {
    if (!isMobile || mobileControls) return;
    
    
    const controls = document.createElement('div');
    controls.className = 'mobile-controls';
    controls.innerHTML = `
      <div class="mobile-dpad">
        <div class="mobile-btn up" data-key="k">↑</div>
        <div class="mobile-btn left" data-key="h">←</div>
        <div class="mobile-btn right" data-key="l">→</div>
        <div class="mobile-action-btn shoot" data-key="f">🎯</div>
        <div class="mobile-btn down" data-key="j">↓</div>
      </div>
    `;
    
    // Prevent scrolling and zooming on the controls
    controls.style.touchAction = 'none';
    controls.style.userSelect = 'none';
    controls.style.webkitUserSelect = 'none';
    
    // Add touch event listeners
    controls.querySelectorAll('.mobile-btn, .mobile-action-btn').forEach(btn => {
      const key = btn.dataset.key;
      let moveInterval = null;
      
      // Prevent default behaviors
      btn.style.touchAction = 'none';
      btn.style.userSelect = 'none';
      btn.style.webkitUserSelect = 'none';
      
      btn.addEventListener('touchstart', (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!stickFigure) return;
        
        // For movement keys, start continuous movement
        if (['h', 'j', 'k', 'l'].includes(key)) {
          activeKeys.add(key);
          updateAnimation();
          
          // Dismiss instructions on first mobile movement
          if (window.currentHint) {
            window.currentHint.remove();
            window.currentHint = null;
          }
          
          // Start continuous movement
          const moveStep = () => {
            if (!stickFigure || !activeKeys.has(key)) return;
            
            const speed = 8;
            let deltaX = 0, deltaY = 0;
            
            if (key === 'h') deltaX = -speed; // left
            if (key === 'l') deltaX = speed;  // right
            if (key === 'k') deltaY = -speed; // up
            if (key === 'j') deltaY = speed;  // down
            
            // Update last direction when moving
            if (deltaX !== 0) {
              lastDirectionX = deltaX > 0 ? 1 : -1;
              lastDirectionY = 0;
            } else if (deltaY !== 0) {
              lastDirectionY = deltaY > 0 ? 1 : -1;
              lastDirectionX = 0;
            }
            
            moveStickFigure(deltaX, deltaY);
          };
          
          // Initial move
          moveStep();
          // Continue moving every 50ms while held
          moveInterval = setInterval(moveStep, 50);
        } else {
          // For action keys (shoot/jump), trigger immediately
          const keyEvent = new KeyboardEvent('keydown', {
            key: key === ' ' ? ' ' : key.toLowerCase(),
            code: key === ' ' ? 'Space' : `Key${key.toUpperCase()}`,
            bubbles: true
          });
          document.dispatchEvent(keyEvent);
        }
      });
      
      btn.addEventListener('touchend', (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!stickFigure) return;
        
        // Stop continuous movement
        if (['h', 'j', 'k', 'l'].includes(key)) {
          activeKeys.delete(key);
          updateAnimation();
          if (moveInterval) {
            clearInterval(moveInterval);
            moveInterval = null;
          }
        }
      });
      
      // Also stop on touch cancel (when finger leaves button area)
      btn.addEventListener('touchcancel', (e) => {
        
        if (['h', 'j', 'k', 'l'].includes(key)) {
          activeKeys.delete(key);
          updateAnimation();
          if (moveInterval) {
            clearInterval(moveInterval);
            moveInterval = null;
          }
        }
      });
    });
    
    document.body.appendChild(controls);
    mobileControls = controls;
    
    
  }

  // Show/hide mobile controls
  function toggleMobileControls(show) {
    if (!mobileControls) return;
    mobileControls.classList.toggle('active', show);
  }
  
  // Find a safe spawn position away from elements
  function findSafeSpawnPosition() {
    const maxAttempts = 100;
    const stickWidth = 30;
    const stickHeight = 40;
    const minPadding = 50; // Minimum distance from elements
    
    // Define safe boundaries (avoid UI elements in corners)
    const bounds = {
      left: 20,
      top: 120, // Avoid score/health displays
      right: window.innerWidth - stickWidth - 200, // Avoid right-side UI
      bottom: window.innerHeight - stickHeight - 100 // Avoid bottom UI
    };
    
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      // Try random position within bounds
      const testX = bounds.left + Math.random() * (bounds.right - bounds.left);
      const testY = bounds.top + Math.random() * (bounds.bottom - bounds.top);
      
      // Check if this position is clear
      if (isPositionSafe(testX, testY, stickWidth, stickHeight, minPadding)) {
        return { x: testX, y: testY };
      }
    }
    
    // Fallback to center-left if no safe position found
    return { 
      x: Math.min(150, window.innerWidth * 0.2), 
      y: Math.min(200, window.innerHeight * 0.4) 
    };
  }
  
  // Check if a position is safe (no collision with elements)
  function isPositionSafe(x, y, width, height, padding) {
    // Check the actual stick figure bounds
    const stickRect = {
      left: x,
      top: y,
      right: x + width,
      bottom: y + height
    };
    
    // Get all elements that are shootable (same as bullet collision targets)
    const elements = document.querySelectorAll('h1, h2, h3, .run-card, .project-entry, .advice-card, .total-item, button, a:not([href="/"]), img, span.icon, .family-photo, select, input, textarea, .contribution-day');
    
    for (let element of elements) {
      const rect = element.getBoundingClientRect();
      
      // Skip very small elements (like icons) or hidden elements
      if (rect.width < 15 || rect.height < 15) continue;
      
      // Skip elements that are not visible text content
      if (element.tagName === 'P' && (!element.textContent || element.textContent.trim().length < 5)) continue;
      
      // Expand element bounds by padding for clearance
      const elementRect = {
        left: rect.left - padding,
        top: rect.top - padding,
        right: rect.right + padding,
        bottom: rect.bottom + padding
      };
      
      // Check if stick figure would overlap with padded element area
      if (stickRect.left < elementRect.right && 
          stickRect.right > elementRect.left && 
          stickRect.top < elementRect.bottom && 
          stickRect.bottom > elementRect.top) {
        return false; // Position not safe
      }
    }
    
    return true; // Position is safe
  }
  
  // Create score display
  function createScoreDisplay() {
    scoreDisplay = document.createElement('div');
    scoreDisplay.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: rgba(0, 0, 0, 0.8);
      color: #fff;
      padding: 10px 15px;
      border-radius: 8px;
      font-family: monospace;
      font-size: 14px;
      z-index: 10001;
      border: 2px solid var(--link-color);
      box-shadow: 0 4px 12px rgba(0,0,0,0.4);
    `;
    updateScoreDisplay();
    document.body.appendChild(scoreDisplay);
  }
  
  // Update score display
  function updateScoreDisplay() {
    if (!scoreDisplay) return;
    const isNewRecord = gameScore > highScore;
    scoreDisplay.innerHTML = `
      <div><strong>🎯 Score:</strong> ${gameScore} ${isNewRecord ? '🔥' : ''}</div>
      <div style="font-size: 12px; opacity: 0.8;"><strong>👑 Best:</strong> ${Math.max(gameScore, highScore)}</div>
    `;
  }
  
  // Add score for destroying elements
  function addScore(points) {
    if (!gameActive) return;
    
    gameScore += points;
    
    // Update high score if needed
    if (gameScore > highScore) {
      highScore = gameScore;
      localStorage.setItem('stickFigureHighScore', highScore.toString());
    }
    
    updateScoreDisplay();
    
    // Show floating score text
    showFloatingScore(points);
  }
  
  // Show floating score animation
  function showFloatingScore(points) {
    const floatingScore = document.createElement('div');
    floatingScore.style.cssText = `
      position: fixed;
      left: ${stickX + 35}px;
      top: ${stickY}px;
      color: #ffff00;
      font-family: monospace;
      font-weight: bold;
      font-size: 16px;
      z-index: 10000;
      pointer-events: none;
      text-shadow: 2px 2px 4px rgba(0,0,0,0.8);
      animation: floating-score 1.5s ease-out forwards;
    `;
    floatingScore.textContent = `+${points}`;
    document.body.appendChild(floatingScore);
    setTimeout(() => floatingScore.remove(), 1500);
  }
  
  // Create health display
  function createHealthDisplay() {
    healthDisplay = document.createElement('div');
    healthDisplay.style.cssText = `
      position: fixed;
      top: 80px;
      right: 20px;
      background: rgba(0, 0, 0, 0.8);
      color: #fff;
      padding: 10px 15px;
      border-radius: 8px;
      font-family: monospace;
      font-size: 14px;
      z-index: 10001;
      border: 2px solid #ff4444;
      box-shadow: 0 4px 12px rgba(0,0,0,0.4);
    `;
    updateHealthDisplay();
    document.body.appendChild(healthDisplay);
  }
  
  // Update health display
  function updateHealthDisplay() {
    if (!healthDisplay) return;
    const hearts = '❤️'.repeat(playerHealth) + '🖤'.repeat(maxHealth - playerHealth);
    healthDisplay.innerHTML = `<strong>💖 Health:</strong> ${hearts}`;
  }
  
  // Take damage
  function takeDamage(amount = 1) {
    if (!gameActive) return;
    
    playerHealth -= amount;
    updateHealthDisplay();
    
    // Flash the stick figure red
    if (stickFigure) {
      stickFigure.classList.add('damaged');
      setTimeout(() => stickFigure.classList.remove('damaged'), 500);
    }
    
    if (playerHealth <= 0) {
      gameOver();
    }
  }
  
  // Game over
  function gameOver() {
    gameActive = false;
    
    if (stickFigure) {
      stickFigure.classList.add('dying');
    }
    
    // Stop all timers
    if (ballSpawnTimer) clearInterval(ballSpawnTimer);
    
    // Show game over message
    const gameOverDiv = document.createElement('div');
    gameOverDiv.id = 'game-over-screen';
    gameOverDiv.style.cssText = `
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: rgba(0, 0, 0, 0.9);
      color: #fff;
      padding: 30px;
      border-radius: 12px;
      text-align: center;
      font-family: monospace;
      font-size: 18px;
      z-index: 10002;
      border: 3px solid #ff4444;
      box-shadow: 0 8px 32px rgba(0,0,0,0.8);
      cursor: pointer;
    `;
    
    const survivalTime = Math.floor((Date.now() - gameStartTime) / 1000);
    gameOverDiv.innerHTML = `
      <h2 style="margin: 0 0 15px 0; color: #ff4444;">💀 GAME OVER 💀</h2>
      <p><strong>Final Score:</strong> ${gameScore}</p>
      <p><strong>Survived:</strong> ${survivalTime}s</p>
      <p><strong>High Score:</strong> ${Math.max(gameScore, highScore)}</p>
      <p style="font-size: 16px; margin-top: 20px; color: #ffff00;">✨ Tap or Click to Restart ✨</p>
    `;
    
    // Add click to restart functionality
    gameOverDiv.addEventListener('click', restartGame);
    document.addEventListener('click', restartGame);
    document.addEventListener('touchstart', restartGame);
    
    document.body.appendChild(gameOverDiv);
  }
  
  // Restart game function
  function restartGame() {
    // Remove event listeners
    document.removeEventListener('click', restartGame);
    document.removeEventListener('touchstart', restartGame);
    
    // Remove game over screen
    const gameOverScreen = document.getElementById('game-over-screen');
    if (gameOverScreen) gameOverScreen.remove();
    
    // Reset all game state
    if (stickFigure) {
      stickFigure.remove();
      stickFigure = null;
    }
    
    // Remove displays
    if (scoreDisplay) {
      scoreDisplay.remove();
      scoreDisplay = null;
    }
    if (healthDisplay) {
      healthDisplay.remove();
      healthDisplay = null;
    }
    
    // Remove all bouncing balls
    bouncingBalls.forEach(ball => {
      if (ball.element.parentNode) ball.element.remove();
    });
    bouncingBalls = [];
    
    // Clear timers
    if (ballSpawnTimer) clearInterval(ballSpawnTimer);
    
    // Reset game variables
    gameScore = 0;
    playerHealth = 3;
    gameActive = true;
    gameStartTime = Date.now();
    
    // Start new game
    stickFigure = createStickFigure();
    
    // Find new safe spawn position
    const spawnPos = findSafeSpawnPosition();
    stickX = spawnPos.x;
    stickY = spawnPos.y;
    
    stickFigure.style.left = stickX + 'px';
    stickFigure.style.top = stickY + 'px';
    document.body.appendChild(stickFigure);
    
    createScoreDisplay();
    createHealthDisplay();
    startEnemyAttacks();
    startBouncingBalls();
    // In restartGame cleanup, remove hearts and clear array
    heartPickups.forEach(h => h.remove());
    heartPickups = [];
  }
  

  
  // Start enemy attacks
  function startEnemyAttacks() {
    setInterval(() => {
      if (!gameActive || !stickFigure) return;
      
      // Random chance for elements to shoot back
      if (Math.random() < 0.15) {
        createEnemyBullet();
      }
    }, 2000);
  }
  
  // Create enemy bullet from random element
  function createEnemyBullet() {
    const elements = document.querySelectorAll('h1, h2, h3, .run-card, .project-entry, button');
    if (elements.length === 0) return;
    
    const randomElement = elements[Math.floor(Math.random() * elements.length)];
    const rect = randomElement.getBoundingClientRect();
    
    const bullet = document.createElement('div');
    bullet.className = 'enemy-bullet';
    
    // Use viewport coordinates (same as stick figure)
    let bulletX = rect.left + rect.width / 2;
    let bulletY = rect.top + rect.height / 2;
    
    bullet.style.left = bulletX + 'px';
    bullet.style.top = bulletY + 'px';
    document.body.appendChild(bullet);
    
    // Calculate direction towards stick figure
    const deltaX = stickX + 15 - bulletX;
    const deltaY = stickY + 20 - bulletY;
    const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
    const directionX = deltaX / distance;
    const directionY = deltaY / distance;
    
    const bulletSpeed = 3;
    
    const moveEnemyBullet = () => {
      if (!gameActive) {
        bullet.remove();
        return;
      }
      
      bulletX += directionX * bulletSpeed;
      bulletY += directionY * bulletSpeed;
      
      bullet.style.left = bulletX + 'px';
      bullet.style.top = bulletY + 'px';
      
      // Check if bullet hit stick figure
      if (stickFigure) {
        const stickRect = {
          left: stickX,
          top: stickY,
          right: stickX + 30,
          bottom: stickY + 40
        };
        
        if (bulletX >= stickRect.left && bulletX <= stickRect.right &&
            bulletY >= stickRect.top && bulletY <= stickRect.bottom) {
          createExplosion(bulletX, bulletY);
          takeDamage(1);
          bullet.remove();
          return;
        }
      }
      
      // Remove if off screen
      if (bulletX < -10 || bulletX > window.innerWidth + 10 || 
          bulletY < -10 || bulletY > window.innerHeight + 10) {
        bullet.remove();
        return;
      }
      
                requestAnimationFrame(moveEnemyBullet);
  };
  
  requestAnimationFrame(moveEnemyBullet);
}

// Start bouncing balls system
function startBouncingBalls() {
  // Start with one ball immediately
  createBouncingBall();
  
  // Escalating ball spawning - more balls over time
  ballSpawnTimer = setInterval(() => {
    if (!gameActive || !stickFigure) return;
    
    // Calculate how long the game has been running (in seconds)
    const gameTime = (Date.now() - gameStartTime) / 1000;
    
    // Spawn more balls as time goes on
    // First 30 seconds: 1 ball every 4 seconds
    // 30-60 seconds: 1 ball every 3 seconds
    // 60+ seconds: 1 ball every 2 seconds
    let spawnRate = 4000; // milliseconds
    if (gameTime > 60) spawnRate = 2000;
    else if (gameTime > 30) spawnRate = 3000;
    
    // Clear old timer and set new one
    clearInterval(ballSpawnTimer);
    ballSpawnTimer = setInterval(() => {
      if (!gameActive || !stickFigure) return;
      createBouncingBall();
      
      // Sometimes spawn 2 balls at once for extra chaos
      if (gameTime > 45 && Math.random() < 0.3) {
        setTimeout(() => createBouncingBall(), 500);
      }
    }, spawnRate);
    
  }, 10000); // Check every 10 seconds for rate changes
  
  // Initial spawn rate
  ballSpawnTimer = setInterval(() => {
    if (!gameActive || !stickFigure) return;
    createBouncingBall();
  }, 4000);
}

// Create a bouncing ball
function createBouncingBall() {
  const ball = {
    element: document.createElement('div'),
    x: Math.random() * (window.innerWidth - 40) + 20,
    y: Math.random() * (window.innerHeight - 40) + 20,
    vx: (Math.random() - 0.5) * 8, // Random velocity X
    vy: (Math.random() - 0.5) * 8, // Random velocity Y
    radius: 15,
    id: Math.random().toString(36)
  };
  
  ball.element.className = 'bouncing-ball';
  ball.element.style.left = ball.x + 'px';
  ball.element.style.top = ball.y + 'px';
  document.body.appendChild(ball.element);
  
  bouncingBalls.push(ball);
  animateBouncingBall(ball);
}

// Animate bouncing ball
function animateBouncingBall(ball) {
  const animate = () => {
    if (!gameActive || !ball.element.parentNode) {
      // Clean up if game over or ball removed
      const index = bouncingBalls.indexOf(ball);
      if (index > -1) bouncingBalls.splice(index, 1);
      return;
    }
    
    // Update position
    ball.x += ball.vx;
    ball.y += ball.vy;
    
    // Bounce off screen edges
    if (ball.x <= ball.radius || ball.x >= window.innerWidth - ball.radius) {
      ball.vx = -ball.vx;
      ball.x = Math.max(ball.radius, Math.min(window.innerWidth - ball.radius, ball.x));
    }
    if (ball.y <= ball.radius || ball.y >= window.innerHeight - ball.radius) {
      ball.vy = -ball.vy;
      ball.y = Math.max(ball.radius, Math.min(window.innerHeight - ball.radius, ball.y));
    }
    
    // Check collision with page elements
    const collidedElement = checkBallElementCollision(ball);
    if (collidedElement) {
      // Bounce off element - simple reflection
      const rect = collidedElement.getBoundingClientRect();
      // Use viewport coordinates (same as ball)
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;
      
      // Determine which side of element we hit
      const deltaX = ball.x - centerX;
      const deltaY = ball.y - centerY;
      
      if (Math.abs(deltaX) > Math.abs(deltaY)) {
        ball.vx = -ball.vx;
      } else {
        ball.vy = -ball.vy;
      }
      
      // Push ball away from element
      ball.x += Math.sign(deltaX) * 5;
      ball.y += Math.sign(deltaY) * 5;
      
      // Visual effect on element
      collidedElement.classList.add('ball-hit');
      setTimeout(() => collidedElement.classList.remove('ball-hit'), 200);
    }
    
    // Check collision with stick figure
    if (stickFigure && checkBallStickCollision(ball)) {
      // Damage player
      takeDamage(1);
      createExplosion(ball.x, ball.y);
      
      // Remove this ball after hit
      ball.element.remove();
      const index = bouncingBalls.indexOf(ball);
      if (index > -1) bouncingBalls.splice(index, 1);
      return;
    }
    
    // Update visual position
    ball.element.style.left = ball.x - ball.radius + 'px';
    ball.element.style.top = ball.y - ball.radius + 'px';
    
    requestAnimationFrame(animate);
  };
  
  requestAnimationFrame(animate);
}

// Check collision between ball and page elements
function checkBallElementCollision(ball) {
  const elements = document.querySelectorAll('h1, h2, h3, .run-card, .project-entry, .advice-card, .total-item, button, a, img, .contribution-day');
  
  for (let element of elements) {
    const rect = element.getBoundingClientRect();
    // Keep element rect in viewport coordinates (same as ball)
    const elementRect = {
      left: rect.left,
      top: rect.top,
      right: rect.right,
      bottom: rect.bottom
    };
    
    // Check if ball overlaps with element
    if (ball.x + ball.radius > elementRect.left && 
        ball.x - ball.radius < elementRect.right && 
        ball.y + ball.radius > elementRect.top && 
        ball.y - ball.radius < elementRect.bottom) {
      return element;
    }
  }
  return null;
}

// Check collision between ball and stick figure
function checkBallStickCollision(ball) {
  const stickRect = {
    left: stickX,
    top: stickY,
    right: stickX + 30,
    bottom: stickY + 40,
    centerX: stickX + 15,
    centerY: stickY + 20
  };
  
  // Distance between ball center and stick figure center
  const deltaX = ball.x - stickRect.centerX;
  const deltaY = ball.y - stickRect.centerY;
  const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
  
  // Collision if distance is less than ball radius + stick figure "radius"
  return distance < ball.radius + 15;
}

  // Create collision effect
  function createCollisionEffect(x, y) {
    // Create multiple sparkle effects
    for (let i = 0; i < 5; i++) {
      setTimeout(() => {
        const sparkle = document.createElement('div');
        sparkle.style.cssText = `
          position: fixed;
          left: ${x + 15 + (Math.random() - 0.5) * 20}px;
          top: ${y + 20 + (Math.random() - 0.5) * 20}px;
          width: 4px;
          height: 4px;
          background: ${Math.random() > 0.5 ? 'gold' : 'var(--link-color)'};
          border-radius: 50%;
          pointer-events: none;
          z-index: 9998;
          animation: sparkle 0.8s ease-out forwards;
        `;
        document.body.appendChild(sparkle);
        setTimeout(() => sparkle.remove(), 800);
      }, i * 50);
    }
  }

  // Create explosion effect
  function createExplosion(x, y) {
    const explosion = document.createElement('div');
    explosion.className = 'explosion';
    explosion.style.left = x + 'px';
    explosion.style.top = y + 'px';
    
    // Create explosion particles
    const colors = ['#ff4444', '#ff8800', '#ffff00', '#ff0000', '#ff6600'];
    for (let i = 0; i < 12; i++) {
      const particle = document.createElement('div');
      particle.className = 'explosion-particle';
      particle.style.background = colors[Math.floor(Math.random() * colors.length)];
      
      const angle = (i / 12) * Math.PI * 2;
      const distance = 30 + Math.random() * 40;
      const dx = Math.cos(angle) * distance;
      const dy = Math.sin(angle) * distance;
      
      particle.style.setProperty('--dx', dx + 'px');
      particle.style.setProperty('--dy', dy + 'px');
      particle.style.animation = 'explode 0.6s ease-out forwards';
      
      explosion.appendChild(particle);
    }
    
    document.body.appendChild(explosion);
    setTimeout(() => explosion.remove(), 600);
  }
  
  // Create special effect for ball destruction
  function createBallDestructionEffect(x, y) {
    const explosion = document.createElement('div');
    explosion.className = 'ball-explosion';
    explosion.style.left = x + 'px';
    explosion.style.top = y + 'px';
    
    // Create orange/yellow particles for ball destruction
    const colors = ['#ff6b35', '#ff8e53', '#ffaa00', '#ffcc33', '#ff9944'];
    for (let i = 0; i < 16; i++) {
      const particle = document.createElement('div');
      particle.className = 'ball-explosion-particle';
      particle.style.background = colors[Math.floor(Math.random() * colors.length)];
      
      const angle = (i / 16) * Math.PI * 2;
      const distance = 25 + Math.random() * 35;
      const dx = Math.cos(angle) * distance;
      const dy = Math.sin(angle) * distance;
      
      particle.style.setProperty('--dx', dx + 'px');
      particle.style.setProperty('--dy', dy + 'px');
      particle.style.animation = 'ball-explode 0.8s ease-out forwards';
      
      explosion.appendChild(particle);
    }
    
    document.body.appendChild(explosion);
    setTimeout(() => explosion.remove(), 800);
  }

  // Shoot bullet
  function shootBullet() {
    if (!stickFigure || !gameActive) return;
    
    // Add shooting animation to stick figure
    stickFigure.classList.add('shooting');
    setTimeout(() => stickFigure.classList.remove('shooting'), 200);
    
    const bullet = document.createElement('div');
    bullet.className = 'bullet';
    
    // Start bullet from stick figure position
    let bulletX = stickX + 15;
    let bulletY = stickY + 15;
    
    bullet.style.left = bulletX + 'px';
    bullet.style.top = bulletY + 'px';
    document.body.appendChild(bullet);
    
    // Determine bullet direction - use current movement if moving, otherwise use last direction
    let directionX = lastDirectionX;
    let directionY = lastDirectionY;
    
    // Override with current movement if actively moving
    if (activeKeys.has('h')) {
      directionX = -1;
      directionY = 0;
    } else if (activeKeys.has('l')) {
      directionX = 1;
      directionY = 0;
    } else if (activeKeys.has('k')) {
      directionX = 0;
      directionY = -1;
    } else if (activeKeys.has('j')) {
      directionX = 0;
      directionY = 1;
    }
    
    const bulletSpeed = 8;
    
    const moveBullet = () => {
      bulletX += directionX * bulletSpeed;
      bulletY += directionY * bulletSpeed;
      
      bullet.style.left = bulletX + 'px';
      bullet.style.top = bulletY + 'px';
      
      // Check if bullet is off screen
      if (bulletX < -10 || bulletX > window.innerWidth + 10 || 
          bulletY < -10 || bulletY > window.innerHeight + 10) {
        bullet.remove();
        return;
      }
      
      // Check for collisions with bouncing balls first
      const hitBall = checkBulletBallCollision(bulletX, bulletY);
      if (hitBall) {
        // Create special explosion for ball destruction
        createBallDestructionEffect(hitBall.x, hitBall.y);
        
        // Award points for destroying ball
        addScore(75); // High points for skill-based defense
        
        // Remove the ball with flash effect
        hitBall.element.classList.add('ball-destroyed');
        setTimeout(() => {
          hitBall.element.remove();
          const ballIndex = bouncingBalls.indexOf(hitBall);
          if (ballIndex > -1) bouncingBalls.splice(ballIndex, 1);
        }, 100);
        
        bullet.remove();
        return;
      }
      
      // Check for collisions with destroyable elements
      const hitElement = checkBulletCollision(bulletX, bulletY);
      if (hitElement) {
        // Create explosion at hit point
        createExplosion(bulletX, bulletY);
        
        // Calculate score based on element type
        let points = 10;
        if (hitElement.matches('h1, h2, h3')) points = 50;
        else if (hitElement.matches('button, a')) points = 25;
        else if (hitElement.matches('img')) points = 30;
        else if (hitElement.matches('.run-card, .project-entry')) points = 40;
        
        // Add score
        addScore(points);
        
        // 25% chance to spawn a heart pickup if player not at max health
        if (playerHealth < maxHealth && Math.random() < 0.25) {
          const elemRect = hitElement.getBoundingClientRect();
          const heartX = elemRect.left + elemRect.width / 2;
          const heartY = elemRect.top + elemRect.height / 2;
          createHeartPickup(heartX, heartY);
        }
        
        // Destroy the element with animation
        hitElement.style.animation = 'element-destroy 0.5s ease-out forwards';
        setTimeout(() => {
          if (hitElement.parentNode) {
            hitElement.remove();
          }
        }, 500);
        
        bullet.remove();
        return;
      }
      
      requestAnimationFrame(moveBullet);
    };
    
    requestAnimationFrame(moveBullet);
  }

  // Check bullet collision with bouncing balls
  function checkBulletBallCollision(bulletX, bulletY) {
    for (let ball of bouncingBalls) {
      // Distance between bullet and ball center
      const deltaX = bulletX - ball.x;
      const deltaY = bulletY - ball.y;
      const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
      
      // Collision if distance is less than ball radius + bullet "radius"
      if (distance < ball.radius + 4) {
        return ball;
      }
    }
    return null;
  }

  // Check bullet collision with elements
  function checkBulletCollision(bulletX, bulletY) {
    const bulletRect = {
      left: bulletX - 3,
      top: bulletY - 1,
      right: bulletX + 3,
      bottom: bulletY + 1
    };
    
    // Get all destructible elements (same as collision targets)
    const elements = document.querySelectorAll('h1, h2, h3, .run-card, .project-entry, .advice-card, .total-item, button, a:not([href="/"]), img, span.icon, .family-photo, select, input, textarea, .contribution-day');
    
    for (let element of elements) {
      const rect = element.getBoundingClientRect();
      // Keep element rect in viewport coordinates (same as bullet)
      const elementRect = {
        left: rect.left,
        top: rect.top,
        right: rect.right,
        bottom: rect.bottom
      };
      
      // Check overlap
      if (bulletRect.left < elementRect.right && 
          bulletRect.right > elementRect.left && 
          bulletRect.top < elementRect.bottom && 
          bulletRect.bottom > elementRect.top) {
        
        return element;
      }
    }
    return null;
  }

  // Check collision with page elements
  function checkCollisions(newX, newY) {
    const stickRect = {
      left: newX,
      top: newY,
      right: newX + 30,
      bottom: newY + 40
    };
    
    // Get all interactive elements
    const elements = document.querySelectorAll('h1, h2, h3, .run-card, .project-entry, .advice-card, .total-item, button, a, img, span.icon, .family-photo, select, input, textarea, .contribution-day, .logo-img');
    
    for (let element of elements) {
      const rect = element.getBoundingClientRect();
      // Keep element rect in viewport coordinates (same as stick figure)
      const elementRect = {
        left: rect.left,
        top: rect.top,
        right: rect.right,
        bottom: rect.bottom
      };
      
      // Check overlap
      if (stickRect.left < elementRect.right && 
          stickRect.right > elementRect.left && 
          stickRect.top < elementRect.bottom && 
          stickRect.bottom > elementRect.top) {
        
        // Collision detected!
        return element;
      }
    }
    return null;
  }

  // Move stick figure with collision detection
  function moveStickFigure(deltaX, deltaY) {
    if (!stickFigure || !gameActive) return;
    
    const maxX = window.innerWidth - 30;
    const maxY = window.innerHeight - 40;
    
    const newX = Math.max(0, Math.min(maxX, stickX + deltaX));
    const newY = Math.max(0, Math.min(maxY, stickY + deltaY));
    
    // Check for collisions
    const collidedElement = checkCollisions(newX, newY);
    
    if (collidedElement) {
      // Normal collision effect
      stickFigure.classList.add('colliding');
      setTimeout(() => stickFigure.classList.remove('colliding'), 300);
      
      // Highlight the collided element
      collidedElement.classList.add('element-highlight');
      setTimeout(() => collidedElement.classList.remove('element-highlight'), 500);
      
      // Create special collision trail
      createCollisionEffect(stickX, stickY);
      
      // Bounce back slightly
      stickX = Math.max(0, Math.min(maxX, stickX - deltaX * 0.5));
      stickY = Math.max(0, Math.min(maxY, stickY - deltaY * 0.5));
    } else {
      // Normal movement
      stickX = newX;
      stickY = newY;
    }
    
    stickFigure.style.left = stickX + 'px';
    stickFigure.style.top = stickY + 'px';
    
    // Create trail when moving
    if (deltaX !== 0 || deltaY !== 0) {
      createTrail(stickX, stickY);
    }
    // In moveStickFigure, after stickFigure.style.top ... and before trail maybe, call checkHeartPickupCollision();
    checkHeartPickupCollision();
  }

  // Handle animations
  function updateAnimation() {
    if (!stickFigure) return;
    
    const isMoving = activeKeys.size > 0;
    
    if (isMoving && !isWalking) {
      stickFigure.classList.add('walking');
      isWalking = true;
    } else if (!isMoving && isWalking) {
      stickFigure.classList.remove('walking');
      isWalking = false;
    }
  }

          // Theme toggle click handler
  themeToggleBtn.addEventListener('click', function(e) {
    logoClickCount++;
    
    // Add a little bounce effect to the theme button
    themeToggleBtn.style.transform = 'scale(0.9)';
    setTimeout(() => themeToggleBtn.style.transform = 'scale(1)', 100);
    
    if (logoClickCount === 4) {
      // Cancel theme toggle and activate easter egg
      e.preventDefault();
      e.stopPropagation();
      
      // Reset click count timeout
      clearTimeout(window.logoClickTimeout);
      
      // Check if easter egg is already activated
      if (easterEggActivated) {
        logoClickCount = 0; // Reset counter
        return; // Don't activate again
      }
      
      // Activate easter egg
      easterEggActivated = true;
      // Spawn stick figure near the theme toggle button
      stickFigure = createStickFigure();
      
      // Initialize game
      gameStartTime = Date.now();
      
      // Find a good spawn position away from elements
      const spawnPos = findSafeSpawnPosition();
      stickX = spawnPos.x;
      stickY = spawnPos.y;
 
      stickFigure.style.left = stickX + 'px';
      stickFigure.style.top = stickY + 'px';
      document.body.appendChild(stickFigure);
      
      // Create and show mobile controls if on mobile
      createMobileControls();
      toggleMobileControls(true);
      
      // Create score display
      createScoreDisplay();
      
      // Create health display
      createHealthDisplay();
      
      // Start danger mechanics
      startEnemyAttacks();
      startBouncingBalls();
      
      // Celebration animation
      stickFigure.classList.add('celebrating');
      setTimeout(() => stickFigure.classList.remove('celebrating'), 1000);
      
      // Show hint
      const hint = document.createElement('div');
      hint.style.cssText = `
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        background: var(--mainbody-bg);
        color: var(--text-color);
        padding: 20px;
        border-radius: 10px;
        box-shadow: 0 4px 20px rgba(0,0,0,0.3);
        z-index: 10000;
        text-align: center;
        font-family: monospace;
        border: 2px solid var(--link-color);
      `;
                                hint.innerHTML = `
         <h3>🎉 Easter Egg Activated! 🎉</h3>
         <p><strong>Desktop:</strong> H/J/K/L (move), F (shoot), Space (jump)</p>
         <p><strong>Mobile:</strong> Use the virtual D-pad and center shoot button!</p>
         <p><strong>Combat:</strong> Shoot elements AND bouncing balls!</p>
         <p><strong>Survival:</strong> Dodge or destroy the orange bouncing balls!</p>
         <small>Survive as long as you can! Balls get more frequent over time. Click to dismiss.</small>
       `;
       document.body.appendChild(hint);
       
                    hint.addEventListener('click', () => hint.remove());
       setTimeout(() => hint.remove(), 15000); // Extended to 15 seconds
       
       // Store reference for ESC key handling
       window.currentHint = hint;
       
       logoClickCount = 0; // Reset counter
    } else {
      // Set timeout to reset click count if not clicked again soon
      clearTimeout(window.logoClickTimeout);
      window.logoClickTimeout = setTimeout(() => {
        logoClickCount = 0;
      }, 1000); // Reset after 1 second of no clicks
    }
  });

  // The tiny page-level loader sends this event after this deferred script is
  // available. Reuse the original activation path so the game behaviour stays
  // identical to the former four-click trigger.
  themeToggleBtn.addEventListener('eastereggactivate', function() {
    logoClickCount = 3;
    themeToggleBtn.click();
  });

  // Keyboard controls
  document.addEventListener('keydown', function(e) {
    const key = e.key.toLowerCase();
    
    // ESC key to dismiss instructions
    if (e.key === 'Escape' && window.currentHint) {
      window.currentHint.remove();
      window.currentHint = null;
      return;
    }
    
    if (!stickFigure) return;
    if (['h', 'j', 'k', 'l'].includes(key)) {
      e.preventDefault();
      // Dismiss instructions if visible
      if (window.currentHint) {
        window.currentHint.remove();
        window.currentHint = null;
      }
      activeKeys.add(key);
      
      const speed = 8;
      let deltaX = 0, deltaY = 0;
      
      if (activeKeys.has('h')) deltaX -= speed; // left
      if (activeKeys.has('l')) deltaX += speed; // right
      if (activeKeys.has('k')) deltaY -= speed; // up
      if (activeKeys.has('j')) deltaY += speed; // down
      
      // Update last direction when actually moving
      if (deltaX !== 0 || deltaY !== 0) {
        if (deltaX !== 0) lastDirectionX = deltaX > 0 ? 1 : -1;
        if (deltaY !== 0) lastDirectionY = deltaY > 0 ? 1 : -1;
        // If moving diagonally, prioritize horizontal direction for shooting
        if (deltaX !== 0) lastDirectionY = 0;
        else if (deltaY !== 0) lastDirectionX = 0;
      }
      
      moveStickFigure(deltaX, deltaY);
      updateAnimation();
    }
    
    // Special moves
    if (key === ' ' && stickFigure) { // spacebar for jump
      e.preventDefault();
      stickFigure.classList.add('jumping');
      setTimeout(() => stickFigure.classList.remove('jumping'), 500);
    }
    
    if (key === 'f' && stickFigure) { // F for shoot
      e.preventDefault();
      shootBullet();
    }
  });

  document.addEventListener('keyup', function(e) {
    if (!stickFigure) return;
    
    const key = e.key.toLowerCase();
    if (['h', 'j', 'k', 'l'].includes(key)) {
      activeKeys.delete(key);
      updateAnimation();
    }
  });

  // Create heart pickup
  function createHeartPickup(x, y) {
    const heart = document.createElement('div');
    heart.className = 'heart-pickup';
    heart.style.left = (x - 12) + 'px';
    heart.style.top = (y - 12) + 'px';
    heart.innerHTML = '❤️';
    document.body.appendChild(heart);
    heartPickups.push(heart);
  }

  // Check stick collision with heart pickups
  function checkHeartPickupCollision() {
    for (let i = heartPickups.length - 1; i >= 0; i--) {
      const heart = heartPickups[i];
      const rect = heart.getBoundingClientRect();
      const heartRect = {
        left: rect.left,
        top: rect.top,
        right: rect.right,
        bottom: rect.bottom
      };
      const stickRect = {
        left: stickX,
        top: stickY,
        right: stickX + 30,
        bottom: stickY + 40
      };
      if (stickRect.left < heartRect.right && stickRect.right > heartRect.left && stickRect.top < heartRect.bottom && stickRect.bottom > heartRect.top) {
        // Collected heart
        heart.remove();
        heartPickups.splice(i, 1);
        if (playerHealth < maxHealth) {
          playerHealth++;
          updateHealthDisplay();
        }
      }
    }
  }

})();
