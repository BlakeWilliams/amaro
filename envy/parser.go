package envy

import (
	"fmt"
	"strings"
	"unicode"
)

type parser struct {
	mapping map[string]string
	content []rune
	error   error
	offset  int
}

func (p *parser) parse() error {
	for p.offset < len(p.content) {
		p.skipWhitespace()
		if p.offset >= len(p.content) {
			return nil
		}

		switch p.content[p.offset] {
		case '#':
			p.parseComment()
		default:
			if err := p.parseAssignment(); err != nil {
				return err
			}

			// Skip to the end of this line
			for p.offset < len(p.content) && p.content[p.offset] != '\n' {
				p.offset++
			}
		}
	}

	return nil
}

func (p *parser) parseComment() {
	for p.offset < len(p.content) && p.content[p.offset] != '\n' {
		p.offset++
	}
}

func (p *parser) parseAssignment() error {
	start := p.offset

	key, err := p.parseKey()
	if err != nil {
		return err
	}
	p.skipWhitespace()
	if p.content[p.offset] != '=' {
		return fmt.Errorf("missing = after key: %s", key)
	}
	p.offset++
	value, err := p.parseValue()
	if err != nil {
		return err
	}
	p.mapping[key] = value

	for p.offset < len(p.content) && p.content[p.offset] != '\n' {
		switch p.content[p.offset] {
		case '#':
			p.parseComment()
			return nil
		case ' ':
			p.offset++
			continue
		default:
			return fmt.Errorf("unexpected character after value: %s", string(p.content[start:p.offset]))
		}
	}

	return nil
}

func (p *parser) parseKey() (string, error) {
	start := p.offset
	for p.offset < len(p.content) && p.content[p.offset] != '=' && p.content[p.offset] != ' ' && p.content[p.offset] != '\n' {
		p.offset++
	}
	if p.offset == start {
		return "", fmt.Errorf("missing key")
	}

	strKey := strings.TrimSpace(string(p.content[start:p.offset]))
	if !validKey.MatchString(strKey) {
		return "", fmt.Errorf("invalid key, keys must be uppercase, start with a letter, and may only contain _'s: %s", strKey)
	}

	return strKey, nil
}

func (p *parser) parseValue() (string, error) {
	start := p.offset
	if p.content[p.offset] == '"' || p.content[p.offset] == '\'' || p.content[p.offset] == '`' {
		quote := p.content[p.offset]
		start = p.offset + 1

		for p.offset < len(p.content) {
			p.offset++
			if p.content[p.offset] == quote && p.content[p.offset-1] != '\\' {
				break
			}
		}
		p.offset += 1 // remove the closing quote

		val := string(p.content[start : p.offset-1])
		if quote == '"' {
			val = strings.ReplaceAll(val, "\\n", "\n")
			val = strings.ReplaceAll(val, "\\\"", "\"")
		}

		return val, nil
	}

	for p.offset < len(p.content) && (p.content[p.offset] != '\n' && p.content[p.offset] != '#') {
		p.offset++
	}

	strVal := string(p.content[start:p.offset])

	return strings.TrimSpace(strVal), nil
}

func (p *parser) skipWhitespace() {
	for p.offset < len(p.content) && unicode.IsSpace(p.content[p.offset]) {
		p.offset++
	}
}
