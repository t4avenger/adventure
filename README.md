# Adventure Game

A classic text-based adventure game engine inspired by ZX81-style gamebooks, built with Go and HTMX.

## Features

- **Interactive Story System**: YAML-based story definitions with branching narratives
- **Character Stats**: Strength, Luck, and Health with bounded values (1-12 for Strength/Luck, 0+ for Health)
- **Combat System**: Opposed-roll battles where player and enemy roll 2d6 + Strength, with multi-round interactive combat
- **Luck-Based Attacks**: Special attacks that deal extra damage but reduce Luck
- **Run Away Option**: Ability to flee from battles
- **Health-Based Game Over**: Reaching 0 health triggers game over
- **Modern UI**: ZX81-inspired layout with character stats on the left, story in the center, and enemy stats on the right during battles
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
│   └── app.css              # Application styles
└── .github/
    └── workflows/
        └── test.yml         # CI/CD test workflow
```

## Requirements

- Go 1.22 or later
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

## Running Tests

Run all tests (with race detection if gcc/CGO is available):
```bash
make test
```

Or run tests directly:
```bash
go test ./...
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

**Combat Actions:**
- **Attack**: Standard attack (1 damage on hit)
- **Luck Attack**: Spend 1 Luck to deal 2 damage on hit (Luck clamped to minimum 1)
- **Run Away**: Flee from battle (enemy state is cleared)

Battles continue round-by-round until:
- Enemy health reaches 0 (victory)
- Player health reaches 0 (defeat/death)

### Game Over

- Health reaching 0 triggers automatic game over
- Player is routed to the `death` node if it exists in the story
- Game can be restarted from the death screen

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
- **Effects**: Stat modifications applied when choice is selected
- **Battles**: `battle` block for combat encounters
- **Mode**: `battle_attack` or `battle_luck` for combat actions

### Battle Definition

```yaml
choices:
  - key: "fight"
    text: "Attack the enemy"
    mode: "battle_attack"
    battle:
      enemyName: "Goblin"
      enemyStrength: 8
      enemyHealth: 3
      onVictoryNext: "victory_node"
      onDefeatNext: "defeat_node"
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
make install-tools  # Install all linting tools
make check          # Run all checks (format, vet, lint, test)
make lint           # Run golangci-lint
make fmt            # Format code
make vet            # Run go vet
make staticcheck    # Run staticcheck
make test           # Run tests with race detection
make build          # Build the application
make run            # Run the application
```

**Note**: The linting tools (`golangci-lint`, `staticcheck`, `goimports`) are not Go module dependencies. They are standalone binaries installed via `go install` and stored in `$GOPATH/bin` or `$HOME/go/bin`. Make sure this directory is in your `PATH`.

### Testing

All game logic should have corresponding tests:
- Engine logic: `internal/game/engine_test.go`
- Character stats: `internal/game/character_test.go`
- Story loading: `internal/game/story_test.go`
- Session store: `internal/session/memory_test.go`

Tests run automatically on pull requests via GitHub Actions, along with linting and static analysis checks.

### Performance Considerations

- Choice lookup uses linear search (O(n)) which is acceptable as nodes typically have < 10 choices
- Session store uses map lookups (O(1)) for efficient state retrieval
- Battle resolution is O(1) per round (no loops)
- Effect application is O(n) where n is the number of effects (typically < 5)

## License

See LICENSE file for details.
