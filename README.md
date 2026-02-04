# Adventure Game

A classic text-based adventure game engine inspired by ZX81-style gamebooks, built with Go and HTMX.

## Features

- **Interactive Story System**: YAML-based story definitions with branching narratives
- **Character Stats**: Strength, Luck, and Health with bounded values (1-12 for Strength/Luck, 0+ for Health)
- **Combat System**: Opposed-roll battles where player and enemy roll 2d6 + Strength, with multi-round interactive combat
- **Multi-Enemy Battles**: Fight 1–3 enemies (choose which to attack or use Luck on) or 4+ as a single **Horde** (combined health, mean strength for balance)
- **Luck-Based Attacks**: Special attacks that deal extra damage but reduce Luck
- **Run Away Option**: Ability to flee from battles
- **Health-Based Game Over**: Reaching 0 health triggers game over
- **Modern UI**: ZX81-inspired layout with character stats on the left, story in the center, and enemy stats on the right during battles
- **ZX81-Style Dice**: Blocky green-on-black dice in the left sidebar (your last roll, or per-stat rolls at character creation) and in the right sidebar during battle (enemy’s roll), with a short roll animation so you can verify outcomes
- **Session Management**: In-memory session store for game state persistence

## Project Structure

```
adventure/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── game/
│   │   ├── engine.go        # Core game logic and battle resolution
│   │   ├── engine_test.go   # Engine tests
│   │   ├── character.go     # Character stat rolling
│   │   ├── character_test.go # Character tests
│   │   ├── story.go         # Story YAML loading
│   │   ├── story_test.go    # Story loading tests
│   │   └── types.go         # Game data structures
│   ├── session/
│   │   ├── memory.go        # In-memory session store
│   │   ├── memory_test.go   # Session store tests
│   │   └── store.go         # Session store interface
│   └── web/
│       ├── handlers.go      # HTTP handlers for gameplay
│       ├── handlers_start.go # HTTP handlers for character creation
│       └── viewmodels.go    # View model structures
├── stories/
│   └── demo.yaml            # Demo adventure story
├── templates/
│   ├── layout.html          # Main page layout
│   ├── game.html            # Game play template
│   └── start.html           # Character creation template
├── static/
│   ├── app.css               # Application styles
│   └── js/
│       ├── app.js             # UI sync and dice animation
│       └── app.test.js        # Jest unit tests
├── package.json              # Node deps: Jest, ESLint
├── .eslintrc.cjs             # ESLint config for static/js
└── .github/
    └── workflows/
        └── test.yml         # CI/CD test workflow
```

## Requirements

- **Go** 1.22 or later
- **Node.js** 18+ and **npm** (for running JavaScript tests and linting; optional if you only run the game)
- A modern web browser (for playing the game)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd adventure
```

2. Install Go module dependencies:
```bash
go mod download
```

3. (Optional) Install code quality tools for development:
```bash
make install-tools
```

This installs:
- `golangci-lint` - Comprehensive linter aggregator
- `staticcheck` - Advanced static analysis
- `goimports` - Code formatting tool

**Important**: After installing tools, ensure `$GOPATH/bin` (or `$HOME/go/bin` if GOPATH is unset) is in your PATH:
```bash
# Check your GOPATH
go env GOPATH

# Add to PATH (add to ~/.bashrc or ~/.zshrc to make permanent)
export PATH=$PATH:$(go env GOPATH)/bin
```

The `make install-tools` command will show you the exact path and instructions if it's not already in your PATH.

## Running the Game

Start the server:
```bash
go run cmd/server/main.go
```

The game will be available at `http://localhost:8080`

### Docker

Build and run with Docker (app listens on port 8080 inside the container):

```bash
docker build -t adventure .
docker run -p 8080:8080 adventure
```

Then open `http://localhost:8080`.

## Running Tests

### Go tests

Run all Go tests (with race detection if gcc/CGO is available):
```bash
make test
```

Or run Go tests directly:
```bash
go test ./...
```

### JavaScript tests

The UI logic in `static/js/` (sidebar sync, dice animation) is tested with Jest. Install JS dependencies once, then run tests:

```bash
make install-js   # or: npm install
make test-js      # or: npm test
```

Run JS tests in watch mode:
```bash
npm run test:watch
```

Run tests with race detector (requires gcc/CGO):
```bash
make test-race
# or
CGO_ENABLED=1 go test -v -race ./...
```

**Note**: Race detection requires CGO, which needs a C compiler (gcc). If gcc is not installed:
- On Ubuntu/Debian: `sudo apt-get install gcc`
- On macOS: `xcode-select --install` (includes gcc via Xcode Command Line Tools)
- On Fedora/RHEL: `sudo dnf install gcc`

If gcc is not available, tests will run without race detection automatically.

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests with race detection:
```bash
go test -race ./...
```

Run tests with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Run tests for a specific package:
```bash
go test ./internal/game/...
go test ./internal/session/...
```

