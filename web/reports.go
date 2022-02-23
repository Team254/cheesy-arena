// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for generating CSV and PDF reports.

package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
)

// Generates a CSV-formatted report of the qualification rankings.
func (web *Web) rankingsCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
	w.Header().Set("Content-Type", "text/plain")
	template, err := web.parseFiles("templates/rankings.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.ExecuteTemplate(w, "rankings.csv", rankings)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the qualification rankings.
func (web *Web) rankingsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Rank": 13, "Team": 22, "RP": 23, "Match": 22, "Hangar": 22, "Auto": 25,
		"W-L-T": 23, "DQ": 23, "Played": 23}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()

	// Render table header row.
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Team Standings - "+web.arena.EventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Rank"], rowHeight, "Rank", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Team", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["RP"], rowHeight, "RP", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Match"], rowHeight, "Match", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Hangar"], rowHeight, "Hangar", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Auto"], rowHeight, "Taxi+A.Cargo", "1", 0, "C", true, 0, "")
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
		pdf.CellFormat(colWidths["Match"], rowHeight, strconv.Itoa(ranking.MatchPoints), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Hangar"], rowHeight, strconv.Itoa(ranking.HangarPoints), "1", 0, "C", false, 0, "")
		pdf.CellFormat(
			colWidths["Auto"], rowHeight, strconv.Itoa(ranking.TaxiAndAutoCargoPoints), "1", 0, "C", false, 0, "",
		)
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

// findBackupTeams takes the list of teams at the event and returns a slice of
// teams with the teams that are already members of alliances removed. The
// second returned value is the set of teams that were backups but have already
// been called back to the field.
//
// At events that run 4 team alliances, this will show all of the 3rd picks and
// remaining teams.
func (web *Web) findBackupTeams(rankings game.Rankings) (game.Rankings, map[int]bool, error) {
	var pruned game.Rankings

	picks, err := web.arena.Database.GetAllAlliances()
	if err != nil {
		return nil, nil, err
	}

	if len(picks) == 0 {
		return nil, nil, errors.New("backup teams report is unavailable until alliances have been selected")
	}

	pickedTeams := make(map[int]bool)
	pickedBackups := make(map[int]bool)

	for _, alliance := range picks {
		for _, team := range alliance {
			// Teams in third in an alliance are backups at events that use 3 team alliances.
			if team.PickPosition == 3 {
				pickedBackups[team.TeamId] = true
				continue
			}
			pickedTeams[team.TeamId] = true
		}
	}

	for _, team := range rankings {
		if !pickedTeams[team.TeamId] {
			pruned = append(pruned, team)
		}
	}

	return pruned, pickedBackups, nil
}

// Define a backupTeam type so that we can pass the additional "Called" field
// to the CSV template parser.
type backupTeam struct {
	Rank          int
	Called        bool
	TeamId        int
	RankingPoints int
}

// Generates a CSV-formatted report of the qualification rankings.
func (web *Web) backupTeamsCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	rankings, pickedBackups, err := web.findBackupTeams(rankings)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Copy the list of teams that are eligible backups and annotate them with
	// whether or not they've been picked already.
	var backupTeams []backupTeam
	for _, r := range rankings {
		backupTeams = append(backupTeams, backupTeam{
			Rank:          r.Rank,
			Called:        pickedBackups[r.TeamId],
			TeamId:        r.TeamId,
			RankingPoints: r.RankingPoints,
		})
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
	w.Header().Set("Content-Type", "text/plain")
	template, err := web.parseFiles("templates/backups.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.ExecuteTemplate(w, "backups.csv", backupTeams)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the backup teams.
func (web *Web) backupsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	rankings, pickedBackups, err := web.findBackupTeams(rankings)
	_ = pickedBackups
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Rank": 13, "Called": 22, "Team": 22, "RP": 23}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()

	// Render table header row.
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Backup Teams - "+web.arena.EventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Rank"], rowHeight, "Rank", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Called"], rowHeight, "Called?", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Team"], rowHeight, "Team", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["RP"], rowHeight, "RP", "1", 1, "C", true, 0, "")
	for _, ranking := range rankings {
		var picked string
		if pickedBackups[ranking.TeamId] {
			picked = "Y"
		}

		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(colWidths["Rank"], rowHeight, strconv.Itoa(ranking.Rank), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["Called"], rowHeight, picked, "1", 0, "C", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(colWidths["Team"], rowHeight, strconv.Itoa(ranking.TeamId), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["RP"], rowHeight, strconv.Itoa(ranking.RankingPoints), "1", 1, "C", false, 0, "")
	}

	// Write out the PDF file as the HTTP response.
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Coupon constants used in laying out the playoff alliance coupons.
const (
	// All units in mm
	cHPad       = 5
	cVPad       = 5
	cWidth      = 95
	cHeight     = 60
	cSideMargin = 10
	cTopMargin  = 10
	cImgWidth   = 25
	cWOffset    = 5
)

func (web *Web) couponsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.SetLineWidth(1)

	alliances, err := web.arena.Database.GetAllAlliances()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(alliances) == 0 {
		handleWebErr(w, errors.New("playoff alliance coupons report is unavailable until alliances have been selected"))
		return
	}

	eventName := web.arena.EventSettings.Name

	for page := 0; page < (len(alliances)+3)/4; page++ {
		heightAcc := cTopMargin
		pdf.AddPage()
		for i := page * 4; i < page*4+4 && i < len(alliances); i++ {
			pdf.SetFillColor(220, 220, 220)

			allianceCaptain := alliances[i][0].TeamId

			pdf.RoundedRect(cSideMargin, float64(heightAcc), cWidth, cHeight, 4, "1234", "D")
			timeoutX := cSideMargin + (cWidth * 0.5)
			timeoutY := float64(heightAcc) + (cHeight * 0.5)
			drawTimeoutCoupon(pdf, eventName, timeoutX, timeoutY, allianceCaptain, i+1)

			pdf.RoundedRect(cWidth+cHPad+cSideMargin, float64(heightAcc), cWidth, cHeight, 4, "1234", "D")
			backupX := cSideMargin + cWidth + cHPad + (cWidth * 0.5)
			backupY := float64(heightAcc) + (cHeight * 0.5)
			heightAcc += cHeight + cVPad
			drawBackupCoupon(pdf, eventName, backupX, backupY, allianceCaptain, i+1)
		}
	}

	// Write out the PDF file as the HTTP response.
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func drawTimeoutCoupon(pdf gofpdf.Pdf, eventName string, x float64, y float64, teamId int, allianceNumber int) {
	pdf.SetTextColor(0, 0, 0)
	drawPdfLogo(pdf, x, y, cImgWidth)

	pdf.SetFont("Arial", "B", 24)
	drawCenteredText(pdf, "Timeout Coupon", x, y+10)

	pdf.SetFont("Arial", "", 14)
	drawCenteredText(pdf, fmt.Sprintf("Alliance: %v    Captain: %v", allianceNumber, teamId), x, y+20)
	drawEventWatermark(pdf, x, y, eventName)
}

func drawBackupCoupon(pdf gofpdf.Pdf, eventName string, x float64, y float64, teamId int, allianceNumber int) {
	pdf.SetTextColor(0, 0, 0)
	drawPdfLogo(pdf, x, y, cImgWidth)

	pdf.SetFont("Arial", "B", 24)
	drawCenteredText(pdf, "Backup Coupon", x, y+10)

	pdf.SetFont("Arial", "", 14)
	drawCenteredText(pdf, fmt.Sprintf("Alliance: %v    Captain: %v", allianceNumber, teamId), x, y+20)
	drawEventWatermark(pdf, x, y, eventName)
}

func drawEventWatermark(pdf gofpdf.Pdf, x float64, y float64, name string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(200, 200, 200)
	textWidth := pdf.GetStringWidth(name)

	// Left mark
	pdf.TransformBegin()
	pdf.TransformRotate(90, x, y)
	pdf.Text(x-textWidth/2, y-cWidth/2+cWOffset, name)
	pdf.TransformEnd()

	// Right mark
	pdf.TransformBegin()
	pdf.TransformRotate(270, x, y)
	pdf.Text(x-textWidth/2, y-cWidth/2+cWOffset, name)
	pdf.TransformEnd()
}

func drawCenteredText(pdf gofpdf.Pdf, txt string, x float64, y float64) {
	width := pdf.GetStringWidth(txt)
	pdf.Text(x-(width/2), y, txt)
}

func drawPdfLogo(pdf gofpdf.Pdf, x float64, y float64, width float64) {
	pdf.ImageOptions("static/img/game-logo.png", x-(width/2), y-25, width, 0, false,
		gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
}

// Generates a CSV-formatted report of the match schedule.
func (web *Web) scheduleCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matches, err := web.arena.Database.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
	w.Header().Set("Content-Type", "text/plain")
	template, err := web.parseFiles("templates/schedule.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.ExecuteTemplate(w, "schedule.csv", matches)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the match schedule.
func (web *Web) schedulePdfReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matches, err := web.arena.Database.GetMatchesByType(vars["type"])
	if err != nil {
		handleWebErr(w, err)
		return
	}
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesPerTeam := 0
	if len(teams) > 0 {
		matchesPerTeam = len(matches) * tournament.TeamsPerMatch / len(teams)
	}

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	colWidths := map[string]float64{"Time": 35, "Type": 25, "Match": 15, "Team": 20}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()

	// Render table header row.
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(195, rowHeight, "Match Schedule - "+web.arena.EventSettings.Name, "", 1, "C", false, 0, "")
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
		pdf.CellFormat(colWidths["Time"], height, match.Time.Local().Format("Mon 1/02 03:04 PM"), borderStr, 0,
			alignStr, false, 0, "")
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
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Red1IsSurrogate), "LBR", 0, "CT", false, 0,
				"")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Red2IsSurrogate), "LBR", 0, "CT", false, 0,
				"")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Red3IsSurrogate), "LBR", 0, "CT", false, 0,
				"")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Blue1IsSurrogate), "LBR", 0, "CT", false, 0,
				"")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Blue2IsSurrogate), "LBR", 0, "CT", false, 0,
				"")
			pdf.CellFormat(colWidths["Team"], height, surrogateText(match.Blue3IsSurrogate), "LBR", 1, "CT", false, 0,
				"")
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

