# Seed Poet

Random BIP39 mnemonic generator that outputs words as rhyming phonetic couplets.

Seed Poet is a Go application that generates new BIP39-compliant mnemonic seeds formatted as memorable, rhyming poems. It uses a structured VERB + ADJECTIVE + NOUN format and a phonetic rhyming engine to ensure the generated seeds have a rhythmic flow that is easier to remember than a standard list of random words.

## Features

- **Memory-Focused**: Generates rhyming couplets that are easier to memorize than standard random word lists.
- **BIP39 Valid**: Read in order, the poem is a valid BIP39 mnemonic, checksum included, so any BIP39 wallet accepts it.
- **CLI Mode**: Generate 12 or 24-word seeds directly from your terminal.
- **Web Mode**: Run a lightweight web server to generate seeds via a browser.
- **Phonetic Engine**: Matches syllable counts and phonetic endings for consistent meter and rhyme.

## Getting Started

### Prerequisites

- Go 1.26 or higher
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

## Security

- **Randomness**: every poem is drawn from a ChaCha8 generator seeded from `crypto/rand` (the OS random source). Nothing is predictable from earlier poems.
- **Checksum**: the last word of a BIP39 phrase carries a checksum. Seed Poet draws random poems and keeps only one whose checksum is correct, so the choice stays uniform among valid poems.
- **Strength**: the rhyme and grammar rules limit which words can appear where, so a poem holds less randomness than a free BIP39 phrase of the same length. Our estimate: about 88 bits for 12 words (a free 12-word phrase has 128) and about 176 bits for 24 words (free: 256). For a wallet that holds real value, use 24 words.
- **Privacy**: poems are generated in memory and never logged or stored. Pages are served with `Cache-Control: no-store` and `Referrer-Policy: no-referrer`. For a real wallet, run the binary offline on your own machine rather than trusting any website.

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
- `internal/bip39`: BIP39 wordlist index and checksum validation.
- `internal/rhyme`: Phonetic rhyming engine and index.
- `internal/server`: Web server and templates.
- `internal/wordlist`: BIP39 word categories and lists.
- `scripts/`: Python tools for data preparation.
- `data/`: Source BIP39 wordlist.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