## Game Mechanics

### Character Creation

- Players start by rolling stats:
  - **Strength**: 2d6 (range: 2-12, clamped to 1-12)
  - **Luck**: 2d6 (range: 2-12, clamped to 1-12)
  - **Health**: 2d6 + 6 (range: 8-18)
- Players can reroll stats before beginning their adventure

### Stat Rules

- **Strength** and **Luck**: Always clamped between 1 and 12
- **Health**: Minimum 0 (death), no maximum cap
- Stats can be modified by story events (effects)
- Effects can specify `clampMin` and `clampMax` values, but global bounds are always enforced

### Combat System

Combat uses opposed rolls:
- **Player Total** = Strength + 2d6
- **Enemy Total** = Enemy Strength + 2d6
- Higher total deals 1 damage to the loser
- Ties result in no damage

**Multi-enemy battles:**
- **1–3 enemies**: Each enemy is shown with name, strength, and health. You choose which to **Attack** or use **Luck** on each round.
- **4+ enemies**: Shown as a single **Horde** with combined health and **mean strength** (average of all enemies) so large groups stay winnable.

**Combat Actions:**
- **Attack**: Standard attack on chosen enemy (1 damage on hit)
- **Luck Attack**: Spend 1 Luck to deal 2 damage on chosen enemy (Luck clamped to minimum 1)
- **Run Away**: Flee from battle (enemy state is cleared)

Battles continue round-by-round until:
- All enemies’ health reaches 0 (victory)
- Player health reaches 0 (defeat/death)

### Dice display

- **Left sidebar**: Your last 2d6 roll is always shown (or, on character creation, one 2d6 pair per stat: Strength, Luck, Health). The display persists until the next roll.
- **Right sidebar**: During battle, the enemy’s 2d6 roll is shown so you can see both totals and verify who won the round.
- Dice use a ZX81-style blocky pip display (CSS only, no images). A brief “roll” animation plays when new dice appear.

### Game Over

- Health reaching 0 triggers automatic game over
- Player is routed to the `death` node if it exists in the story
- Game can be restarted from the death screen

### Printable map

- During play, use **Download map** to get a PDF map of the current adventure (all locations and paths, with your current location marked). The map uses an old-map style and is intended for printing. The route is `GET /map`; the same session cookie as play is used.

## Story Format

Stories are defined in YAML format. See `stories/demo.yaml` for a complete example.

### Node Structure

```yaml
nodes:
  node_id:
    text: "Story text displayed to player"
    choices:
      - key: "choice_key"
        text: "Choice text"
        next: "destination_node"
    effects:
      - op: "add"
        stat: "health"
        value: -2
    ending: false
```

### Choices

Choices can include:
- **Simple navigation**: `next` field
- **Stat checks**: `check` with `onSuccessNext` and `onFailureNext`
- **Prompted answers**: `prompt` with `answers` mapping to `next` nodes
- **Effects**: Stat modifications applied when choice is selected
- **Battles**: `battle` block for combat encounters
- **Mode**: `battle_attack` or `battle_luck` for combat actions

### Battle Definition

Single enemy (legacy style):

```yaml
battle:
  enemyName: "Goblin"
  enemyStrength: 8
  enemyHealth: 3
  onVictoryNext: "victory_node"
  onDefeatNext: "defeat_node"
```

Multiple enemies (1–3 shown individually, 4+ as a Horde):

```yaml
battle:
  enemies:
    - name: "Goblin"
      strength: 6
      health: 3
    - name: "Orc"
      strength: 8
      health: 4
  onVictoryNext: "victory_node"
  onDefeatNext: "defeat_node"
```

Battle choices are generated automatically (e.g. “Attack Goblin”, “Luck Orc”, “Run away”). Horde strength is the **mean** of all enemy strengths; health is the sum.

### Prompted answers

Use a `prompt` block on a choice to accept a typed answer and route to a different node.
Answers are normalized (trimmed, case-insensitive, punctuation ignored). If no
answer matches, `defaultNext` is used; otherwise `next` acts as a fallback.
`prompt` is mutually exclusive with `check` and `battle` on the same choice:
when `prompt` is present, `check` and `battle` are ignored.

```yaml
choices:
  - key: "riddle"
    text: "Answer the riddle"
    prompt:
      question: "I speak without a mouth and hear without ears. What am I?"
      placeholder: "Your answer"
      answers:
        - match: "echo"
          next: "riddle_success"
        - matches: ["shadow", "a shadow"]
          next: "riddle_wrong"
      defaultNext: "riddle_wrong"
```

### Effects

Effects modify player stats:
```yaml
effects:
  - op: "add"
    stat: "strength"  # or "luck" or "health"
    value: 1
    clampMax: 12      # Optional: maximum value
    clampMin: 1       # Optional: minimum value
```

### Scenery and animations

Each node can optionally set a **scenery** value so the story area shows a backdrop image. Story text appears in a strip along the bottom and scrolls when long.

