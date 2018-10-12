package main

import (


	"fmt"
	"encoding/json"
	"net/http"
	"strconv"
	"google.golang.org/appengine"

)


type Meta struct {
	Info string `json:"info"`
	Version string `json:"version"`
}


/*type Id struct {
	Id int `json: "id"`
}*/


// lastId
// map id->url

var lastId int
var ids [] string
var igcs map[int]string



func metaHandler(w http.ResponseWriter, r * http.Request){

	meta :=		Meta{
					Info: "Api for IGC files",
					Version: "v0.1"}
					
	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil{
		fmt.Fprintln(w, err)
		return
	}
	
	fmt.Fprintf(w, string(m))	
}




func inputHandler(w http.ResponseWriter, r * http.Request){
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
}



func idManager(w http.ResponseWriter){
	ids = append(ids, strconv.Itoa(lastId))
    lastId++
    
    idsJson, _ := json.MarshalIndent(ids, "", "    ")
    fmt.Fprintln(w, string(idsJson))
    
}




func main(){
	igcs = make(map[int]string)
	http.HandleFunc("/api", metaHandler);
	http.HandleFunc("/api/igc", inputHandler);
	//http.HandleFunc("/api/igc/<id>", inputHandler);
	//http.ListenAndServe(":8080", nil);
	appengine.Main()


}






































