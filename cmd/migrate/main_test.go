package main

import (
    "errors"
    "os"
    "testing"
)

func TestIsFileMissingErr(t *testing.T) {
    if !isFileMissingErr(os.ErrNotExist) {
        t.Fatalf("expected true for os.ErrNotExist")
    }
    if !isFileMissingErr(errors.New("file does not exist")) {
        t.Fatalf("expected true for 'file does not exist'")
    }
    if !isFileMissingErr(errors.New("no such file or directory")) {
        t.Fatalf("expected true for 'no such file'")
    }
    if isFileMissingErr(errors.New("some other error")) {
        t.Fatalf("expected false for unrelated error")
    }
}
