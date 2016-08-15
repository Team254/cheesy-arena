// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for generating CSV and PDF reports.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
	"net/http"
	"strconv"
	"text/template"
)

// Generates a CSV-formatted report of the qualification rankings.
func RankingsCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	rankings, err := db.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
	w.Header().Set("Content-Type", "text/plain")
	template, err := template.ParseFiles("templates/rankings.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.Execute(w, rankings)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the qualification rankings.
func RankingsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	rankings, err := db.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Rank": 13, "Team": 23, "RP": 20, "Auto": 20, "SC": 21, "Goal": 20,
		"Defense": 20, "W-L-T": 20, "DQ": 20, "Played": 20}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()

	// Render table header row.
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Team Standings - "+eventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Rank"], rowHeight, "Rank", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Team", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["RP"], rowHeight, "RB", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Auto"], rowHeight, "Auto", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["SC"], rowHeight, "Scale/Chal.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Goal"], rowHeight, "Goal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Defense", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["W-L-T"], rowHeight, "W-L-T", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["DQ"], rowHeight, "DQ", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Played"], rowHeight, "Played", "1", 1, "C", true, 0, "")
	for _, ranking := range rankings {
		// Render ranking info row.
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(colWidths["Rank"], rowHeight, strconv.Itoa(ranking.Rank), "1", 0, "C", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(colWidths["Team"], rowHeight, strconv.Itoa(ranking.TeamId), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["RP"], rowHeight, strconv.Itoa(ranking.RankingPoints), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Auto"], rowHeight, strconv.Itoa(ranking.AutoPoints), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["SC"], rowHeight, strconv.Itoa(ranking.ScaleChallengePoints), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Goal"], rowHeight, strconv.Itoa(ranking.GoalPoints), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, strconv.Itoa(ranking.DefensePoints), "1", 0, "C", false, 0, "")
		record := fmt.Sprintf("%d-%d-%d", ranking.Wins, ranking.Losses, ranking.Ties)
		pdf.CellFormat(colWidths["W-L-T"], rowHeight, record, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["DQ"], rowHeight, strconv.Itoa(ranking.Disqualifications), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Played"], rowHeight, strconv.Itoa(ranking.Played), "1", 1, "C", false, 0, "")
	}

	// Write out the PDF file as the HTTP response.
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a CSV-formatted report of the match schedule.
func ScheduleCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	vars := mux.Vars(r)
	matches, err := db.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
	w.Header().Set("Content-Type", "text/plain")
	template, err := template.ParseFiles("templates/schedule.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.Execute(w, matches)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the match schedule.
func SchedulePdfReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	vars := mux.Vars(r)
	matches, err := db.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesPerTeam := 0
	if len(teams) > 0 {
		matchesPerTeam = len(matches) * teamsPerMatch / len(teams)
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Time": 35, "Type": 25, "Match": 15, "Team": 20}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()

	// Render table header row.
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Match Schedule - "+eventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Time"], rowHeight, "Time", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Type"], rowHeight, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Match"], rowHeight, "Match", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Red 1", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Red 2", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Red 3", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Blue 1", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Blue 2", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Blue 3", "1", 1, "C", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	for _, match := range matches {
		height := rowHeight
		borderStr := "1"
		alignStr := "CM"
		surrogate := false
		if match.Red1IsSurrogate || match.Red2IsSurrogate || match.Red3IsSurrogate ||
			match.Blue1IsSurrogate || match.Blue2IsSurrogate || match.Blue3IsSurrogate {
			// If the match contains surrogates, the row needs to be taller to fit some text beneath team numbers.
			height = 5.0
			borderStr = "LTR"
			alignStr = "CB"
			surrogate = true
		}

		// Capitalize match types.
		matchType := match.CapitalizedType()

		formatTeam := func(teamId int) string {
			if teamId == 0 {
				return ""
			} else {
				return strconv.Itoa(teamId)
			}
		}

		// Render match info row.
		pdf.CellFormat(colWidths["Time"], height, match.Time.Local().Format("Mon 1/02 03:04 PM"), borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Type"], height, matchType, borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Match"], height, match.DisplayName, borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Team"], height, formatTeam(match.Red1), borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Team"], height, formatTeam(match.Red2), borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Team"], height, formatTeam(match.Red3), borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Team"], height, formatTeam(match.Blue1), borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Team"], height, formatTeam(match.Blue2), borderStr, 0, alignStr, false, 0, "")
		pdf.CellFormat(colWidths["Team"], height, formatTeam(match.Blue3), borderStr, 1, alignStr, false, 0, "")
		if surrogate {
			// Render the text that indicates which teams are surrogates.
			height := 4.0
			pdf.SetFont("Arial", "", 8)
			pdf.CellFormat(colWidths["Time"], height, "", "LBR", 0, "C", false, 0, "")
			pdf.CellFormat(colWidths["Type"], height, "", "LBR", 0, "C", false, 0, "")
			pdf.CellFormat(colWidths["Match"], height, "", "LBR", 0, "C", false, 0, "")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Red1IsSurrogate), "LBR", 0, "CT", false, 0, "")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Red2IsSurrogate), "LBR", 0, "CT", false, 0, "")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Red3IsSurrogate), "LBR", 0, "CT", false, 0, "")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Blue1IsSurrogate), "LBR", 0, "CT", false, 0, "")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Blue2IsSurrogate), "LBR", 0, "CT", false, 0, "")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Blue3IsSurrogate), "LBR", 1, "CT", false, 0, "")
			pdf.SetFont("Arial", "", 10)
		}
	}

	if vars["type"] != "elimination" {
		// Render some summary info at the bottom.
		pdf.CellFormat(195, 10, fmt.Sprintf("Matches Per Team: %d", matchesPerTeam), "", 1, "L", false, 0, "")
	}

	// Write out the PDF file as the HTTP response.
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the defenses schedule.
func DefensesPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	vars := mux.Vars(r)
	matches, err := db.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Type": 25, "Match": 15, "Defense": 15.5}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()

	// Render table header row.
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Defenses Schedule - "+eventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Type"], rowHeight, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Match"], rowHeight, "Match", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Red 1", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Red 2", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Red 3", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Red 4", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Red 5", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Blue 1", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Blue 2", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Blue 3", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Blue 4", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Defense"], rowHeight, "Blue 5", "1", 1, "C", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	for _, match := range matches {
		// Render match defenses row.
		pdf.CellFormat(colWidths["Type"], rowHeight, match.CapitalizedType(), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Match"], rowHeight, match.DisplayName, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.RedDefense1, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.RedDefense2, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.RedDefense3, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.RedDefense4, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.RedDefense5, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.BlueDefense1, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.BlueDefense2, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.BlueDefense3, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.BlueDefense4, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Defense"], rowHeight, match.BlueDefense5, "1", 1, "C", false, 0, "")
	}

	// Write out the PDF file as the HTTP response.
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a CSV-formatted report of the team list.
func TeamsCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
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