**Scenery images** are image-based and linked from the story YAML. Each story has a strict directory for its scenery: `stories/<story_id>/scenery/`. In the YAML, set `scenery` to the **filename** (with or without extension) of an image in that directory, e.g. `scenery: "forest"` or `scenery: "forest.png"`. The server looks for that file and tries `.png`, `.jpg`, and `.jpeg` if no extension is given. Only files under `stories/<story_id>/scenery/` are served (no path traversal). If the file is missing, the request returns 404 and the UI may show a CSS fallback. Omit `scenery` or use `default` to request `default.png` (or `default.jpg`) from the same directory.

**Scene audio**: Each node can optionally set an **audio** value so the scene plays a looping ambient track. Audio files live in `stories/<story_id>/audio/`. In the YAML, set `audio` to the **filename** (with or without extension), e.g. `audio: "forest_ambient"` or `audio: "forest_ambient.mp3"`. The server serves them at `/audio/<storyID>/<filename>` and tries `.mp3`, `.ogg`, `.wav`, and `.m4a` if no extension is given. Only files under `stories/<story_id>/audio/` are served (no path traversal). The UI uses a single shared audio element: when you navigate to a node with `audio` set, the previous track stops and the new one plays (looped, at 50% volume). Omit `audio` for no scene music.

**Entry animations**: Optional `entry_animation` plays when entering the node (e.g. going through a door):

- `door_open` — short “door opening” effect when entering an interior

**Example:**

```yaml
nodes:
  camp:
    text: "You wake at the edge of a quiet camp."
    scenery: "clearing"
    choices: [...]
  forest:
    text: "The forest is cold and still."
    scenery: "forest"
    choices: [...]
  cottage_inside:
    text: "You step inside. The door swings shut."
    scenery: "house_inside"
    entry_animation: "door_open"
    choices: [...]
```

## Development

### Adding New Features

1. Game logic: Modify `internal/game/engine.go`
2. Story loading: Modify `internal/game/story.go` and `types.go`
3. Web handlers: Modify `internal/web/handlers.go`
4. UI: Modify templates in `templates/` and styles in `static/app.css`

### Code Quality Tools

The project uses several static analysis tools to maintain code quality. These are external binaries (not Go modules) that need to be installed separately.

**Install all tools at once:**
```bash
make install-tools
```

**Individual tools:**

**golangci-lint**: Comprehensive linter aggregator
```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run
golangci-lint run --out-format=colored-line-number ./...
```

**go vet**: Built-in Go static analyzer (no installation needed)
```bash
go vet ./...
```

**staticcheck**: Advanced static analysis
```bash
# Install
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run
staticcheck ./...
```

**Makefile**: Convenient commands for common tasks
```bash
make install-tools  # Install Go linting tools (golangci-lint, staticcheck, goimports)
make install-js     # Install JS dependencies (Jest, ESLint)
make check          # Run all checks (Go + JS: format, vet, lint, test)
make test           # Run Go tests with race detection
make test-js        # Run JavaScript unit tests (Jest)
make test-short     # Run Go tests without race (faster)
make lint           # Run golangci-lint (Go)
make lint-js        # Run ESLint on static/js
make fmt            # Format Go code
make vet            # Run go vet
make staticcheck    # Run staticcheck
make build          # Build the application
make run            # Run the application
make clean          # Remove bin/ and coverage.out
```

**JavaScript**: Linting and testing
```bash
npm install        # Install dependencies (or make install-js)
npm test           # Run Jest tests
npm run lint       # Run ESLint on static/js
npm run lint:fix   # ESLint with auto-fix
```

**Note**: The Go linting tools (`golangci-lint`, `staticcheck`, `goimports`) are not Go module dependencies. They are standalone binaries installed via `go install` and stored in `$GOPATH/bin` or `$HOME/go/bin`. Make sure this directory is in your `PATH`. JavaScript tools (Jest, ESLint) are installed via `npm install` and live in `node_modules/`.

### Testing

**Go**: All game logic should have corresponding tests:
- Engine logic: `internal/game/engine_test.go`
- Character stats: `internal/game/character_test.go`
- Story loading: `internal/game/story_test.go`
- Session store: `internal/session/memory_test.go`

**JavaScript**: UI logic in `static/js/` (sidebar sync, dice animation, HTMX glue) is covered by Jest in `static/js/app.test.js`. Run with `make test-js` or `npm test`.

**CI**: Go and JS tests, plus Go formatting (gofmt), vet, staticcheck, golangci-lint and ESLint, run automatically on pull requests and pushes to `main`/`master` via GitHub Actions (`.github/workflows/test.yml`).

### Performance Considerations

- Choice lookup uses linear search (O(n)) which is acceptable as nodes typically have < 10 choices
- Session store uses map lookups (O(1)) for efficient state retrieval
- Battle resolution is O(1) per round (no loops)
- Effect application is O(n) where n is the number of effects (typically < 5)

## License

See LICENSE file for details.
