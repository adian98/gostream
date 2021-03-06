package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/itsubaki/gostream/pkg/clause"
	"github.com/itsubaki/gostream/pkg/lexer"
	"github.com/itsubaki/gostream/pkg/statement"
)

type Registry map[string]interface{}

type Parser struct {
	Registry Registry
}

func New() *Parser {
	return &Parser{make(map[string]interface{})}
}

func (p *Parser) Register(name string, t interface{}) {
	p.Registry[name] = t
}

func (p *Parser) ParseFunction(s *statement.Statement, l *lexer.Lexer) error {
	for {
		token, literal := l.Tokenize()
		switch token {
		case lexer.EOF:
			return fmt.Errorf("invalid token=%s", literal)
		case lexer.FROM:
			return nil
		case lexer.ASTERISK:
			s.SetFunction(clause.SelectAll{})
		case lexer.COUNT:
			s.SetFunction(clause.Count{As: "count(*)"})
		case lexer.MAX:
			_, name := l.TokenizeIdentifier()
			if IntField(s.EventType, name) {
				s.SetFunction(clause.MaxInt{Name: name, As: fmt.Sprintf("max(%s)", name)})
			}
			if FloatField(s.EventType, name) {
				s.SetFunction(clause.MaxFloat{Name: name, As: fmt.Sprintf("max(%s)", name)})
			}
		case lexer.MIN:
			_, name := l.TokenizeIdentifier()
			if IntField(s.EventType, name) {
				s.SetFunction(clause.MinInt{Name: name, As: fmt.Sprintf("min(%s)", name)})
			}
			if FloatField(s.EventType, name) {
				s.SetFunction(clause.MinFloat{Name: name, As: fmt.Sprintf("min(%s)", name)})
			}
		case lexer.MED:
			_, name := l.TokenizeIdentifier()
			if IntField(s.EventType, name) {
				s.SetFunction(clause.MedianInt{Name: name, As: fmt.Sprintf("med(%s)", name)})
			}
			if FloatField(s.EventType, name) {
				s.SetFunction(clause.MedianFloat{Name: name, As: fmt.Sprintf("med(%s)", name)})
			}
		case lexer.SUM:
			_, name := l.TokenizeIdentifier()
			if IntField(s.EventType, name) {
				s.SetFunction(clause.SumInt{Name: name, As: fmt.Sprintf("sum(%s)", name)})
			}
			if FloatField(s.EventType, name) {
				s.SetFunction(clause.SumFloat{Name: name, As: fmt.Sprintf("sum(%s)", name)})
			}
		case lexer.AVG:
			_, name := l.TokenizeIdentifier()
			if IntField(s.EventType, name) {
				s.SetFunction(clause.AverageInt{Name: name, As: fmt.Sprintf("avg(%s)", name)})
			}
			if FloatField(s.EventType, name) {
				s.SetFunction(clause.AverageFloat{Name: name, As: fmt.Sprintf("avg(%s)", name)})
			}
		}
	}
}

func (p *Parser) ParseEventType(s *statement.Statement, l *lexer.Lexer) error {
	for {
		if token, _ := l.Tokenize(); token == lexer.FROM {
			break
		}
	}

	for {
		token, literal := l.Tokenize()
		switch token {
		case lexer.EOF:
			return fmt.Errorf("invalid token=%s", literal)
		case lexer.DOT:
			return nil
		case lexer.IDENTIFIER:
			v, ok := p.Registry[literal]
			if !ok {
				return fmt.Errorf("EventType [%s] is not registered", literal)
			}

			s.SetEventType(v)
		}
	}
}

func (p *Parser) ParseWindow(s *statement.Statement, l *lexer.Lexer) error {
	for {
		if token, _ := l.Tokenize(); token == lexer.DOT {
			break
		}
	}

	token, literal := l.Tokenize()
	if token == lexer.EOF {
		return fmt.Errorf("invalid token=%s", literal)
	}

	if token == lexer.LENGTH {
		s.SetWindow(token)

		_, lex := l.TokenizeIdentifier()
		length, err := strconv.Atoi(lex)
		if err != nil {
			return fmt.Errorf("atoi=%s: %v", lex, err)
		}

		s.SetLength(length)
		return nil
	}

	if token == lexer.TIME {
		s.SetWindow(token)

		_, lex := l.TokenizeIdentifier()
		ct, err := strconv.Atoi(lex)
		if err != nil {
			return fmt.Errorf("atoi=%s: %v", lex, err)
		}

		t, _ := l.TokenizeIgnoreWhiteSpace()
		switch t {
		case lexer.SEC:
			s.SetTime(time.Duration(ct) * time.Second)
		case lexer.MIN:
			s.SetTime(time.Duration(ct) * time.Minute)
		}

		return nil
	}

	return fmt.Errorf("invalid token=%s", literal)
}

func (p *Parser) ParseWhere(s *statement.Statement, l *lexer.Lexer) error {
	for {
		if token, _ := l.Tokenize(); token == lexer.DOT {
			break
		}
	}

	list := make([]clause.Where, 0)
	for {
		token, _ := l.Tokenize()
		if token == lexer.EOF {
			break
		}

		if token != lexer.WHERE && token != lexer.AND && token != lexer.OR {
			continue
		}

		_, name := l.TokenizeIdentifier()
		sel, _ := l.TokenizeIgnoreIdentifier()
		_, value := l.TokenizeIdentifier()

		if IntField(s.EventType, name) {
			val, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("atoi=%s", value)
			}

			switch sel {
			case lexer.LARGER:
				list = append(list, clause.LargerThanInt{Name: name, Value: val})
			case lexer.LESS:
				list = append(list, clause.LessThanInt{Name: name, Value: val})
			}
		}

		if FloatField(s.EventType, name) {
			_, value2 := l.TokenizeIdentifier()
			fvalue := fmt.Sprintf("%s.%s", value, value2)

			val, err := strconv.ParseFloat(fvalue, 64)
			if err != nil {
				return fmt.Errorf("parse float=%s", fvalue)
			}

			switch sel {
			case lexer.LARGER:
				list = append(list, clause.LargerThanFloat{Name: name, Value: val})
			case lexer.LESS:
				list = append(list, clause.LessThanFloat{Name: name, Value: val})
			}
		}
	}

	s.SetWhere(list...)
	return nil
}

func (p *Parser) Parse(query string) (*statement.Statement, error) {
	s := statement.New()

	if token, literal := lexer.New(strings.NewReader(query)).Tokenize(); token != lexer.SELECT {
		return nil, fmt.Errorf("invalid token=%s", literal)
	}

	if err := p.ParseEventType(s, lexer.New(strings.NewReader(query))); err != nil {
		return nil, fmt.Errorf("parse event type: %v", err)
	}

	if err := p.ParseFunction(s, lexer.New(strings.NewReader(query))); err != nil {
		return nil, fmt.Errorf("parse function: %v", err)
	}

	if err := p.ParseWindow(s, lexer.New(strings.NewReader(query))); err != nil {
		return nil, fmt.Errorf("parse window: %v", err)
	}

	if err := p.ParseWhere(s, lexer.New(strings.NewReader(query))); err != nil {
		return nil, fmt.Errorf("parse selector: %v", err)
	}

	return s, nil
}
