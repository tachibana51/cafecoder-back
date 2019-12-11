package api

import (
	"fmt"
	"net/http"
	"encoding/json"
	"../values"
	"net"
	"strings"
)
type postCode struct {
	codeSession	string `json:"code_session"`
	userId		string `json:"user_id"`
	language	string `json:"language"`
	contestId	string `json:"contest_id"`
	problem		string `json:"problem"`
}

type testcase struct {
	caseName	string `json:"testcase_name"`
	result		string `json:"result"`
	runtime 	int		`json:"runtime"`
}

type getResult struct {
	codeSession	string `json:"code_session"`
	code 		string `json:"code"`
	result 		string `json:"result"`	
	testcases	[]testcase `json:"testcases"`
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	switch (r.Method) {
		case "GET":

		default:
			return
	}

}

func codeHandler(w http.ResponseWriter, r *http.Request) {
	switch (r.Method) {
		case "GET":

		case "POST":
			body := make([]byte, 4096)
			jsonData := postCode{}
			_, err := r.Body.Read(body)
			//read json
			err = json.Unmarshal(body, &jsonData)
			if err != nil {
				fmt.Println(err)
				return
			}
			//con job_order
			con, err := net.Dial("tcp", values.QUE_HOST_PORT)
			if err != nil {
				fmt.Println(err)
				return
			}
			argStr := []string{jsonData.codeSession, "path", jsonData.language, "testcasedir", "point"}
			con.Write([]byte(strings.Join(argStr ,",")))
			con.Close()
			fmt.Fprintf(w, "OK")

		default:
			return
	}
} 

func evenvListenerThread() {
	http.HandleFunc("/api/v1/result", resultHandler)
	http.HandleFunc("/api/v1/code", codeHandler)
	http.ListenAndServe(":8080", nil)
}