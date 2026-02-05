/**
 * Unit tests for Adventure UI (sidebar and dice sync).
 * Run: npm test
 */
let AdventureUI;

function buildDom() {
  const container = document.createElement('div');
  container.innerHTML =
    '<main class="wrap">' +
    '  <div class="container">' +
    '    <aside id="sidebar-left" class="sidebar">' +
    '      <div class="character-stats">' +
    '        <div>Strength: <strong id="stat-strength">0</strong></div>' +
    '        <div>Luck: <strong id="stat-luck">0</strong></div>' +
    '        <div>Health: <strong id="stat-health">0</strong></div>' +
    '      </div>' +
    '      <div class="player-dice-last" style="display:none;">' +
    '        <div class="dice-pair"><div class="die zx81-die" data-face="1"></div><div class="die zx81-die" data-face="1"></div></div>' +
    '      </div>' +
    '      <div class="player-dice-stats" style="display:none;">' +
    '        <div class="dice-pair"><div class="die zx81-die"></div><div class="die zx81-die"></div></div>' +
    '        <div class="dice-pair"><div class="die zx81-die"></div><div class="die zx81-die"></div></div>' +
    '        <div class="dice-pair"><div class="die zx81-die"></div><div class="die zx81-die"></div></div>' +
    '      </div>' +
    '    </aside>' +
    '    <section id="game" class="main-content"></section>' +
    '    <aside id="sidebar-right" class="enemy-sidebar" style="display:none;">' +
    '      <div class="enemy-dice-area" style="display:none;">' +
    '        <div class="dice-pair"><div class="die zx81-die"></div><div class="die zx81-die"></div></div>' +
    '      </div>' +
    '      <div class="enemy-panels">' +
    '        <div class="enemy-panel"><div class="enemy-stats"><div>Name: <strong></strong></div><div>Strength: <strong></strong></div><div>Health: <strong></strong></div></div></div>' +
    '        <div class="enemy-panel"><div class="enemy-stats"><div>Name: <strong></strong></div><div>Strength: <strong></strong></div><div>Health: <strong></strong></div></div></div>' +
    '      </div>' +
    '    </aside>' +
    '  </div>' +
    '</main>';
  document.body.appendChild(container);
  return container;
}

function clearGameContent() {
  const game = document.getElementById('game');
  if (game) game.innerHTML = '';
}

