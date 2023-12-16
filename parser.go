package amaro

import (
	"fmt"
	"strings"
	"unicode"
)

type arg struct {
	name  string
	value string
}

type parser struct {
	input []rune
	i     int
}

var errParsingComplete = fmt.Errorf("parsing complete")

func parse(input string) (map[string]arg, error) {
	args := make(map[string]arg)

	parser := parser{
		input: []rune(input),
		i:     0,
	}

	if len(input) == 0 {
		return args, nil
	}

	return parser.parse()
}

func (p *parser) parse() (map[string]arg, error) {
	args := make(map[string]arg)
	p.skipSpaces()

	for {
		p.skipSpaces()
		if p.i >= len(p.input) {
			break
		}

		newArg, err := p.parsePair()
		if err != nil {
			if err == errParsingComplete {
				args["*"] = arg{name: "*", value: string(p.input[p.i:])}
				break
			}
			return nil, err
		}
		args[newArg.name] = newArg
	}

	return args, nil
}

func (p *parser) parsePair() (arg, error) {
	name, err := p.parseName()
	if err != nil {
		return arg{}, err
	}

	if p.i >= len(p.input) {
		return arg{name: name}, nil
	}

	if p.input[p.i] == '=' {
		p.i++
	} else if p.input[p.i] == ' ' {
		p.skipSpaces()
	} else {
		panic("invalid argument format")
	}

	if p.i >= len(p.input) {
		return arg{name: name}, nil
	}

	if len(p.input) >= p.i+2 && p.input[p.i] == '-' && p.input[p.i+1] == '-' {
		return arg{name: name}, nil
	}

	return arg{name: name, value: p.parseValue()}, nil
}

func (p *parser) parseName() (string, error) {
	var name strings.Builder

	if p.i+2 >= len(p.input) {
		return "", fmt.Errorf("expected an argument, got %q", string(p.input[p.i:]))
	}

	if p.input[p.i] != '-' || p.input[p.i+1] != '-' {
		return "", fmt.Errorf("expected an argument, got %q", string(p.input[p.i:]))
	}

	// remove `--`
	p.i += 2

	if p.input[p.i] == ' ' {
		return "", errParsingComplete
	}

	if !unicode.IsLetter(p.input[p.i]) {
		return "", fmt.Errorf("expected a valid argument name, got %q", string(p.input[p.i-2:]))
	}

	for {
		if p.i >= len(p.input) {
			break
		}

		if p.input[p.i] == ' ' || p.input[p.i] == '=' {
			break
		}

		name.WriteRune(p.input[p.i])
		p.i++
	}

	return name.String(), nil
}

func (p *parser) parseValue() string {
	var value strings.Builder

	if p.input[p.i] == '"' || p.input[p.i] == '\'' {
		return p.parseQuotedValue(p.input[p.i])
	}

	for {
		if p.i >= len(p.input) {
			break
		}

		if p.input[p.i] == ' ' {
			break
		}

		value.WriteRune(p.input[p.i])
		p.i++
	}

	return value.String()
}

func (p *parser) parseQuotedValue(quote rune) string {
	var value strings.Builder

	p.i++
	for {
		if p.i >= len(p.input) {
			break
		}

		if p.input[p.i] == quote && p.input[p.i-1] != '\\' {
			p.i++
			break
		}

		value.WriteRune(p.input[p.i])
		p.i++
	}

	return value.String()
}

func (p *parser) skipSpaces() {
	for {
		if p.i >= len(p.input) {
			break
		}

		if p.input[p.i] != ' ' {
			break
		}

		p.i++
	}
}
