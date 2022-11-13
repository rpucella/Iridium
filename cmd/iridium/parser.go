package main

// From https://blog.gopheracademy.com/advent-2014/parsers-lexers/

import (
	"bufio"
	"io"
	"bytes"
	"fmt"
)


/*
   A passage is an array of blocks and a set of options
   Each block is an array of strings
*/

type BlockKind int
type TextKind int

const (
	TEXT BlockKind = iota
	IMAGE
)

const (
	TEXT_WORD TextKind = iota
	TEXT_QUOTE
	TEXT_EMPH
	TEXT_STRONG
)

type Text struct {
	Kind TextKind
	Word string
	Content []Text
}

type Block struct {
	Kind BlockKind
	Content []Text
	Image string
	Style string
}

type Passage struct {
	Blocks []Block
	Options []Option
}

type Option struct {
	Target string
	Content []Text
}


// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS
	NL
	
	// Literals
	WORD
	STRING

	QUOTE   // "
	
	ANNOTATION  // #(
	OPEN    // (
	CLOSE   // )

)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

var eof = rune(0)


// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
	parenCount int
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r), parenCount: 0}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) incr() {
	s.parenCount += 1
}

func (s *Scanner) decr() {
	if (s.parenCount > 0) { 
		s.parenCount -= 1
	}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// We're in an SExpression?
	if (s.parenCount > 0) {
		return s.ScanDirective()
	}
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if ch == eof {
		return EOF, ""
	} else if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if ch == '#' {
		return s.scanHash()
	} else if ch == '"' {
		return QUOTE, ""
	}
	s.unread()
	return s.scanWord()
}

// ScanDirective returns the next token and literal value in the context of a directive
func (s *Scanner) ScanDirective() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if ch == eof {
		return EOF, ""
	} else if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if ch == '(' {
		s.incr()
		return OPEN, ""
	} else if ch == ')' {
		s.decr()
		return CLOSE, ""
	} else if ch == '"' {
		return s.scanString()
	}
	s.unread()
	return s.scanWordInDirective()
}

func (s *Scanner) scanHash() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()
	if (ch == '#') {
		// Forget the first # and treat as current word
		s.unread()
		return s.scanWord()
	}
	if (ch != '(') {
		return ILLEGAL, ""
	}
	s.incr()
	return ANNOTATION, ""
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	numNL := 0
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			if ch == '\n' {
				numNL += 1
			}
			buf.WriteRune(ch)
		}
	}
	// count two NLs as a "paragraph"
	if numNL > 1 {
		return NL, buf.String()
	}
	return WS, buf.String()
}

