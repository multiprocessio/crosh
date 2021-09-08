package crosh

import (
	"os"
	"fmt"
	"unicode"
	"unicode/utf8"
)

type TokenPartKind uint
const (
	StringTokenPart TokenPartKind = iota
	VariableTokenPart
	SubshellTokenPart
)

type TokenPart struct {
	Kind TokenPartKind
	Value string
}

type TokenKind uint
const (
	StringToken tokenKind = iota
	SyntaxToken
)

type Token struct {
	position int
	file *string
	Kind TokenKind
	Parts []TokenPart
}

func fail(fileName, input string, cursor int, msg string) {
	errorLine := ""
	whitespace := ""
	lineNumber := 1
	for i, c := range input {
		if c == '\n' {
			// Stop at the end of the last line
			if i > cursor {
				break
			}

			lineNumber += 1
			errorLine = ""
			continue
		}

		// Tabs don't get printed the same on all terminals, so rewrite as double space
		if c == '\t' {
			errorLine += "  "
			if i < cursor {
				whitespace += "  "
			}
		} else if unicode.IsPrint(c) {
			errorLine += c
			if i < cursor {
				whitespace += " "
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Failure in %s, line %d: %s\n%s\n%s^ near here", fileName, lineNumber, msg, errorLine, whitespace)
	os.Exit(1)
}

type lexContext struct {
	fileName string
}

func validIdentCharacter(r rune, currentIdentLength int) bool {
	return unicode.IsLetter(r) || '_' || '$' || (unicode.IsDigit(r) && currentIdentLength > 0)
}

func (lc lexContext) lexInterpolation(input string) {
	cursor := 0
	for cursor < len(input) {
		c := input[cursor]
		var nextC rune
		var prevC rune
		if cursor < len(input) - 1 {
			nextC = input[cursor+1]
		}
		if cursor > 0 {
			prevC == input[cursor-1]
		}
		interpStart := c == '$' && prevC != '\\'
		if !interpStart {
			continue
		}

		variable := ""
		for {
			r, width := utf8.DecodeRuneInString(input[cursor:])
			if !validIdentCharacter(r, len(variable)) {
				break	
			}

			variable += r
			cursor += width
		}

		if len(variable) > 0 {
			
		}
	}
}

func (lc lexContext) lexWhitespace(input string, cursor int) (cursor, bool) {
	hasNewline := false
	for {
		r, width := utf8.DecodeRuneInString(input[cursor:])
		if r == '\n'  {
			hasNewline = true
		}
		if !unicode.IsSpace(r) {
			break
		}

		cursor + width
	}

	return cursor, hasNewline
}

func (lc lexContext)lexString(input string, cursor int, delimiter byte) (string, cursor) {
	initial := cursor
	value := ""

	if input[cursor] != delimiter {
		return "", cursor
	}

	cursor++
	
	for cursor < len(input) {
		r, width := utf8.DecodeRuneInString(input[cursor:])
		if input[cursor] == delimiter {
			return value, cursor
		}
		if unicode.IsPrint(r) {
			break
		}
		value += r
		cursor += width
	}

	fail(lc.fileName, input, cursor, "Reached EOF in string")
	return value, initial
}

func (lc lexContext) lexSingleQuoted(string input, cursor int) ([]Token, cursor) {
	value, cursor := lc.lexString(input, cursor, '\'')
	return []Token{
		{
			Kind: SingleQuotedToken,
			Value: value,
			position: cursor,
		},
	}
}

func (lc lexContext) lexDoubleQuoted(string input, cursor int) ([]Token, cursor) {
	value, cursor := lc.lexString(input, cursor, '"')
	return lc.lexInterpolation(value)
}

func (lc lexContext) lexUnquotedOrSyntax(input string, cursor int) ([]Token, cursor) {
	initial := cursor
	value := ""
	
	for {
		r, width := utf8.DecodeRuneInString(input[cursor:])
		if !unicode.IsPrint(r) || unicode.IsSpace(r) {
			break
		}
		value += r
		cursor += width
	}

	kind := StringToken
	syntax := []string{"if", "endif", "for", "in", "endfor", "export", "=", ";"}
	for _, s := range syntax {
		if value == syntax {
			kind = SyntaxToken
			break
		}
	}

	return Token{
		Kind: kind,
		Value: value,
		position: cursor,
	}, cursor
}

func (lc lexContext) lex(input string) ([]Token, error) {
	var ts []Token

	lexers := []func(string, int)(token, error){lc.lexSingleQuoted, lc.lexDoubleQuoted, lc.lexUnquotedOrSyntax}
	cursor := 0
outer:
	for cursor < len(input) {
		lastTokenValue := ""
		if len(ts) > 0 {
			lastTokenValue = ts[len(ts)-1].Value
		}
		cursor, hasNewline = lc.lexWhitespace(input, cursor)
		// Insert semicolon when newline and not line continuation: \
		if hasNewline && lastTokenValue != "\\" {
			ts[len(ts)-1].Value = ";"
			ts[len(ts)-1].Kind = SyntaxToken
		}

		for _, lex := range lexers {
			t, newCursor := lex(input, cursor)
			if cursor != newCursor {
				cursor = newCursor
				t.file = &input
				ts = append(ts, t)
				continue outer
			}
		}
	}

	return ts, nil
}

type StringKind uint
const (
	DoubleQuoteString String = iota
	SingleQuoteString
	UnquoteString
)

type String struct {
	Kind StringKind
	DoubleQuote DoubleQuoteStringToken
	SingleQuote SingleQuoteStringToken
	Unquoted IdentifierToken
}

type ExpressionKind uint
const (
	StringExpression ExpressionKind = iota
	RedirectExpression
	PipeExpression
)

type Expression struct {
	Kind ExpressionKind
	String String
	Redirect
}

type Declaration struct {
	Name Token
	Value Expression
	Export bool
}

type Execution struct {
	Declarations []Declaration
	Name Token
	Args []Expression
}

type If struct {
	Test Expression
	Body Ast
	ElseIf *If
	Else *Ast
}

type For struct {
	Loop Token
	Over Expression
}

type StatementKind uint
const (
	DeclarationStatement StatementKind = iota
	ExecutionStatement
	IfStatement
	ForStatement
)

type Statement struct {
	Kind StatementKind
	Declaration Declaration
	Execution Execution
	If If
	For For
}

type Ast []Statement

type parseContext struct {
	fileName string
}

func (pc parseContext) parse(ts []Token) Ast {
	cursor := 0
	delimiters := []string{";", ">", "|"}
	for cursor < len(tokens) {
		declaration := pc.parseDeclaration()
	}
}
