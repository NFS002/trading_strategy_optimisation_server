// Writing a basic HTTP server is easy using the
// `net/http` package.
package main

import (
	"os"
	"fmt"
	"net/http"
	"encoding/csv"
	"encoding/json"
)

type json_req_body struct {
	file_name string
	parameter_values string
	candlestick_interval string
	other string
	commit_hash string
	symbol string 
	perc_profitable string
	n_trades string
	sharpe_ratio string
}

// A fundamental concept in `net/http` servers is
// *handlers*. A handler is an object implementing the
// `http.Handler` interface. A common way to write
// a handler is by using the `http.HandlerFunc` adapter
// on functions with the appropriate signature.
func submit(w http.ResponseWriter, req *http.Request) {
	path := "strategy_comparison.csv"
	file, err := os.Open(path)
	if err != nil {
		//flash("Error opening $path ")
		fmt.Fprintf(w, "Path Error")
	} else {
		writer := csv.NewWriter(file)
		decoder := json.NewDecoder(req.Body)
    	var req_body json_req_body
    	json_decode_err := decoder.Decode(&req_body)
   		if json_decode_err != nil {
        	//flash("JSON decode error")
			fmt.Fprintf(w, "Json Error: %v", err)
    	} else {
			csv_records := []string{"let", "a", "n****", "try", "me"}
			writer.Write(csv_records) // calls Flush internally
			if err := writer.Error(); err != nil {
				fmt.Fprintf(w, "CSV write error: %v", err)
			} else {
				fmt.Fprintf(w, "Success: %s", req_body.symbol)
			}
		}
	}
}

func main() {
	http.HandleFunc("/submit",submit)

	http.ListenAndServe(":8090", nil)
}

