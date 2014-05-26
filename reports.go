// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for generating CSV and PDF reports.

package main

import (
	"code.google.com/p/gofpdf"
	"fmt"
	"net/http"
	"strconv"
	"text/template"
)

func TeamsCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	template, err := template.ParseFiles("templates/teams.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.Execute(w, teams)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func TeamsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	eventSettings, err := db.GetEventSettings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	colWidths := map[string]float64{"Id": 12, "Name": 80, "Location": 80, "RookieYear": 23}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Team List - "+eventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Id"], rowHeight, "Team", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Name"], rowHeight, "Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Location"], rowHeight, "Location", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["RookieYear"], rowHeight, "Rookie Year", "1", 1, "C", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	for _, team := range teams {
		pdf.CellFormat(colWidths["Id"], rowHeight, strconv.Itoa(team.Id), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths["Name"], rowHeight, team.Nickname, "1", 0, "L", false, 0, "")
		location := fmt.Sprintf("%s, %s, %s", team.City, team.StateProv, team.Country)
		pdf.CellFormat(colWidths["Location"], rowHeight, location, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths["RookieYear"], rowHeight, strconv.Itoa(team.RookieYear), "1", 1, "L", false, 0, "")
	}
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
