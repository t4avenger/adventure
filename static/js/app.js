/**
 * Adventure UI: syncs sidebars and dice from #game fragment after HTMX swaps.
 * Exported for testing; in browser attached as window.AdventureUI.
 */
(function (global) {
  'use strict';

  function updateSidebarStats() {
    const statsEl = document.querySelector('#game .stats-update') || document.querySelector('.stats-update');
    if (statsEl) {
      const strength = statsEl.getAttribute('data-strength');
      const luck = statsEl.getAttribute('data-luck');
      const health = statsEl.getAttribute('data-health');
      const sidebarStrength = document.querySelector('#stat-strength');
      const sidebarLuck = document.querySelector('#stat-luck');
      const sidebarHealth = document.querySelector('#stat-health');
      if (sidebarStrength && strength !== null) sidebarStrength.textContent = strength;
      if (sidebarLuck && luck !== null) sidebarLuck.textContent = luck;
      if (sidebarHealth && health !== null) sidebarHealth.textContent = health;
    }
  }

  function updateEnemySidebar() {
    const gameEl = document.querySelector('#game');
    const enemyUpdates = gameEl ? gameEl.querySelectorAll('.enemy-update') : document.querySelectorAll('.enemy-update');
    const enemySidebar = document.querySelector('.enemy-sidebar');
    if (!enemySidebar) return;

    enemySidebar.style.display = 'none';
    const panels = enemySidebar.querySelectorAll('.enemy-panel');
    panels.forEach(function (p) { p.style.display = 'none'; });

    const list = [];
    for (let i = 0; i < enemyUpdates.length; i++) {
      const el = enemyUpdates[i];
      const name = (el.getAttribute('data-enemy-name') || '').trim();
      const strength = parseInt(el.getAttribute('data-enemy-strength') || '0', 10);
      const health = parseInt(el.getAttribute('data-enemy-health') || '0', 10);
      if (name !== '' && !isNaN(health) && health > 0) {
        list.push({ name: name, strength: isNaN(strength) ? 0 : strength, health: health });
      }
    }

    if (list.length > 0) {
      enemySidebar.classList.add('show');
      enemySidebar.style.display = 'flex';
      const showCount = list.length;
      for (let i = 0; i < panels.length; i++) {
        const panel = panels[i];
        if (i < showCount) {
          panel.style.display = 'flex';
          const e = list[i];
          const nameEl = panel.querySelector('.enemy-stats div:first-child strong');
          const strengthEl = panel.querySelector('.enemy-stats div:nth-child(2) strong');
          const healthEl = panel.querySelector('.enemy-stats div:nth-child(3) strong');
          if (nameEl) nameEl.textContent = e.name;
          if (strengthEl) strengthEl.textContent = String(e.strength);
          if (healthEl) healthEl.textContent = String(e.health);
        }
      }
    } else {
      enemySidebar.classList.remove('show');
    }
  }

  function setDieFace(dieEl, face, animate) {
    if (!dieEl) return;
    face = Math.min(6, Math.max(1, parseInt(face, 10) || 1));
    if (animate) {
      let steps = 0;
      const maxSteps = 4;
      const interval = setInterval(function () {
        dieEl.setAttribute('data-face', String(Math.floor(Math.random() * 6) + 1));
        steps++;
        if (steps >= maxSteps) {
          clearInterval(interval);
          dieEl.setAttribute('data-face', String(face));
        }
      }, 60);
    } else {
      dieEl.setAttribute('data-face', String(face));
    }
  }

  function updatePlayerDice() {
    const gameEl = document.querySelector('#game');
    if (!gameEl) return;
    const playerLast = gameEl.querySelector('.player-dice-update');
    const statRolls = gameEl.querySelector('.stat-rolls');
    const lastSection = document.querySelector('.player-dice-last');
    const statsSection = document.querySelector('.player-dice-stats');
    const lastDice = document.querySelectorAll('.player-dice-last .dice-pair .die');
    const statsDice = document.querySelectorAll('.player-dice-stats .dice-pair .die');
    if (playerLast) {
      const d1 = parseInt(playerLast.getAttribute('data-dice1') || '1', 10);
      const d2 = parseInt(playerLast.getAttribute('data-dice2') || '1', 10);
      if (lastSection) lastSection.style.display = 'block';
      if (statsSection) statsSection.style.display = 'none';
      if (lastDice.length >= 2) {
        setDieFace(lastDice[0], d1, true);
        setDieFace(lastDice[1], d2, true);
      }
    } else if (statRolls) {
      const sd1 = parseInt(statRolls.getAttribute('data-strength-dice1') || '1', 10);
      const sd2 = parseInt(statRolls.getAttribute('data-strength-dice2') || '1', 10);
      const ld1 = parseInt(statRolls.getAttribute('data-luck-dice1') || '1', 10);
      const ld2 = parseInt(statRolls.getAttribute('data-luck-dice2') || '1', 10);
      const hd1 = parseInt(statRolls.getAttribute('data-health-dice1') || '1', 10);
      const hd2 = parseInt(statRolls.getAttribute('data-health-dice2') || '1', 10);
      if (lastSection) lastSection.style.display = 'none';
      if (statsSection) statsSection.style.display = 'block';
      if (statsDice.length >= 6) {
        setDieFace(statsDice[0], sd1, false);
        setDieFace(statsDice[1], sd2, false);
        setDieFace(statsDice[2], ld1, false);
        setDieFace(statsDice[3], ld2, false);
        setDieFace(statsDice[4], hd1, false);
        setDieFace(statsDice[5], hd2, false);
      }
    }
  }

  function updateEnemyDice() {
    const gameEl = document.querySelector('#game');
    const enemyUpdate = gameEl ? gameEl.querySelector('.enemy-dice-update') : null;
    const enemyDiceArea = document.querySelector('.enemy-dice-area');
    const enemyDice = document.querySelectorAll('.enemy-dice-area .dice-pair .die');
    if (enemyUpdate && enemyDiceArea) {
      const d1 = parseInt(enemyUpdate.getAttribute('data-dice1') || '1', 10);
      const d2 = parseInt(enemyUpdate.getAttribute('data-dice2') || '1', 10);
      enemyDiceArea.style.display = 'block';
      if (enemyDice.length >= 2) {
        setDieFace(enemyDice[0], d1, true);
        setDieFace(enemyDice[1], d2, true);
      }
    } else if (enemyDiceArea) {
      enemyDiceArea.style.display = 'none';
    }
  }

  var sceneAudio = null;

  function getSceneAudio() {
    if (!sceneAudio) {
      sceneAudio = typeof Audio !== 'undefined' ? new Audio() : null;
    }
    return sceneAudio;
  }

  function updateSceneAudio() {
    const gameEl = document.querySelector('#game');
    const container = gameEl ? gameEl.querySelector('[data-audio-url]') : null;
    const url = container ? container.getAttribute('data-audio-url') : null;
    const audio = getSceneAudio();
    if (!audio) return;
    try {
      if (url) {
        audio.pause();
        audio.currentTime = 0;
        audio.src = url;
        audio.loop = true;
        audio.volume = 0.5;
        audio.play().catch(function () {
          var once = function () {
            document.body.removeEventListener('click', once, true);
            audio.play().catch(function () {});
          };
          document.body.addEventListener('click', once, true);
        });
      } else {
        audio.pause();
        audio.currentTime = 0;
        audio.removeAttribute('src');
        audio.load();
      }
    } catch (e) {
      /* HTMLMediaElement.pause/load not implemented in some environments (e.g. jsdom) */
    }
  }

  function runUpdaters() {
    updateSidebarStats();
    updateEnemySidebar();
    updatePlayerDice();
    updateEnemyDice();
    updateSceneAudio();
  }

  /** Run dice roll animation on sidebar dice (used after OOB swap; server already set content). */
  function animateSidebarDice() {
    const leftLast = document.querySelector('#sidebar-left .player-dice-last');
    if (leftLast && leftLast.style.display !== 'none') {
      const dice = leftLast.querySelectorAll('.dice-pair .die');
      for (let i = 0; i < dice.length; i++) {
        const face = dice[i].getAttribute('data-face') || '1';
        setDieFace(dice[i], face, true);
      }
    }
    const enemyArea = document.querySelector('#sidebar-right .enemy-dice-area');
    if (enemyArea && enemyArea.style.display !== 'none') {
      const dice = enemyArea.querySelectorAll('.dice-pair .die');
      for (let i = 0; i < dice.length; i++) {
        const face = dice[i].getAttribute('data-face') || '1';
        setDieFace(dice[i], face, true);
      }
    }
  }

  function init() {
    runUpdaters();
    document.body.addEventListener('htmx:afterSwap', function (evt) {
      if (evt.detail && evt.detail.target && evt.detail.target.id !== 'game') return;
      runUpdaters();
      const xhr = evt.detail && evt.detail.xhr;
      const isOOB = xhr && xhr.getResponseHeader && xhr.getResponseHeader('X-Adventure-OOB') === 'true';
      if (isOOB) {
        setTimeout(animateSidebarDice, 10);
      }
    });
  }

  const AdventureUI = {
    updateSidebarStats,
    updateEnemySidebar,
    setDieFace,
    updatePlayerDice,
    updateEnemyDice,
    runUpdaters,
    updateSceneAudio,
    animateSidebarDice,
    init
  };

  if (typeof module !== 'undefined' && module.exports) {
    module.exports = AdventureUI;
  } else {
    global.AdventureUI = AdventureUI;
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', init);
    } else {
      init();
    }
  }
})(typeof window !== 'undefined' ? window : this);
