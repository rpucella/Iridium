package main

import (
	"strings"
)

type SExpType int

const (
	T_INT SExpType = iota
	T_STRING
	T_SYMBOL
	T_CONS
	T_NIL
)

type SExp struct {
	kind SExpType
	value string
	car *SExp
	cdr *SExp
}

func (s *SExp) str() (string) {
	if s == nil {
		return "<nil>"
	}
	if s.kind == T_NIL {
		return "()"
	}
	if s.kind == T_STRING {
		return "\"" + s.value + "\""
	}
	if s.kind == T_SYMBOL {
		return s.value
	}
	if s.kind == T_CONS {
		result := "("
		curr := s
		for {
			if curr.kind == T_NIL {
				return result + ")"
			} 
			result += " " + curr.car.str()
			curr = curr.cdr
		}
	}
	return "??"
}

func newNil() (*SExp) {
	return &SExp{kind: T_NIL}
}

func newString(s string) (*SExp) {
	return &SExp{kind: T_STRING, value: s}
}

func newSymbol(s string) (*SExp) {
	return &SExp{kind: T_SYMBOL, value: strings.ToLower(s)}
}

func newCons(car *SExp, cdr *SExp) (*SExp) {
	return &SExp{kind: T_CONS, car: car, cdr: cdr}
}

func (s *SExp) isString() (bool) {
	return s != nil && s.kind == T_STRING
}

func (s *SExp) isSymbol() (bool) {
	return s != nil && s.kind == T_SYMBOL
}

func (s *SExp) isCons() (bool) {
	return s != nil && s.kind == T_CONS
}

func (s *SExp) isNil() (bool) {
	return s != nil && s.kind == T_NIL
}

func (s *SExp) length() (int) {
	if s == nil {
		return 0
	}
	l := 0
	curr := s
	for {
		if curr.isCons() {
			l += 1
			curr = curr.cdr
		} else {
			return l
		}
	}
}

func (s *SExp) index(i int) (*SExp) {
	if i < 0 {
		return nil
	}
	curr := s
	count := i
	for {
		if !curr.isCons() {
			return nil
		} else if count == 0 {
			return curr.car
		} else {
			count -= 1
			curr = curr.cdr
		}
	}
}
		
