# DB Console

A modern terminal-based database console written in Go using the Charm BubbleTea framework. Connect to MySQL and PostgreSQL databases, browse tables, run queries, and edit data directly from your terminal.

## Features

- **Multiple Database Support**: MySQL and PostgreSQL
- **TUI Interface**: Beautiful terminal user interface with BubbleTea
- **Table Browsing**: View and navigate database tables
- **Query Editor**: Write and execute SQL queries
- **Query History**: Track and reuse previous queries

## Installation

### From Source

```bash
make build
```

### Install to ~/.local/bin

```bash
make install
```

Make sure to add `~/.local/bin` to your PATH:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

Run the application:

```bash
./otto
```

### Connecting to a Database

1. Select your database type (MySQL or PostgreSQL)
2. Enter connection details:
   - Host (default: localhost)
   - Port (default: 3306 for MySQL, 5432 for PostgreSQL)
   - Username
   - Password
   - Database name

### Navigation

- Use arrow keys to navigate
- Tab to switch between panels
- Enter to execute actions
- Esc to go back

## Requirements

- Go 1.25+
- MySQL 5.7+ or PostgreSQL 10+

## License

MIT

---

This README was created with opencode zen
