package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Record defines general WMArchive record
type Record map[string]interface{}

// DataHandler handles all WMArchive requests
func DataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(info()))
		return
	}
	out, err := processRequest(r)
	if err != nil {
		log.Println(r.Method, r.URL.Path, r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data, err := json.Marshal(out)
	var headers []interface{}
	headers = append(headers, r.Method)
	headers = append(headers, r.URL.Path)
	headers = append(headers, r.RemoteAddr)
	for _, h := range []string{"User-Agent", "Cms-Authn-Dn", "X-Forwarded-For"} {
		if v, ok := r.Header[h]; ok {
			headers = append(headers, v)
		}
	}
	if err == nil {
		if Config.Verbose > 0 {
			headers = append(headers, string(data))
		} else {
		}
		log.Println(headers...)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	headers = append(headers, err)
	log.Println(headers...)
	w.WriteHeader(http.StatusInternalServerError)
}

// helper function to process http request with list of records
func processRequest(r *http.Request) ([]Record, error) {
	var out, records []Record
	defer r.Body.Close()
	// it is better to read whole body instead of using json decoder
	//     err := json.NewDecoder(r.Body).Decode(&rec)
	// since we can print body later for debugging purposes
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read request body", err)
	}
	err = json.Unmarshal(body, &records)
	if err != nil {
		if Config.Verbose > 0 {
			log.Printf("Unable to decode input request, error %v, request %+v\n%+v\n", err, r, string(body))
		} else {
			log.Printf("Unable to decode input request, error %v\n", err)
		}
		return out, err
	}
	// send data with this stomp connection
	var ids []string
	for _, rec := range records {
		uid := genUUID()
		producer := Config.Producer
		metadata := make(Record)
		//metadata["timestamp"] = time.Now().Unix() * 1000
		metadata["producer"] = producer
		metadata["_id"] = uid
		metadata["uuid"] = uid
		rec["metadata"] = metadata
		data, err := json.Marshal(rec)
		if err != nil {
			if Config.Verbose > 0 {
				log.Printf("Unable to marshal, error: %v, data: %+v\n", err, rec)
			} else {
				log.Printf("Unable to marshal, error: %v, data\n", err)
			}
			continue
		}

		// dump message to our log
		if Config.Verbose > 1 {
			log.Println("New record", string(data))
		}

		// send data to Stomp endpoint
		if Config.Endpoint != "" {
			err := stompMgr.Send(data)
			if err == nil {
				ids = append(ids, uid)
			} else {
				// get new stomp Manager
				initStompManager()
				record := make(Record)
				record["status"] = "fail"
				record["reason"] = fmt.Sprintf("Unable to send data to MONIT, error: %v", err)
				record["ids"] = ids
				out = append(out, record)
				return out, err
			}
		} else {
			ids = append(ids, uid)
		}

	}
	// prepare output wmarhchive response record
	record := make(Record)
	if len(ids) > 0 {
		record["status"] = "ok"
		record["ids"] = ids
	} else {
		record["status"] = "empty"
		record["ids"] = ids
		record["reason"] = "no input data is provided"
	}
	out = append(out, record)
	return out, nil
}
