package main

import (
	"fmt"
	"regexp"
)

type tokenType int

const (
	tokenWord tokenType = iota
	tokenWhitespace
	tokenPunctuation
	tokenEof
)

var names = map[tokenType]string{
	tokenWord:        "WORD",
	tokenWhitespace:  "SPACE",
	tokenPunctuation: "PUNCTUATION",
	tokenEof:         "EOF",
}

var wordRegexp = regexp.MustCompile("[A-za-z]+")
var whitespaceRegexp = regexp.MustCompile("[\\s]+")
var punctuationRegexp = regexp.MustCompile("[\\p{P}\\p{S}]+")

type token struct {
	value     string
	pos       int
	tokenType tokenType
}

func (tok token) String() string {
	return fmt.Sprintf("{%s '%s' %d}", names[tok.tokenType], tok.value, tok.pos)
}

type tokenStream chan token

type stateFn func(*lexer) stateFn

type lexer struct {
	start  int // The position of the last emission
	pos    int // The current position of the lexer
	input  string
	tokens tokenStream
	state  stateFn
}

func (l *lexer) next() (val string) {
	if l.pos >= len(l.input) {
		l.pos++
		return ""
	}

	val = l.input[l.pos : l.pos+1]

	l.pos++

	return
}

func (l *lexer) backup() {
	l.pos--
}

func (l *lexer) peek() (val string) {
	val = l.next()

	l.backup()

	return
}

func (l *lexer) emit(t tokenType) {
	val := l.input[l.start:l.pos]
	tok := token{val, l.start, t}
	l.tokens <- tok
	l.start = l.pos
}

func lexData(l *lexer) stateFn {
	v := l.peek()
	switch {
	case v == "":
		l.emit(tokenEof)
		return nil

	case punctuationRegexp.MatchString(v):
		return lexPunctuation

	case whitespaceRegexp.MatchString(v):
		return lexWhitespace
	}

	return lexWord
}

func lexPunctuation(l *lexer) stateFn {
	matched := punctuationRegexp.FindString(l.input[l.pos:])
	l.pos += len(matched)
	l.emit(tokenPunctuation)

	return lexData
}

func lexWhitespace(l *lexer) stateFn {
	matched := whitespaceRegexp.FindString(l.input[l.pos:])
	l.pos += len(matched)
	l.emit(tokenWhitespace)

	return lexData
}

func lexWord(l *lexer) stateFn {
	matched := wordRegexp.FindString(l.input[l.pos:])
	l.pos += len(matched)
	l.emit(tokenWord)

	return lexData
}

func main() {
	lex := &lexer{0, 0, "This   is  a test, sir...", make(tokenStream), lexData}

	go func() {
		for lex.state = lexData; lex.state != nil; {
			lex.state = lex.state(lex)
		}
	}()

	for {
		tok := <-lex.tokens
		fmt.Println(tok)
		if tok.tokenType == tokenEof {
			break
		}
	}
}
