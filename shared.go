package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// RGBColor RBG Color Type
type RGBColor struct {
	Red   int
	Green int
	Blue  int
}

// GetHex Converts a decimal number to hex representations
func getHex(num int) string {
	hex := fmt.Sprintf("%x", num)
	if len(hex) == 1 {
		hex = "0" + hex
	}
	return hex
}

func GetRandomColorInRgb() RGBColor {
	Red := rand.Intn(255)
	Green := rand.Intn(255)
	blue := rand.Intn(255)
	c := RGBColor{Red, Green, blue}
	return c
}

// GetRandomColorInHex returns a random color in HEX format
func GetRandomColorInHex() string {
	color := GetRandomColorInRgb()
	hex := "#" + getHex(color.Red) + getHex(color.Green) + getHex(color.Blue)
	return hex
}

func GetRootPath() (string, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "os error")
	}

	dd := os.Getenv("SNAP_USER_COMMON")

	if strings.HasPrefix(dd, filepath.Join(hd, "snap", "go")) || dd == "" {
		dd = filepath.Join(hd, "Images376")
		os.MkdirAll(dd, 0777)
	}

	return dd, nil
}
