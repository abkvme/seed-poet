#!/usr/bin/env python3
"""Convert BIP39 wordlist to Go code."""

def main():
    # Read the wordlist (one word per line)
    with open('../data/english.txt', 'r', encoding='utf-8') as f:
        words = [line.strip() for line in f if line.strip()]

    # Generate Go code
    go_code = '''package wordlist

// BIP39Words contains all 2048 BIP39 English mnemonic words.
var BIP39Words = []string{
'''

    # Add words in groups of 8 per line for readability
    for i in range(0, len(words), 8):
        chunk = words[i:i+8]
        line = '\t' + ', '.join(f'"{w}"' for w in chunk) + ','
        go_code += line + '\n'

    go_code += '}\n'

    # Write Go file
    with open('../internal/wordlist/words.go', 'w') as f:
        f.write(go_code)

    print(f"Generated words.go with {len(words)} words")

if __name__ == '__main__':
    main()