func (s *Scanner) scanWord() (tok Token, lit string) {
	// Create a buffer.
	var buf bytes.Buffer
	
	// Read every subsequent character into the buffer until a
	// special character
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if isWhitespace(ch) {
			s.unread()
			break
		} else if ch == '"' {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return WORD, buf.String()
}

func (s *Scanner) scanWordInDirective() (tok Token, lit string) {
	// Create a buffer.
	var buf bytes.Buffer
	
	// Read every subsequent character into the buffer until a
	// special character
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if isWhitespace(ch) || ch == ')' || ch == '(' {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return WORD, buf.String()
}

func (s *Scanner) scanString() (tok Token, lit string) {
	// Create a buffer.
	var buf bytes.Buffer
	
	// Read every subsequent character into the buffer until a "
	// TODO: Allow escaped "
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if ch == '"' {
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return STRING, buf.String()
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	///fmt.Println("In scan()", tok, lit)
	return
}

// scanDirective returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scanDirective() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.ScanDirective()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	///fmt.Println("In scanDirective()", tok, lit)
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// scanDirectiveIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanDirectiveIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scanDirective()
	if tok == WS {
		tok, lit = p.scanDirective()
	}
	return
}

func (p *Parser) Parse() (*Passage, error) {
	// There is probably a nicer way to write this, possibly recursively.
	passage := &Passage{make([]Block, 0, 10), make([]Option, 0, 10)}
	inQuote := false
	var savedText []Text
	blockText := make([]Text, 0, 10)
	for {
		tok, lit := p.scanIgnoreWhitespace()

		if tok == ILLEGAL {
			return nil, fmt.Errorf("Illegal lexeme")
		}
		if tok == WORD {
			blockText = append(blockText, Text{TEXT_WORD, lit, nil})
		}
		if tok == NL {
			if len(blockText) > 0 { 
				passage.Blocks = append(passage.Blocks, Block{TEXT, blockText, "", ""})
				blockText = make([]Text, 0, 10)
			}
		}
		if tok == QUOTE {
			if inQuote {
				inQuote = false
				savedText = append(savedText, Text{TEXT_QUOTE, "", blockText})
				blockText = savedText
			} else {
				inQuote = true
				savedText = blockText
				blockText = make([]Text, 0, 10)
			}
		}
		if tok == EOF {
			if inQuote {
				inQuote = false
				savedText = append(savedText, Text{TEXT_QUOTE, "", blockText})
				blockText = savedText
			}				
			if len(blockText) > 0 { 
				passage.Blocks = append(passage.Blocks, Block{TEXT, blockText, "", ""})
			}
			return passage, nil
		}
		if tok == ANNOTATION {
			if inQuote {
				inQuote = false
				savedText = append(savedText, Text{TEXT_QUOTE, "", blockText})
				blockText = savedText
			}				
			if len(blockText) > 0 { 
				passage.Blocks = append(passage.Blocks, Block{TEXT, blockText, "", ""})
				blockText = make([]Text, 0, 10)
			}
			sexp, err := p.parseSExpressions()
			if err != nil {
				return nil, err
			}
			if sexp.index(0).isSymbol() && sexp.index(0).value == "option" {
				if !sexp.index(1).isString() {
					return nil, fmt.Errorf("No name supplied with option")
				}
				target := sexp.index(1).value
				if sexp.index(2) != nil  {
					return nil, fmt.Errorf("Extra junk after option name")
				}
				inQuote := false
				var savedText []Text
				text := make([]Text, 0, 10)
				for {
					tok, lit = p.scanIgnoreWhitespace()
					if tok == WORD {
						text = append(text, Text{TEXT_WORD, lit, nil})
					} else if tok == QUOTE {
						if inQuote {
							inQuote = false
							savedText = append(savedText, Text{TEXT_QUOTE, "", text})
							text = savedText
						} else {
							inQuote = true
							savedText = text
							text = make([]Text, 0, 10)
						}
					} else if tok == ANNOTATION {
						if inQuote {
							inQuote = false
							savedText = append(savedText, Text{TEXT_QUOTE, "", text})
							text = savedText
						}				
						sexp, err := p.parseSExpressions()
						if err != nil {
							return nil, err
						}
						if sexp.index(0).isSymbol() && sexp.index(0).value == "end" {
							if sexp.index(1) != nil {
								return nil, fmt.Errorf("Extra junk after end")
							}
							break
						} else {
							return nil, fmt.Errorf("Illegal token in option text")
						}
					} else {
						return nil, fmt.Errorf("Illegal token in option text")
					}
				}
				passage.Options = append(passage.Options, Option{target, text})
			} else if sexp.index(0).isSymbol() && sexp.index(0).value == "image" {
				if !sexp.index(1).isString() {
					return nil, fmt.Errorf("No image name supplied with image")
				}
				target := sexp.index(1).value
				if sexp.index(2) != nil  {
					return nil, fmt.Errorf("Extra junk after image name")
				}
				if len(blockText) > 0 { 
					passage.Blocks = append(passage.Blocks, Block{TEXT, blockText, "", ""})
					blockText = make([]Text, 0, 10)
				}
				passage.Blocks = append(passage.Blocks, Block{IMAGE, nil, target, ""})
			}
		}
	}
}

func (p *Parser) parseSExpressions() (*SExp, error) {
	sNil := newNil()
	result := sNil
	var curr *SExp
	for {
		tok, lit := p.scanIgnoreWhitespace()
		//fmt.Printf("temp: %s\n", result)
		//fmt.Printf("temp: %s\n", result.str())
		//fmt.Printf("Token = %d %s\n", tok, lit)
		var car *SExp
		if tok == OPEN {
			sexp, err := p.parseSExpressions()
			if err != nil {
				return nil, err
			}
			car = sexp
			//fmt.Printf("sub-sexp: %s\n", car)
			//fmt.Printf("sub-sexp: %s\n", car.str())
		} else if tok == STRING {
			car = newString(lit)
		} else if tok == WORD {
			car = newSymbol(lit)
		} else if tok == CLOSE {
			//fmt.Printf("result: %s\n", result)
			//fmt.Printf("result: %s\n", result.str())
			return result, nil
		} else {
			return nil, fmt.Errorf("Illegal token in s-expression: %d %s", tok, lit)
		}
		new_node := newCons(car, sNil)
		if curr == nil {
			result = new_node
		} else {
			curr.cdr = new_node
		}
		curr = new_node
	}
}
