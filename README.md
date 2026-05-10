# Seed Poet

Random BIP39 mnemonic generator that outputs words as rhyming phonetic couplets.

Seed Poet is a Go application that generates new BIP39-compliant mnemonic seeds formatted as memorable, rhyming poems. It uses a structured VERB + ADJECTIVE + NOUN format and a phonetic rhyming engine to ensure the generated seeds have a rhythmic flow that is easier to remember than a standard list of random words.

## Features

- **Memory-Focused**: Generates rhyming couplets that are easier to memorize than standard random word lists.
- **BIP39 Compliant**: Every word in the generated poem is a valid BIP39 English mnemonic word.
- **CLI Mode**: Generate 12 or 24-word seeds directly from your terminal.
- **Web Mode**: Run a lightweight web server to generate seeds via a browser.
- **Phonetic Engine**: Matches syllable counts and phonetic endings for consistent meter and rhyme.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Python 3 (for data preparation scripts)

### Installation

```bash
git clone https://github.com/abkvme/seed-poet.git
cd seed-poet
go build -o seedpoet ./cmd/bip39poem
```

### Usage

#### CLI

Generate a new 12-word seed:
```bash
./seedpoet 12
```

Generate a new 24-word seed:
```bash
./seedpoet 24
```

#### Web Server

Start the server on default port 8080:
```bash
./seedpoet serve
```

Access it at `http://localhost:8080`.

## Development

The project includes Python scripts for processing the BIP39 wordlist and generating phonetic rhyme groups.

### Data Preparation

If you modify the word categories or rhyme groups in the Python scripts:

```bash
make generate
```

This will run the scripts in `scripts/` and update the Go source files in `internal/wordlist/` and `internal/rhyme/`.

### Makefile Commands

- `make build`: Build the binary.
- `make run`: Build and run the web server.
- `make clean`: Remove built binaries.
- `make generate`: Run Python scripts to update internal data.

## Project Structure

- `cmd/bip39poem`: Application entry point.
- `internal/generator`: Core poem generation logic.
- `internal/rhyme`: Phonetic rhyming engine and index.
- `internal/server`: Web server and templates.
- `internal/wordlist`: BIP39 word categories and lists.
- `scripts/`: Python tools for data preparation.
- `data/`: Source BIP39 wordlist.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
