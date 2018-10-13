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


type Meta struct {
	Uptime string `json:"uptime"`
	Info string `json:"info"`
	Version string `json:"version"`
}


/*type Id struct {
	Id int `json: "id"`
}*/


// lastId
// map id->url

// constants

const ROOT = "/igcinfo"		// this is the root of the app
const ID_ARG = 4			// URL index for ID
const FIELD_ARG = 5 		// URL index for FIELD


var lastId int				// Unique last id
var ids [] string			// array of ids
var igcs map[int]string		// urls get associated with ids
var startTime time.Time



//		Serves /igcinfo/api/
//		Outputs metadata for this app in json
func metaHandler(w http.ResponseWriter, r * http.Request){

	meta :=		Meta{
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
func argsHandler(w http.ResponseWriter, r * http.Request){

	parts := strings.Split(r.URL.Path, "/")				// array of url parts
	
	if len(parts) > ID_ARG{
		fmt.Fprintln(w, "first arg " + parts[ID_ARG]);
	}
	
	if len(parts) > FIELD_ARG {
		fmt.Fprintln(w, "second arg " + parts[FIELD_ARG]);
		index, _ := strconv.Atoi(parts[FIELD_ARG])
		s := igcs[index]
		//s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
		track, err := igc.ParseLocation(s, r)
		if err != nil {
			//fmt.Fprintln(w, "OMG NO")
		    http.Error(w, err.Error(), 500)
		}

		fmt.Fprintf(w, "Pilot: %s, gliderType: %s, date: %s", 
		    track.Pilot, track.GliderType, track.Date.String())
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
		    
		    
		    link := r.FormValue("link")
		    // check url first
		    
		    
		    igcs[lastId] = string(link)
		    fmt.Fprintln(w, igcs);
		    idManager(w);
		 
		    
	   
		    fmt.Fprintf(w, "link = %s\n", link)
		default:
		    fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
    } else {
    	fmt.Fprintln(w, "More params!")
    }
}



func idManager(w http.ResponseWriter){
	ids = append(ids, strconv.Itoa(lastId))
    lastId++
    
    idsJson, _ := json.MarshalIndent(ids, "", "    ")
    fmt.Fprintln(w, string(idsJson))
    
}


func calculateDuration(t time.Duration)(string){
	totalTime := int(t) / int(time.Second)

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
	//http.HandleFunc("/api/igc/<id>", inputHandler);
	//http.ListenAndServe(":8080", nil);
	appengine.Main()


}






































