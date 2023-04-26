package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

/*
Algorithm ->

1. Open the log file.
2. Seek to the EOF.
3. Read the latest data and print it in the console.
4. Close the file if program terminates or ends.
*/

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/log", logHandler).Methods(http.MethodGet)
	http.ListenAndServe("localhost:8080", router)
}

func logHandler(w http.ResponseWriter, req *http.Request) {

	offset, err := strconv.Atoi(req.Header.Get("offset"))
	w.Header().Set("offset", strconv.Itoa(offset))

	if err != nil {
		fmt.Println("Invalid offset sent.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response, offset, err := watchLog("log", offset)
	if err != nil {
		fmt.Println("Error getting logs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("offset", strconv.Itoa(offset))
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		log.Fatalln("There was an error encoding the fetched logs")
	}
	return
}

func watchLog(fileName string, offset int) ([]string, int, error) {

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error in opening file.")
		return nil, offset, err
	}

	defer file.Close()
	file.Seek(int64(offset), 2)

	response := []string{}

	for start := time.Now(); time.Now().Sub(start) < 5*time.Second; {
		data := make([]byte, 1024)
		_, err := file.Read(data)
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Println("Error in reading the file contents.")
			return nil, offset, err
		}
		fmt.Println(string(data))
		response = append(response, filterData(string(data))...)
		offset += 1
	}
	return response, offset, nil
}

func filterData(data string) []string {

	filteredData := []string{}
	splittedString := strings.Split(data, "\n")
	for _, string := range splittedString {
		if string == "" {
			continue
		}
		if string == "\u0000" {
			string = strings.Split(string, "\u0000")[0]
		}
		filteredData = append(filteredData, string)
	}
	return filteredData[:len(filteredData)-1]
}
