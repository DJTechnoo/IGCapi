package main


import (

	"fmt"
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"github.com/DJTechnoo/goigc"
	"google.golang.org/appengine"

)



// Meta will hold json for metadata /api
type Meta struct {
	Uptime string `json:"uptime"`
	Info string `json:"info"`
	Version string `json:"version"`
}



// Fields will hold json for Track
type Fields struct {
	H_date time.Time `json:"H_date"`
	Pilot string `json:"pilot"`
	Glider string `json:"glider"`
	Glider_id string `json:"glider_id"`
	Track_len float64 `json:"track_lenght"`
}



// constants

const ROOT = "/igcinfo"		// this is the root of the app
const ID_ARG = 4			// URL index for ID
const FIELD_ARG = 5 		// URL index for FIELD
const APPTIME = 1539395670	// unix time of deployment


// GLOBAL Variables and datastructures
var lastId int				// Unique last id
var ids [] string			// slice of ids
var igcs map[int]string		// urls get associated with ids
var startTime time.Time		// for UPTIME



//		Serves /igcinfo/api/
//		Outputs metadata for this app in json
func metaHandler(w http.ResponseWriter, r * http.Request){

	meta :=	Meta{
				Uptime: calculateDuration(time.Since(startTime)),
				Info: "Api for IGC files",
				Version: "v0.1"}
					
	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil{
		fmt.Fprintln(w, err)
		return
	}
	
	fmt.Fprintf(w, string(m))
	
	
}




//		Handles arguments passed in the URL
//		and ID and FIELD and searches through the URL map
//		to get the Track

func trackJson(trackUrl string , w http.ResponseWriter, r * http.Request){
		
		
		track, err := igc.ParseLocation(trackUrl, r)
		if err != nil {
		    http.Error(w, err.Error(), 500)
		    return
		}
		
		
		trackLen := track.Task.Distance()
		
		fields := Fields {track.Date, track.Pilot, track.GliderType, track.GliderID, trackLen}
		m, err := json.MarshalIndent(&fields, "", "    ")
		if err != nil{
			fmt.Fprintln(w, err)
			return
		}
	
		fmt.Fprintf(w, string(m))

}




//	Takes the IGC ID, and Field as arguments
//	prints on the screen the specified field
//	of specified ID
func trackField(index int, field string, w http.ResponseWriter, r * http.Request){

		trackUrl := igcs[index]
		track, err := igc.ParseLocation(trackUrl, r)
		if err != nil {
		    http.Error(w, err.Error(), 500)
		    return
		}
		
		trackLen := track.Task.Distance()
		
		switch field {
			case "pilot": fmt.Fprintln(w, track.Pilot)
			case "track_length": fmt.Fprintln(w, trackLen)
			case "glider":	fmt.Fprintln(w, track.GliderType)
			case "glider_id": fmt.Fprintln(w, track.GliderID)
			case "H_date": fmt.Fprintln(w, track.Date)
			default: fmt.Fprintln(w, "NOT FOUND")
		
		}


}



//	Handles the last two arguments for <ID> and <FIELD>
//
//
func argsHandler(w http.ResponseWriter, r * http.Request){

	parts := strings.Split(r.URL.Path, "/")				// array of url parts
	
	if len(parts) > ID_ARG && len(parts) < FIELD_ARG+1{
		index, _ := strconv.Atoi(parts[ID_ARG])
		s := igcs[index]
		trackJson(s, w, r)
		
	}
	
	if len(parts) > FIELD_ARG {
		index, _ := strconv.Atoi(parts[ID_ARG])
		field := string(parts[FIELD_ARG])
		if index >= 0 && index < lastId {
			trackField(index, field, w, r)
		}
		
		
	}
}




//		handles POST and GET
//		makes use of "form.html" to get a URL to igc file
//		The URL gets stored with a unique ID in a map
//		json array outputs list of id's
func inputHandler(w http.ResponseWriter, r * http.Request){

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
		    trackUrl := r.FormValue("link")
		    if _, err := igc.ParseLocation(trackUrl, r); err != nil {
		    	http.Error(w, err.Error(), 500)
		    	return
		    }
		    
		    igcs[lastId] = string(trackUrl)
		    idManager(w);
		    
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
func idManager(w http.ResponseWriter){
	ids = append(ids, strconv.Itoa(lastId))
    lastId++
    
    idsJson, _ := json.MarshalIndent(ids, "", "    ")
    fmt.Fprintln(w, string(idsJson))
    
}



//	Input: Time in seconds
//	Output: string of ISO 8601 of said time
//
func calculateDuration(t time.Duration)(string){
	startTime = time.Now()
	totalTime := int(startTime.Unix()) - APPTIME //int(t) / int(time.Second)

	remainderSeconds 	:= totalTime%60				// final seconds
	minutes				:= totalTime / 60
	remainderMinutes	:= minutes%60					// final minutes
	hours				:= minutes / 60
	remainderHours		:= hours%24					// final hours
	days				:= hours / 24
	remainderDays		:= days%7						// final days
	months				:= days / 28
	remainderMonths		:= months%12 					// final months
	years				:= months / 12		// final years

	
	s := "P"+strconv.Itoa(years)+"Y"+strconv.Itoa(remainderMonths)+"M"+strconv.Itoa(remainderDays)+"D"+strconv.Itoa(remainderHours)+"H"+strconv.Itoa(remainderMinutes)+"M"+strconv.Itoa(remainderSeconds)+"S"
	return s	
}




func main(){
	startTime = time.Now()
	igcs = make(map[int]string)
	http.HandleFunc(ROOT + "/api", metaHandler);
	http.HandleFunc(ROOT + "/api/igc", inputHandler);
	http.HandleFunc(ROOT + "/api/igc/", argsHandler);
	appengine.Main()


}






































