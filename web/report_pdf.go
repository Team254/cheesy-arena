// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/flopp/go-findfont"
	"github.com/jung-kurt/gofpdf"
	"path/filepath"
	"strings"
	"sync"
)

const reportCoreFontFamily = "Arial"
const reportUtf8FontFamily = "ArialUTF8"

var (
	reportArialFontOnce sync.Once
	reportArialFonts    reportArialFontFiles
)

type reportArialFontFiles struct {
	regularPath string
	boldPath    string
}

type reportPdf struct {
	*gofpdf.Fpdf
	utf8Regular bool
	utf8Bold    bool
}

// newReportPdf creates a PDF configured to use installed regular and bold Arial TTFs when available. If font discovery
// or registration fails, the PDF falls back to gofpdf's built-in Arial/Helvetica behavior.
func newReportPdf() *reportPdf {
	reportArialFontOnce.Do(
		func() {
			reportArialFonts = findArialFontFiles(findfont.List())
		},
	)
	pdf := gofpdf.New("P", "mm", "Letter", "font")
	if reportArialFonts.regularPath == "" {
		return &reportPdf{Fpdf: pdf}
	}

	if !registerReportFont(pdf, reportArialFonts.regularPath, "") {
		return &reportPdf{Fpdf: gofpdf.New("P", "mm", "Letter", "font")}
	}
	if reportArialFonts.boldPath != "" && !registerReportFont(pdf, reportArialFonts.boldPath, "B") {
		// A bad bold font should not break regular UTF-8 text; rebuild the PDF with only the regular font registered.
		pdf = gofpdf.New("P", "mm", "Letter", "font")
		if !registerReportFont(pdf, reportArialFonts.regularPath, "") {
			return &reportPdf{Fpdf: gofpdf.New("P", "mm", "Letter", "font")}
		}
		return &reportPdf{Fpdf: pdf, utf8Regular: true}
	}
	return &reportPdf{Fpdf: pdf, utf8Regular: true, utf8Bold: reportArialFonts.boldPath != ""}
}

func registerReportFont(pdf *gofpdf.Fpdf, fontPath string, style string) bool {
	pdf.SetFontLocation(filepath.Dir(fontPath))
	pdf.AddUTF8Font(reportUtf8FontFamily, style, filepath.Base(fontPath))
	return pdf.Ok()
}

// SetFont preserves the existing report call sites while swapping regular and bold Arial text to installed UTF-8 fonts.
func (pdf *reportPdf) SetFont(familyStr string, styleStr string, size float64) {
	if strings.EqualFold(familyStr, reportCoreFontFamily) {
		switch strings.ToUpper(styleStr) {
		case "":
			if pdf.utf8Regular {
				pdf.Fpdf.SetFont(reportUtf8FontFamily, "", size)
				return
			}
		case "B":
			if pdf.utf8Bold {
				pdf.Fpdf.SetFont(reportUtf8FontFamily, "B", size)
				return
			}
		}
	}
	pdf.Fpdf.SetFont(familyStr, styleStr, size)
}

// findArialFontFiles picks Arial-family TTFs from the font list returned by go-findfont. Arial Unicode is preferred for
// regular text when installed because regular Arial can lack symbols such as △ and ◅.
func findArialFontFiles(fontPaths []string) reportArialFontFiles {
	var fontFiles reportArialFontFiles
	var regularPath string
	var unicodePath string
	for _, fontPath := range fontPaths {
		style, ok := arialFontStyle(fontPath)
		if !ok {
			continue
		}
		switch style {
		case "":
			if regularPath == "" {
				regularPath = fontPath
			}
		case "U":
			if unicodePath == "" {
				unicodePath = fontPath
			}
		case "B":
			if fontFiles.boldPath == "" {
				fontFiles.boldPath = fontPath
			}
		}
		if unicodePath != "" && fontFiles.boldPath != "" {
			fontFiles.regularPath = unicodePath
			return fontFiles
		}
	}
	if fontFiles.regularPath == "" {
		if unicodePath != "" {
			fontFiles.regularPath = unicodePath
		} else {
			fontFiles.regularPath = regularPath
		}
	}
	return fontFiles
}

// arialFontStyle accepts common filenames for regular, Unicode, and bold Arial only.
func arialFontStyle(fontPath string) (string, bool) {
	fontName := strings.ToLower(filepath.Base(fontPath))
	if !strings.HasSuffix(fontName, ".ttf") {
		return "", false
	}
	normalizedName := normalizeFontName(strings.TrimSuffix(fontName, filepath.Ext(fontName)))
	for _, styleName := range []string{"italic", "oblique", "narrow", "black"} {
		if strings.Contains(normalizedName, styleName) {
			return "", false
		}
	}
	switch normalizedName {
	case "arial", "arialregular", "arialmt", "arialmtregular":
		return "", true
	case "arialunicode", "arialunicoderegular", "arialunicodems", "arialunicodemsregular", "arialuni":
		return "U", true
	case "arialbold", "arialboldmt", "arialmtbold", "arialbd":
		return "B", true
	default:
		return "", false
	}
}

func normalizeFontName(fontName string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "")
	return replacer.Replace(fontName)
}