// Generates a CSV-formatted report of the team list.
func (web *Web) teamsCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// Don't set the content type as "text/csv", as that will trigger an automatic download in the browser.
	w.Header().Set("Content-Type", "text/plain")
	template, err := web.parseFiles("templates/teams.csv")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.ExecuteTemplate(w, "teams.csv", teams)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Generates a PDF-formatted report of the team list.
func (web *Web) teamsPdfReportHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}

	showHasConnected := r.URL.Query().Get("showHasConnected") == "true"

	// The widths of the table columns in mm, stored here so that they can be referenced for each row.
	var colWidths map[string]float64
	if showHasConnected {
		colWidths = map[string]float64{"Id": 12, "Name": 70, "Location": 65, "RookieYear": 23, "HasConnected": 25}
	} else {
		colWidths = map[string]float64{"Id": 12, "Name": 80, "Location": 80, "RookieYear": 23}
	}
	rowHeight := 6.5

	pdf := gofpdf.New("P", "mm", "Letter", "font")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)

	// Render table header row.
	pdf.CellFormat(195, rowHeight, "Team List - "+web.arena.EventSettings.Name, "", 1, "C", false, 0, "")
	pdf.CellFormat(colWidths["Id"], rowHeight, "Team", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Name"], rowHeight, "Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["Location"], rowHeight, "Location", "1", 0, "C", true, 0, "")
	if showHasConnected {
		pdf.CellFormat(colWidths["RookieYear"], rowHeight, "Rookie Year", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colWidths["HasConnected"], rowHeight, "Connected?", "1", 1, "C", true, 0, "")
	} else {
		pdf.CellFormat(colWidths["RookieYear"], rowHeight, "Rookie Year", "1", 1, "C", true, 0, "")
	}
	pdf.SetFont("Arial", "", 10)
	for _, team := range teams {
		// Render team info row.
		pdf.CellFormat(colWidths["Id"], rowHeight, strconv.Itoa(team.Id), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths["Name"], rowHeight, team.Nickname, "1", 0, "L", false, 0, "")
		location := fmt.Sprintf("%s, %s, %s", team.City, team.StateProv, team.Country)
		pdf.CellFormat(colWidths["Location"], rowHeight, location, "1", 0, "L", false, 0, "")
		if showHasConnected {
			pdf.CellFormat(colWidths["RookieYear"], rowHeight, strconv.Itoa(team.RookieYear), "1", 0, "L", false, 0, "")
			var hasConnected string
			if team.HasConnected {
				hasConnected = "Yes"
			}
			pdf.CellFormat(colWidths["HasConnected"], rowHeight, hasConnected, "1", 1, "L", false, 0, "")
		} else {
			pdf.CellFormat(colWidths["RookieYear"], rowHeight, strconv.Itoa(team.RookieYear), "1", 1, "L", false, 0, "")
		}
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
func (web *Web) wpaKeysCsvReportHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	teams, err := web.arena.Database.GetAllTeams()
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
