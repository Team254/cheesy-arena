// Copyright 20## Team ###. All Rights Reserved.
// Author: cpapplefamily@gmail.com (Corey Applegate)
//
// Alternate IO handlers for the ###.

package plc

import (
	//"github.com/Team254/cheesy-arena/game"
	//"github.com/Team254/cheesy-arena/model"
	//"encoding/json"
	//"net/http"
	"time"
	"log"
	"net"
	"strings"
)

type Esp32 interface {
	Run()
	IsScoreTableIOEnabled() bool
	IsRedEstopsEnabled() bool
	IsBlueEstopsEnabled() bool
	IsScoreTableHealthy() bool
	IsRedEstopsHealthy() bool
	IsBlueEstopsHealthy() bool
	SetScoreTableAddress(string) 
	SetRedAllianceStationEstopAddress(string) 
	SetBlueAllianceStationEstopAddress(string) 
}

type Esp32IO struct {
	ScoreTableIP		string
	RedAllianceEstopsIP		string
	BlueAllianceEstopsIP		string
	scoreTableHealthy 	bool
	RedEstopsHealthy 	bool
	BlueEstopsHealthy 	bool
}
const LoopPeriodMs = 1000 // Define the loop period in milliseconds


// RequestPayload represents the structure of the incoming POST data.
type RequestPayload struct {
	Channel int  `json:"channel"`
	State   bool `json:"state"`
}

func (esp32 *Esp32IO) SetScoreTableAddress(address string) {
	address = strings.TrimSpace(address)
	if address == "" {
		esp32.ScoreTableIP = address
        return
    }
    if net.ParseIP(address) == nil {
        log.Printf("Invalid Score Table IP address: %s", address)
        return
    }
    esp32.ScoreTableIP = address
    log.Printf("Set Score Table IP to: %s", esp32.ScoreTableIP)
}
func (esp32 *Esp32IO) SetRedAllianceStationEstopAddress(address string) {
	address = strings.TrimSpace(address)
	if address == "" {
		esp32.RedAllianceEstopsIP = address
        return
    }
    if net.ParseIP(address) == nil {
        log.Printf("Invalid Red Alliance Estops IP address: %s", address)
        return
    }
    esp32.RedAllianceEstopsIP = address
	log.Printf("Red Alliance Estops IP to: %s", esp32.RedAllianceEstopsIP)
}
func (esp32 *Esp32IO) SetBlueAllianceStationEstopAddress(address string) {
	address = strings.TrimSpace(address)
	if address == "" {
		esp32.BlueAllianceEstopsIP = address
        return
    }
    if net.ParseIP(address) == nil {
        log.Printf("Invalid Blue Alliance Estops IP address: %s", address)
        return
    }
    esp32.BlueAllianceEstopsIP = address
	log.Printf("Blue Alliance Estops IP to: %s", esp32.BlueAllianceEstopsIP)
}

// Checks if an IP address is reachable by attempting a TCP connection.
func isDevicePresent(ip string, port string) error {
    address := net.JoinHostPort(ip, port)
    conn, err := net.DialTimeout("tcp", address, time.Second*2)
    if err != nil {
        //log.Printf("Device not reachable at %s: %v", address, err)
        return err
    } 
    conn.Close()
    return err
}

// Run starts the ESP32 IO monitoring loop.
func (esp32 *Esp32IO) Run() {
	for {
		// Check if the Score Table Estops are reachable.
		if !esp32.IsScoreTableIOEnabled() {
			// If the Score Table is not enabled, don't check it.
			esp32.scoreTableHealthy = false
		} else {
			//log.Println("ScoreTable Check")
			err := isDevicePresent(esp32.ScoreTableIP, "80")
			if err != nil {
				log.Printf("Score Table not reachable at %s: %v", esp32.ScoreTableIP, err)
				time.Sleep(time.Second * plcRetryIntevalSec)
				esp32.scoreTableHealthy = false
				continue
				}else{
					if (!esp32.scoreTableHealthy){
						log.Printf("Score Table Connected at: %s", esp32.ScoreTableIP)
					}
					esp32.scoreTableHealthy = true
				}
			}
			// Check if the Red Alliance Estops are healthy.
			if !esp32.IsRedEstopsEnabled() {
				// If the Red Alliance Estops are not enabled, don't check them.
				esp32.RedEstopsHealthy= false
				} else {
			//log.Println("Red Estops IO Check")
			err := isDevicePresent(esp32.RedAllianceEstopsIP, "80")
			if err != nil {
				log.Printf("Red Alliance Estops not reachable at %s: %v", esp32.RedAllianceEstopsIP, err)
				time.Sleep(time.Second * plcRetryIntevalSec)
				esp32.RedEstopsHealthy = false
				continue
				}else{
					if (!esp32.RedEstopsHealthy){
						log.Printf("Red Estops Connected at: %s ", esp32.RedAllianceEstopsIP)
					}
					esp32.RedEstopsHealthy = true
				}
			}
			// Check if the Blue Alliance Estops are healthy.
			if !esp32.IsBlueEstopsEnabled() {
				// If the Blue Alliance Estops are not enabled, don't check them.
				esp32.BlueEstopsHealthy = false
				} else {
			//log.Println("Blue Estops IO Check")
			err := isDevicePresent(esp32.BlueAllianceEstopsIP, "80")
			if err != nil {
				log.Printf("Blue Alliance Estops not reachable at %s: %v", esp32.BlueAllianceEstopsIP, err)
				time.Sleep(time.Second * plcRetryIntevalSec)
				esp32.BlueEstopsHealthy = false
				continue
			}else{
				if (!esp32.BlueEstopsHealthy){
					log.Printf("Blue Estops Connected at: %s ", esp32.BlueAllianceEstopsIP)
				}
				esp32.BlueEstopsHealthy = true
			}
		}
		
		startTime := time.Now()
		time.Sleep(time.Until(startTime.Add(time.Millisecond * LoopPeriodMs)))
	}
}

// Returns whether the alternate IO is enabled.
func (esp32 *Esp32IO) IsScoreTableIOEnabled() bool {
	return esp32.ScoreTableIP != ""
}

// Returns whether the alternate IO is enabled.
func (esp32 *Esp32IO) IsRedEstopsEnabled() bool {
	return esp32.RedAllianceEstopsIP != ""
}

// Returns whether the alternate IO is enabled.
func (esp32 *Esp32IO) IsBlueEstopsEnabled() bool {
	return esp32.BlueAllianceEstopsIP != ""
}

// Returns the health status of the alternate IO.
func (esp32 *Esp32IO) IsScoreTableHealthy() bool {
	return esp32.scoreTableHealthy
}

// Returns the health status of the alternate IO.
func (esp32 *Esp32IO) IsRedEstopsHealthy() bool {
	return esp32.RedEstopsHealthy
}

// Returns the health status of the alternate IO.
func (esp32 *Esp32IO) IsBlueEstopsHealthy() bool {
	return esp32.BlueEstopsHealthy
}
