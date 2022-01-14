package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

// Record defines general CMSAMQProxy record
type Record map[string]interface{}

// GzipReader struct to handle GZip'ed content of HTTP requests
type GzipReader struct {
	*gzip.Reader
	io.Closer
}

// Close function closes gzip reader
func (gz GzipReader) Close() error {
	return gz.Closer.Close()
}

// StatusHandler handles all CMSAMQProxy requests
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(info()))
		return
	}
	var headers []interface{}
	headers = append(headers, r.Method)
	headers = append(headers, r.URL.Path)
	headers = append(headers, r.RemoteAddr)
	log.Println(headers...)
	w.WriteHeader(http.StatusOK)
}

// HttpRequestHandler handles all CMSAMQProxy requests
func HttpRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL, r.Proto, r.Host, r.RemoteAddr, r.Header)
	if r.Method == "GET" {
		// print out all request headers
		fmt.Fprintf(w, "%s %s %s \n", r.Method, r.URL, r.Proto)
		for k, v := range r.Header {
			h := strings.ToLower(k)
			if strings.Contains(h, "hmac") || strings.Contains(h, "cookie") {
				continue
			}
			fmt.Fprintf(w, "Header field %q, Value %q\n", k, v)
		}
		fmt.Fprintf(w, "Host = %q\n", r.Host)
		fmt.Fprintf(w, "RemoteAddr= %q\n", r.RemoteAddr)
		fmt.Fprintf(w, "\n\nFinding value of \"Accept\" %q\n", r.Header["Accept"])

		page := "Hello from Go\n"
		w.Write([]byte(page))
	} else {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Fprint(w, err.Error())
		} else {
			fmt.Fprint(w, string(requestDump))
		}
	}
}

// DataHandler handles all CMSAMQProxy requests
func DataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(info()))
		return
	}
	authStatus := CMSAuth.CheckAuthnAuthz(r.Header)
	if !authStatus {
		log.Printf("fail to auth request %+v", r.Header)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// check CMS role/group (or site)
	if Config.CMSRole != "" && Config.CMSGroup != "" {
		authzStatus := CMSAuth.CheckCMSAuthz(r.Header, Config.CMSRole, Config.CMSGroup, Config.CMSSite)
		if !authzStatus {
			log.Printf("fail to authorize request %+v, cms role=%s group=%s site=%s status=%v", r.Header, Config.CMSRole, Config.CMSGroup, Config.CMSSite, authzStatus)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
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
	body := r.Body
	// handle gzip content encoding
	if r.Header.Get("Content-Encoding") == "gzip" {
		r.Header.Del("Content-Length")
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Println("unable to get gzip reader", err)
			return out, err
		}
		body = GzipReader{reader, r.Body}
	}
	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println("Unable to read request body", err)
	}

	err = json.Unmarshal(data, &records)
	if err != nil {
		if Config.Verbose > 0 {
			log.Printf("Unable to decode input request, error %v, request %+v\n%+v\n", err, r, string(data))
		} else {
			log.Printf("Unable to decode input request, error %v\n", err)
		}
		return out, err
	}
	// send data with this stomp connection
	var ids []string
	for _, rec := range records {
		uid := genUUID()
		rrr := make(Record)
		rrr["data"] = rec
		producer := Config.Producer
		metadata := make(Record)
		//metadata["timestamp"] = time.Now().Unix() * 1000
		metadata["producer"] = producer
		metadata["_id"] = uid
		metadata["uuid"] = uid
		rrr["metadata"] = metadata
		data, err := json.Marshal(rrr)
		if err != nil {
			if Config.Verbose > 0 {
				log.Printf("Unable to marshal, error: %v, data: %+v\n", err, rrr)
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
