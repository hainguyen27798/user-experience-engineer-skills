package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestAskYesNoSharesReaderAcrossSequentialPrompts(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("no\nyes\n"))
	first, err := askYesNo(reader, "first: ")
	if err != nil || first {
		t.Fatalf("first answer = %t, error = %v; want no", first, err)
	}
	second, err := askYesNo(reader, "second: ")
	if err != nil || !second {
		t.Fatalf("second answer = %t, error = %v; want yes", second, err)
	}
}

func TestAskYesNoRejectsInvalidAnswer(t *testing.T) {
	_, err := askYesNo(bufio.NewReader(strings.NewReader("maybe\n")), "prompt: ")
	if err == nil || !strings.Contains(err.Error(), "y or n") {
		t.Fatalf("askYesNo() error = %v, want y-or-n validation", err)
	}
}