// Generates a PDF-formatted report of the team list.
func TeamsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsReader(w, r) {
		return
	}

	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Id": 12, "Name": 80, "Location": 80, "RookieYear": 23}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)

	// Render table header row.
	pdf.CellFormat(195, rowHeight, "Team List - "+eventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Id"], rowHeight, "Team", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Name"], rowHeight, "Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Location"], rowHeight, "Location", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["RookieYear"], rowHeight, "Rookie Year", "1", 1, "C", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	for _, team := range teams {
		// Render team info row.
		pdf.CellFormat(colWidths["Id"], rowHeight, strconv.Itoa(team.Id), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths["Name"], rowHeight, team.Nickname, "1", 0, "L", false, 0, "")
		location := fmt.Sprintf("%s, %s, %s", team.City, team.StateProv, team.Country)
		pdf.CellFormat(colWidths["Location"], rowHeight, location, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths["RookieYear"], rowHeight, strconv.Itoa(team.RookieYear), "1", 1, "L", false, 0, "")
	}

	// Write out the PDF file as the HTTP response.
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a CSV-formatted report of the WPA keys, for import into the radio kiosk.
func WpaKeysCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	if !UserIsAdmin(w, r) {
		return
	}

	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=wpa_keys.csv")
	for _, team := range teams {
		_, err := w.Write([]byte(fmt.Sprintf("%d,%s\r\n", team.Id, team.WpaKey)))
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}
}

// Returns the text to display if a team is a surrogate.
func surrogateText(isSurrogate bool) string {
	if isSurrogate {
		return "(surrogate)"
	} else {
		return ""
	}
}
