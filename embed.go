package main

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed Roboto-Light.ttf
var DefaultFont []byte

func getDefaultFontPath() string {
	fontPath := filepath.Join(os.TempDir(), "i376_font.ttf")
	os.WriteFile(fontPath, DefaultFont, 0777)
	return fontPath
}
