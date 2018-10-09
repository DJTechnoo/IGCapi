package main

import (


	"fmt"
	"encoding/json"
	"net/http"

)


type Meta struct {
	Info string `json:"info"`
	Version string `json:"version"`
}



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
        fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
        link := r.FormValue("link")
   
        fmt.Fprintf(w, "Name = %s\n", link)
    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
    }
}




func main(){

	http.HandleFunc("/api", metaHandler);
	http.HandleFunc("/api/igc", inputHandler);
	http.ListenAndServe(":8080", nil);


}






































