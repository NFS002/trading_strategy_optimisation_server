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
	File_Name string `json:"file_name"`
	Parameter_Values string `json:"parameter_values"`
	Candlestick_Interval string `json:"candlestick_interval"`
	Other string `json:"other"`
	Commit_Hash string `json:"commit_hash"`
	Symbol string `json:"symbol"`
	Perc_Profitable string `json:"perc_profitable"`
	N_Trades string `json:"n_trades"`
	Sharpe_Ratio string `json:"sharpe_ratio"`
}


// A fundamental concept in `net/http` servers is
// *handlers*. A handler is an object implementing the
// `http.Handler` interface. A common way to write
// a handler is by using the `http.HandlerFunc` adapter
// on functions with the appropriate signature.
func submit(w http.ResponseWriter, req *http.Request) {
	path := "strategy_comparison.csv"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(w, "Error opening %s: %v", path, err)
	} else {
		writer := csv.NewWriter(file)
		decoder := json.NewDecoder(req.Body)
    	var req_body json_req_body
    	json_decode_err := decoder.Decode(&req_body)
   		if json_decode_err != nil {
			fmt.Fprintf(w, "Json decode error: %v", err)
    	} else {
			csv_records := []string{req_body.File_Name, req_body.Parameter_Values, req_body.Candlestick_Interval, req_body.Other, 
				req_body.Commit_Hash, req_body.Symbol, req_body.Perc_Profitable, req_body.N_Trades, req_body.Sharpe_Ratio}
			writer.Write(csv_records) 
			writer.Flush()
			if err := writer.Error(); err != nil {
				fmt.Fprintf(w, "IO error: %v", err)
			} else {
				fmt.Fprintf(w, "Success: %s", req_body.Symbol)
			}
		}
	}
}

func main() {
	http.HandleFunc("/submit", submit)

	http.ListenAndServe(":8090", nil)
}

