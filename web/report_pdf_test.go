// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestFindArialFontFilesFindsRegularOnly(t *testing.T) {
	fontDir := t.TempDir()
	fontPath := filepath.Join(fontDir, "Arial.ttf")

	assert.Equal(t, reportArialFontFiles{regularPath: fontPath}, findArialFontFiles([]string{fontPath}))
}

func TestFindArialFontFiles(t *testing.T) {
	fontDir := t.TempDir()
	regularPath := filepath.Join(fontDir, "Arial.ttf")
	boldPath := filepath.Join(fontDir, "Arial Bold.ttf")

	assert.Equal(
		t,
		reportArialFontFiles{regularPath: regularPath, boldPath: boldPath},
		findArialFontFiles([]string{regularPath, boldPath}),
	)
}

func TestFindArialFontFilesAcceptsSyntheticPlatformPaths(t *testing.T) {
	tempDir := t.TempDir()
	testCases := []struct {
		name     string
		fontPath string
	}{
		{
			name:     "Windows",
			fontPath: filepath.Join(tempDir, "Windows", "Fonts", "Arial.ttf"),
		},
		{
			name:     "macOS",
			fontPath: filepath.Join(tempDir, "Library", "Fonts", "ArialMT.ttf"),
		},
		{
			name:     "Linux",
			fontPath: filepath.Join(tempDir, "usr", "share", "fonts", "truetype", "msttcorefonts", "arial.ttf"),
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				assert.Equal(
					t,
					reportArialFontFiles{regularPath: testCase.fontPath},
					findArialFontFiles([]string{testCase.fontPath}),
				)
			},
		)
	}
}

func TestFindArialFontFilesIgnoresUnsupportedStyles(t *testing.T) {
	fontDir := t.TempDir()

	assert.Equal(
		t,
		reportArialFontFiles{},
		findArialFontFiles([]string{
			filepath.Join(fontDir, "ariali.ttf"),
			filepath.Join(fontDir, "Arial Bold Italic.ttf"),
			filepath.Join(fontDir, "Arial Narrow.ttf"),
		}),
	)
}

func TestFindArialFontFilesPrefersArialUnicodeForRegularText(t *testing.T) {
	fontDir := t.TempDir()
	arialUnicodePath := filepath.Join(fontDir, "Arial Unicode MS.ttf")
	arialPath := filepath.Join(fontDir, "Arial.ttf")

	assert.Equal(
		t,
		reportArialFontFiles{regularPath: arialUnicodePath},
		findArialFontFiles([]string{arialPath, arialUnicodePath}),
	)
}

func TestFindArialFontFilesUsesArialAfterRejectedFonts(t *testing.T) {
	fontDir := t.TempDir()
	arialItalicPath := filepath.Join(fontDir, "Arial Italic.ttf")
	arialPath := filepath.Join(fontDir, "ArialMT.ttf")
	boldPath := filepath.Join(fontDir, "arialbd.ttf")

	assert.Equal(
		t,
		reportArialFontFiles{regularPath: arialPath, boldPath: boldPath},
		findArialFontFiles([]string{arialItalicPath, arialPath, boldPath}),
	)
}

func TestFindArialFontFilesMissingCandidatesUsesCoreFallback(t *testing.T) {
	assert.Equal(t, reportArialFontFiles{}, findArialFontFiles([]string{}))
}
