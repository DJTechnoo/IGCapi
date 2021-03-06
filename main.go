// App made by Askel Eirik Johansson
// with use of goigc

package main

import (
	"encoding/json"
	"fmt"
	//"io"
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
	http.Header.Add(w.Header(), "content-type", "application/json")
	meta := Meta{
		Uptime:  calculateDuration(time.Since(startTime)),
		Info:    "Service for IGC tracks.",
		Version: "v1"}

	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil {
		status := 400
		http.Error(w, http.StatusText(status), status)
		return
	}

	_, err = fmt.Fprintf(w, string(m))
	if err != nil {
		status := 500
		http.Error(w, http.StatusText(status), status)
		return
	}

}

//		Handles arguments passed in the URL
//		and ID and FIELD and searches through the URL map
//		to get the Track

func trackJSON(trackURL string, w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	track, err := igc.ParseLocation(trackURL, r)
	if err != nil {
		status := 400
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Calculate total track distance
	totalDistance := 0.0
	for i := 0; i < len(track.Points)-1; i++ {
		totalDistance += track.Points[i].Distance(track.Points[i+1])
	}

	fields := Fields{track.Date, track.Pilot, track.GliderType, track.GliderID, totalDistance}
	m, err := json.MarshalIndent(&fields, "", "    ")
	if err != nil {
		status := 500
		http.Error(w, http.StatusText(status), status)
		return
	}

	_, err = fmt.Fprintf(w, string(m))
	if err != nil {
		status := 500
		http.Error(w, http.StatusText(status), status)
		return
	}

}

//	Takes the IGC ID, and Field as arguments
//	prints on the screen the specified field
//	of specified ID
func trackField(index int, field string, w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "text/plain")
	trackURL := igcs[index]
	track, err := igc.ParseLocation(trackURL, r)
	if err != nil {
		status := 404
		http.Error(w, http.StatusText(status), status)
		return
	}

	switch field {
	case "pilot":
		_, _ = fmt.Fprintln(w, track.Pilot)

	case "track_length":
		// Calculate total track distance
		totalDistance := 0.0
		for i := 0; i < len(track.Points)-1; i++ {
			totalDistance += track.Points[i].Distance(track.Points[i+1])
		}

		_, _ = fmt.Fprintln(w, totalDistance)

	case "glider":
		_, _ = fmt.Fprintln(w, track.GliderType)

	case "glider_id":
		_, _ = fmt.Fprintln(w, track.GliderID)

	case "H_date":
		_, _ = fmt.Fprintln(w, track.Date)

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

		field := parts[fieldArg]
		trackField(index, field, w, r)

	}
}

//		handles POST and GET
//		makes use of "form.html" to get a URL to igc file
//		The URL gets stored with a unique ID in a map
//		json array outputs list of id's
func inputHandler(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 5 {
		switch r.Method {
		case "GET":
			showIDs(w)
		case "POST":

			type reqURL struct {
				URL string `json:"url"`
			}

			req := reqURL{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				status := 400
				http.Error(w, http.StatusText(status), status)
				return
			}

			trackURL := req.URL

			if _, err := igc.ParseLocation(trackURL, r); err != nil {
				status := 400
				http.Error(w, http.StatusText(status), status)
				return
			}

			igcs[lastID] = trackURL
			idManager(w)

		default:
			status := 400
			http.Error(w, http.StatusText(status), status)
			return
		}
	} else {
		_, err := fmt.Fprintln(w, "More params!")
		if err != nil {
			status := 400
			http.Error(w, http.StatusText(status), status)
			return
		}
	}
}

//	Creates a new ID for the next url
//	and appends to ID-slice.
//
func idManager(w http.ResponseWriter) {
	ids = append(ids, strconv.Itoa(lastID))
	type responseID struct {
		ID string `json:"id"`
	}

	res := responseID{}
	res.ID = strconv.Itoa(lastID)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		status := 400
		http.Error(w, http.StatusText(status), status)
		return
	}

	lastID++
}

//	Returns the entire list of used IDs
func showIDs(w http.ResponseWriter) {
	if len(ids) <= 0 {
		//http.Header.Add(w.Header(), "content-type", "text/plain")
		status := 404
		http.Error(w, http.StatusText(status), status)
		return
	}

	idsJSON, err := json.MarshalIndent(ids, "", "    ")
	if err != nil {
		status := 500
		http.Error(w, http.StatusText(status), status)
		return
	}
	_, err = fmt.Fprintln(w, string(idsJSON))
	if err != nil {
		status := 500
		http.Error(w, http.StatusText(status), status)
		return
	}
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
