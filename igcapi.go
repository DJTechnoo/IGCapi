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




func main(){

	http.HandleFunc("/api/meta/", metaHandler);
	http.ListenAndServe(":8080", nil);


}






































