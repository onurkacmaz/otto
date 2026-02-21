# otto

A terminal-based database client written in Go using the Charm BubbleTea framework. Connect to MySQL and PostgreSQL databases, browse tables, and run SQL queries — all from your terminal.

## Features

- **MySQL & PostgreSQL** support
- **Persistent sidebar** with table list and live search
- **Split SQL editor** — editor and results always visible side by side
- **Table viewer** with pagination and horizontal scrolling
- **Connection history** — recent connections saved and reusable

## Installation

### From source

```bash
make build
```

### Install to ~/.local/bin

```bash
make install
```

Add to PATH if needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

```bash
./otto
```

### Connecting

1. Select driver (Tab to toggle MySQL / PostgreSQL)
2. Fill in host, port, user, password, database
3. Press Enter to connect
4. Previous connections are shown on launch — select with ↑↓ and Enter

### Navigation

| Key | Action |
|-----|--------|
| `↑↓` / `j k` | Navigate sidebar or table rows |
| `Enter` | Open selected table |
| `/` | Search tables in sidebar |
| `s` | Open SQL editor |
| `Tab` | Switch focus: sidebar ↔ content panel |
| `Ctrl+R` | Switch focus: editor ↔ results (in SQL editor) |
| `Esc` | Return to sidebar |
| `Ctrl+C` | Quit |

### Table viewer

| Key | Action |
|-----|--------|
| `n` / `p` | Next / previous page |
| `←→` / `h l` | Scroll columns |
| `r` | Refresh |

### SQL editor

| Key | Action |
|-----|--------|
| `Ctrl+E` | Execute query |
| `Ctrl+R` | Switch between editor and results |

## Requirements

- Go 1.21+
- MySQL 5.7+ or PostgreSQL 10+

## License

MIT
