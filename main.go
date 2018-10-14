// App made by Askel Eirik Johansson
// with use of goigc

package main

import (
	"encoding/json"
	"fmt"
	"github.com/DJTechnoo/goigc"
	"google.golang.org/appengine"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Meta will hold json for metadata /api
type Meta struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

// Fields will hold json for Track
type Fields struct {
	HDate    time.Time `json:"H_date"`
	Pilot    string    `json:"pilot"`
	Glider   string    `json:"glider"`
	GliderID string    `json:"glider_id"`
	TrackLen float64   `json:"track_lenght"`
}

// constants

const root = "/igcinfo"    // this is the root of the app
const idArg = 4            // URL index for ID
const fieldArg = 5         // URL index for FIELD
const appTime = 1539395670 // unix time of deployment

// GLOBAL Variables and datastructures
var lastID int          // Unique last id
var ids []string        // slice of ids
var igcs map[int]string // urls get associated with ids
var startTime time.Time // for UPTIME

//		Serves /igcinfo/api/
//		Outputs metadata for this app in json
func metaHandler(w http.ResponseWriter, r *http.Request) {

	meta := Meta{
		Uptime:  calculateDuration(time.Since(startTime)),
		Info:    "Api for IGC files",
		Version: "v0.1"}

	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, string(m))

}

//		Handles arguments passed in the URL
//		and ID and FIELD and searches through the URL map
//		to get the Track

func trackJSON(trackURL string, w http.ResponseWriter, r *http.Request) {

	track, err := igc.ParseLocation(trackURL, r)
	if err != nil {
		status := 404
		http.Error(w, http.StatusText(status), status)
		return
	}

	trackLen := track.Task.Distance()

	fields := Fields{track.Date, track.Pilot, track.GliderType, track.GliderID, trackLen}
	m, err := json.MarshalIndent(&fields, "", "    ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, string(m))

}

//	Takes the IGC ID, and Field as arguments
//	prints on the screen the specified field
//	of specified ID
func trackField(index int, field string, w http.ResponseWriter, r *http.Request) {

	trackURL := igcs[index]
	track, err := igc.ParseLocation(trackURL, r)
	if err != nil {
		status := 404
		http.Error(w, http.StatusText(status), status)
		return
	}

	trackLen := track.Task.Distance()

	switch field {
	case "pilot":
		fmt.Fprintln(w, track.Pilot)
	case "track_length":
		fmt.Fprintln(w, trackLen)
	case "glider":
		fmt.Fprintln(w, track.GliderType)
	case "glider_id":
		fmt.Fprintln(w, track.GliderID)
	case "H_date":
		fmt.Fprintln(w, track.Date)
	default:
		status := 404
		http.Error(w, http.StatusText(status), status)
		return

	}

}

//	Handles the last two arguments for <ID> and <FIELD>
//
//
func argsHandler(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/") // array of url parts

	if len(parts) > fieldArg+1 {
		status := 400
		http.Error(w, http.StatusText(status), status)
		return
	}

	if len(parts) > idArg && len(parts) < fieldArg+1 {
		index, err := strconv.Atoi(parts[idArg])
		if err != nil {
			status := 404
			http.Error(w, http.StatusText(status), status)
			return
		}
		s := igcs[index]
		trackJSON(s, w, r)

	}

	if len(parts) > fieldArg {
		index, err := strconv.Atoi(parts[idArg])
		if err != nil {
			status := 404
			http.Error(w, http.StatusText(status), status)
			return
		}

		if index < 0 || index >= lastID {
			status := 404
			http.Error(w, http.StatusText(status), status)
			return
		}

		field := string(parts[fieldArg])
		trackField(index, field, w, r)

	}
}

//		handles POST and GET
//		makes use of "form.html" to get a URL to igc file
//		The URL gets stored with a unique ID in a map
//		json array outputs list of id's
func inputHandler(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 5 {
		switch r.Method {
		case "GET":
			http.ServeFile(w, r, "form.html")
		case "POST":

			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}

			// get url from form and ten see if the url is valid. If it is, store in map
			trackURL := r.FormValue("link")
			if _, err := igc.ParseLocation(trackURL, r); err != nil {
				status := 400
				http.Error(w, http.StatusText(status), status)
				return
			}

			igcs[lastID] = string(trackURL)
			idManager(w)

		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
	} else {
		fmt.Fprintln(w, "More params!")
	}
}

//	Creates a new ID for the next url
//	and appends to ID-slice.
//
func idManager(w http.ResponseWriter) {
	ids = append(ids, strconv.Itoa(lastID))
	lastID++

	idsJSON, _ := json.MarshalIndent(ids, "", "    ")
	fmt.Fprintln(w, string(idsJSON))

}

//	Input: Time in seconds
//	Output: string of ISO 8601 of said time
//
func calculateDuration(t time.Duration) string {
	startTime = time.Now()
	totalTime := int(startTime.Unix()) - appTime //int(t) / int(time.Second)

	remainderSeconds := totalTime % 60 // final seconds
	minutes := totalTime / 60
	remainderMinutes := minutes % 60 // final minutes
	hours := minutes / 60
	remainderHours := hours % 24 // final hours
	days := hours / 24
	remainderDays := days % 7 // final days
	months := days / 28
	remainderMonths := months % 12 // final months
	years := months / 12           // final years

	s := "P" + strconv.Itoa(years) + "Y" + strconv.Itoa(remainderMonths) + "M" + strconv.Itoa(remainderDays) + "D" + strconv.Itoa(remainderHours) + "H" + strconv.Itoa(remainderMinutes) + "M" + strconv.Itoa(remainderSeconds) + "S"
	return s
}

func main() {
	startTime = time.Now()
	igcs = make(map[int]string)
	http.HandleFunc(root+"/api", metaHandler)
	http.HandleFunc(root+"/api/igc", inputHandler)
	http.HandleFunc(root+"/api/igc/", argsHandler)
	appengine.Main()

}
