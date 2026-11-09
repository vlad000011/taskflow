package handler

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestAllowedUploadExtension(t *testing.T) {
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".pdf", ".txt", ".docx"} {
		if !allowedUploadExtension(ext) {
			t.Errorf("expected %s to be allowed", ext)
		}
	}
	for _, ext := range []string{".exe", ".html", ".svg", ".zip", ""} {
		if allowedUploadExtension(ext) {
			t.Errorf("expected %s to be rejected", ext)
		}
	}
}

func TestValidUploadContent(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		data []byte
		want bool
	}{
		{"valid png signature", ".png", []byte{137, 80, 78, 71, 13, 10, 26, 10}, true},
		{"invalid png", ".png", []byte("not a png"), false},
		{"valid jpeg signature", ".jpg", []byte{0xff, 0xd8, 0xff, 0x00}, true},
		{"invalid jpeg", ".jpeg", []byte("not jpeg"), false},
		{"valid pdf", ".pdf", []byte("%PDF-1.7"), true},
		{"invalid pdf", ".pdf", []byte("hello"), false},
		{"valid text", ".txt", []byte("hello, TaskFlow\n"), true},
		{"binary text", ".txt", []byte{'a', 0, 'b'}, false},
		{"invalid utf8 text", ".txt", []byte{0xff, 0xfe}, false},
		{"unknown extension", ".exe", []byte("MZ"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validUploadContent(tt.ext, tt.data); got != tt.want {
				t.Fatalf("validUploadContent(%q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
	}
}

func TestValidDocxContainer(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range []string{"[Content_Types].xml", "word/document.xml"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write([]byte("<xml/>")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if !validUploadContent(".docx", buf.Bytes()) {
		t.Fatal("expected a DOCX ZIP with required Word entries to be accepted")
	}
	if validUploadContent(".docx", []byte("PK but not a valid archive")) {
		t.Fatal("expected malformed DOCX archive to be rejected")
	}
}
