package stream

import (
	"testing"
)

var (
	testInput1 = "test {{ abc }} %.2f"
	testInput2 = "%.2f"
	testInput3 = "12345678"
)

func TestPositionMethods(t *testing.T) {
	t.Parallel()
	l := NewConfig(testInput1)
	if l == nil {
		t.Fatal("output lexer should not be nil")
	}
	r := l.next()
	if r != 't' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "t", string(r))
	}
	if l.start != 0 {
		t.Errorf("unexpected start value: wanted %d ; got %d", 0, l.start)
	}
	if l.width != 1 {
		t.Errorf("unexpected width value: wanted %d ; got %d", 1, l.width)
	}
	if l.pos != 1 {
		t.Errorf("unexpected pos value: wanted %d ; got %d", 1, l.pos)
	}

	l.ignore()
	if l.start != 1 {
		t.Errorf("unexpected start value: wanted %d ; got %d", 1, l.start)
	}

	l.next()
	l.next()
	l.backup()
	if l.pos != 2 {
		t.Errorf("unexpected pos value: wanted %d ; got %d", 2, l.pos)
	}

	r = l.peek()
	if r != 's' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "s", string(r))
	}
	if l.width == 2 {
		t.Errorf("unexpected width value: wanted %d ; got %d", 2, l.width)
	}
}

func TestAccept(t *testing.T) {
	t.Parallel()
	l := NewConfig(testInput2)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	res := l.accept("%")
	if !res {
		t.Errorf("rune was not accepted")
	}
	res = l.accept("?")
	if res {
		t.Errorf("rune should not be accepted")
	}
}

func TestAcceptRun(t *testing.T) {
	t.Parallel()
	l := NewConfig(testInput3)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	res := l.acceptRun("12345678")
	if !res {
		t.Errorf("rune was not accepted")
	}
	res = l.acceptRun("abcdefgh")
	if res {
		t.Errorf("rune should not be accepted")
	}
}
