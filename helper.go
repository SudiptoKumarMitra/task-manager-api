package main
import (
	"net/http"
	"strconv"
)
func getID(r *http.Request) (int,error){
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	return id,err
}

func headerSet(w http.ResponseWriter){
	w.Header().Set("Content-Type", "application/json")
}