describe('AdventureUI', function () {
  let container;

  beforeEach(function () {
    jest.resetModules();
    AdventureUI = require('./app.js');
    document.body.innerHTML = '';
    container = buildDom();
  });

  afterEach(function () {
    if (container && container.parentNode) container.parentNode.removeChild(container);
  });

  describe('updateSidebarStats', function () {
    it('updates sidebar from #game .stats-update data attributes', function () {
      const game = document.getElementById('game');
      game.innerHTML = '<div class="stats-update" data-strength="10" data-luck="8" data-health="14" style="display:none;"></div>';
      AdventureUI.updateSidebarStats();
      expect(document.getElementById('stat-strength').textContent).toBe('10');
      expect(document.getElementById('stat-luck').textContent).toBe('8');
      expect(document.getElementById('stat-health').textContent).toBe('14');
    });

    it('does nothing when stats-update is missing', function () {
      AdventureUI.updateSidebarStats();
      expect(document.getElementById('stat-strength').textContent).toBe('0');
    });
  });

  describe('updateEnemySidebar', function () {
    it('shows enemy sidebar and fills one panel when enemy-update present', function () {
      const game = document.getElementById('game');
      game.innerHTML = '<div class="enemy-update" data-enemy-name="Goblin" data-enemy-strength="6" data-enemy-health="3" style="display:none;"></div>';
      const enemySidebar = document.querySelector('.enemy-sidebar');
      AdventureUI.updateEnemySidebar();
      expect(enemySidebar.classList.contains('show')).toBe(true);
      expect(enemySidebar.style.display).toBe('flex');
      const panel = enemySidebar.querySelector('.enemy-panel');
      const nameEl = panel.querySelector('.enemy-stats div:first-child strong');
      expect(nameEl.textContent).toBe('Goblin');
    });

    it('hides enemy sidebar when no valid enemy-update', function () {
      const game = document.getElementById('game');
      game.innerHTML = '<div class="enemy-update" data-enemy-name="" data-enemy-health="0" style="display:none;"></div>';
      AdventureUI.updateEnemySidebar();
      expect(document.querySelector('.enemy-sidebar.show')).toBeNull();
    });
  });

  describe('setDieFace', function () {
    it('sets data-face to value 1-6 without animation', function () {
      const die = document.createElement('div');
      die.className = 'die';
      AdventureUI.setDieFace(die, 4, false);
      expect(die.getAttribute('data-face')).toBe('4');
    });

    it('clamps value to 1-6', function () {
      const die = document.createElement('div');
      AdventureUI.setDieFace(die, 0, false);
      expect(die.getAttribute('data-face')).toBe('1');
      AdventureUI.setDieFace(die, 10, false);
      expect(die.getAttribute('data-face')).toBe('6');
    });

    it('does nothing when el is null', function () {
      expect(function () { AdventureUI.setDieFace(null, 3, false); }).not.toThrow();
    });
  });

  describe('updatePlayerDice', function () {
    it('shows last-roll dice when player-dice-update present', function () {
      jest.useFakeTimers();
      const game = document.getElementById('game');
      game.innerHTML = '<div class="player-dice-update" data-dice1="3" data-dice2="5" style="display:none;"></div>';
      const lastSection = document.querySelector('.player-dice-last');
      const lastDice = document.querySelectorAll('.player-dice-last .dice-pair .die');
      AdventureUI.updatePlayerDice();
      jest.advanceTimersByTime(300);
      expect(lastSection.style.display).toBe('block');
      expect(lastDice[0].getAttribute('data-face')).toBe('3');
      expect(lastDice[1].getAttribute('data-face')).toBe('5');
      jest.useRealTimers();
    });

    it('shows stat-rolls dice when stat-rolls present', function () {
      const game = document.getElementById('game');
      game.innerHTML =
        '<div class="stat-rolls" data-strength-dice1="2" data-strength-dice2="4" data-luck-dice1="1" data-luck-dice2="6" data-health-dice1="3" data-health-dice2="3" style="display:none;"></div>';
      const statsSection = document.querySelector('.player-dice-stats');
      const statsDice = document.querySelectorAll('.player-dice-stats .dice-pair .die');
      AdventureUI.updatePlayerDice();
      expect(statsSection.style.display).toBe('block');
      expect(statsDice[0].getAttribute('data-face')).toBe('2');
      expect(statsDice[1].getAttribute('data-face')).toBe('4');
      expect(statsDice[4].getAttribute('data-face')).toBe('3');
      expect(statsDice[5].getAttribute('data-face')).toBe('3');
    });
  });

  describe('updateEnemyDice', function () {
    it('shows enemy dice area and sets faces when enemy-dice-update present', function () {
      jest.useFakeTimers();
      const game = document.getElementById('game');
      game.innerHTML = '<div class="enemy-dice-update" data-dice1="2" data-dice2="6" style="display:none;"></div>';
      const enemyDiceArea = document.querySelector('.enemy-dice-area');
      const enemyDice = document.querySelectorAll('.enemy-dice-area .dice-pair .die');
      AdventureUI.updateEnemyDice();
      jest.advanceTimersByTime(300);
      expect(enemyDiceArea.style.display).toBe('block');
      expect(enemyDice[0].getAttribute('data-face')).toBe('2');
      expect(enemyDice[1].getAttribute('data-face')).toBe('6');
      jest.useRealTimers();
    });

    it('hides enemy dice area when no enemy-dice-update', function () {
      clearGameContent();
      const enemyDiceArea = document.querySelector('.enemy-dice-area');
      AdventureUI.updateEnemyDice();
      expect(enemyDiceArea.style.display).toBe('none');
    });
  });

  describe('updateSceneAudio', function () {
    let mockAudio;

    beforeEach(function () {
      mockAudio = {
        pause: jest.fn(),
        load: jest.fn(),
        currentTime: 0,
        src: '',
        loop: false,
        volume: 1,
        play: jest.fn(function () { return Promise.resolve(); }),
        removeAttribute: jest.fn()
      };
      global.Audio = jest.fn(function () { return mockAudio; });
    });

    it('starts playback when data-audio-url is present', function () {
      const game = document.getElementById('game');
      game.innerHTML = '<div data-audio-url="/audio/demo/ambient"></div>';
      AdventureUI.updateSceneAudio();
      expect(global.Audio).toHaveBeenCalledTimes(1);
      expect(mockAudio.src).toBe('/audio/demo/ambient');
      expect(mockAudio.loop).toBe(true);
      expect(mockAudio.volume).toBe(0.5);
      expect(mockAudio.play).toHaveBeenCalledTimes(1);
    });

    it('does not restart when URL is unchanged', function () {
      const game = document.getElementById('game');
      game.innerHTML = '<div data-audio-url="/audio/demo/ambient"></div>';
      AdventureUI.updateSceneAudio();
      mockAudio.pause.mockClear();
      mockAudio.currentTime = 7;
      AdventureUI.updateSceneAudio();
      expect(mockAudio.pause).not.toHaveBeenCalled();
      expect(mockAudio.currentTime).toBe(7);
    });

    it('stops playback and clears src when audio removed', function () {
      const game = document.getElementById('game');
      game.innerHTML = '<div data-audio-url="/audio/demo/ambient"></div>';
      AdventureUI.updateSceneAudio();
      mockAudio.pause.mockClear();
      mockAudio.removeAttribute.mockClear();
      mockAudio.load.mockClear();
      game.innerHTML = '';
      AdventureUI.updateSceneAudio();
      expect(mockAudio.pause).toHaveBeenCalledTimes(1);
      expect(mockAudio.removeAttribute).toHaveBeenCalledWith('src');
      expect(mockAudio.load).toHaveBeenCalledTimes(1);
    });
  });

  describe('runUpdaters', function () {
    beforeEach(function () {
      global.Audio = jest.fn(function () {
        return {
          pause: function () {},
          load: function () {},
          currentTime: 0,
          src: '',
          loop: false,
          volume: 1,
          play: function () { return Promise.resolve(); },
          removeAttribute: function () {}
        };
      });
    });

    it('runs all updaters without throwing', function () {
      document.getElementById('game').innerHTML = '<div class="stats-update" data-strength="7" data-luck="7" data-health="12" style="display:none;"></div>';
      expect(function () { AdventureUI.runUpdaters(); }).not.toThrow();
      expect(document.getElementById('stat-strength').textContent).toBe('7');
    });
  });

  describe('animateSidebarDice', function () {
    it('animates visible player and enemy dice by data-face', function () {
      jest.useFakeTimers();
      const leftLast = document.querySelector('#sidebar-left .player-dice-last');
      leftLast.style.display = 'block';
      const leftDice = leftLast.querySelectorAll('.dice-pair .die');
      leftDice[0].setAttribute('data-face', '3');
      leftDice[1].setAttribute('data-face', '5');
      AdventureUI.animateSidebarDice();
      jest.advanceTimersByTime(300);
      expect(leftDice[0].getAttribute('data-face')).toBe('3');
      expect(leftDice[1].getAttribute('data-face')).toBe('5');
      jest.useRealTimers();
    });
  });

  describe('startStoryTextAutoScroll', function () {
    it('does nothing when #game has no story-text-strip', function () {
      clearGameContent();
      expect(function () { AdventureUI.startStoryTextAutoScroll(); }).not.toThrow();
    });

    it('does nothing when story-text-strip content does not overflow', function () {
      const game = document.getElementById('game');
      const strip = document.createElement('div');
      strip.className = 'story-text-strip';
      strip.scrollTop = 99;
      Object.defineProperty(strip, 'scrollHeight', { value: 100, configurable: true });
      Object.defineProperty(strip, 'clientHeight', { value: 100, configurable: true });
      game.appendChild(strip);
      AdventureUI.startStoryTextAutoScroll();
      expect(strip.scrollTop).toBe(99);
    });

    it('resets scrollTop and starts scroll when content overflows', function () {
      jest.useFakeTimers();
      const game = document.getElementById('game');
      const strip = document.createElement('div');
      strip.className = 'story-text-strip';
      strip.scrollTop = 50;
      Object.defineProperty(strip, 'scrollHeight', { value: 200, configurable: true });
      Object.defineProperty(strip, 'clientHeight', { value: 80, configurable: true });
      game.appendChild(strip);
      AdventureUI.startStoryTextAutoScroll();
      expect(strip.scrollTop).toBe(0);
      jest.advanceTimersByTime(800);
      expect(strip.scrollTop).toBeGreaterThan(0);
      jest.useRealTimers();
    });
  });
});
