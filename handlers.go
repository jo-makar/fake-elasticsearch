package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Nested, anonymous structs are not easy to deal with hence just returning hard-coded literals for the time being.
// TODO Are there better alternatives that support nested, anonymous struct literals elegantly?
//
// If later fields require runtime modification, can always unmarshal a string to interface map and re-marshal:
//
//	var respMap map[string]interface{}
//	if json.Unmarshal([]byte(resp), &respMap) != nil { ... }
//	if v, ok := respMap["foo"]; ok {
//		s, ok2 := v.(string)
//		if ok2 { ... } else { ... }
//	} else { ...} 
//	if resp2, err := json.Marshal(respMap); err != nil { ... }

type RootHandler struct { *State }
type BulkHandler struct { *State }
type PipelineHandler struct { *State }
type TemplateHandler struct { *State }
type XpackHandler struct { *State }

func write(b []byte, w *http.ResponseWriter, r *http.Request) {
	if _, err := (*w).Write(b); err != nil {
		log.Printf("%s: %s: unable to write response: %s\n", r.RequestURI, GetIp(r), err)
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/_bulk") {
		(&BulkHandler{h.State}).ServeHTTP(w, r)
		return
	}

	if r.URL.Path != "/" {
		log.Printf("%s: %s: not found\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		log.Printf("%s: %s: unsupported method (%s)\n", r.RequestURI, GetIp(r), r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resp := `
	{
	  "name" : "fake-node-1",
	  "cluster_name" : "fake-cluster",
	  "cluster_uuid" : "SNDZ4jXx8pg3iEbakRdB",
	  "version" : {
	    "number" : "7.3.1",
	    "build_flavor" : "default",
	    "build_type" : "rpm",
	    "build_hash" : "de777fa",
	    "build_date" : "2019-07-24T18:30:11.767338Z",
	    "build_snapshot" : false,
	    "lucene_version" : "8.1.0",
	    "minimum_wire_compatibility_version" : "6.8.0",
	    "minimum_index_compatibility_version" : "6.0.0-beta1"
	  },
	  "tagline" : "You Know, for Search"
	}

	`

	w.WriteHeader(http.StatusOK)
	write([]byte(resp), &w, r)
}

func (h *XpackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/_xpack" {
		log.Printf("%s: %s: not found\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		log.Printf("%s: %s: unsupported method (%s)\n", r.RequestURI, GetIp(r), r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Index Lifecycle Management (ILM) is disabled to avoid requiring support for /_ilm/policy/ endpoints

	resp := `
	{
	  "build" : {
	    "hash" : "de777fa",
	    "date" : "2019-07-24T18:30:11.767338Z"
	  },
	  "license" : {
	    "uid" : "9ff14a29-65b3-4c8b-bbc2-bf06ecdeb52b",
	    "type" : "basic",
	    "mode" : "basic",
	    "status" : "active"
	  },
	  "features" : {
	            "ccr" : { "available" : false, "enabled" : true },
	     "data_frame" : { "available" : true,  "enabled" : true },
	      "flattened" : { "available" : true,  "enabled" : true },
	          "graph" : { "available" : false, "enabled" : true },
	            "ilm" : { "available" : true,  "enabled" : false },
	       "logstash" : { "available" : false, "enabled" : true },
	             "ml" : { "available" : false, "enabled" : true },
	     "monitoring" : { "available" : true,  "enabled" : true },
	         "rollup" : { "available" : true,  "enabled" : true },
	       "security" : { "available" : true,  "enabled" : false },
	            "sql" : { "available" : true,  "enabled" : true },
	        "vectors" : { "available" : true,  "enabled" : true },
	    "voting_only" : { "available" : true,  "enabled" : true },
	        "watcher" : { "available" : false, "enabled" : true }
	  }
	}
	`

	w.WriteHeader(http.StatusOK)
	write([]byte(resp), &w, r)
}

func (h *PipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/_ingest/pipeline/") {
		log.Printf("%s: %s: not found\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		log.Printf("%s: %s: unsupported method (%s)\n", r.RequestURI, GetIp(r), r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	t := strings.Split(r.URL.Path, "/")
	if len(t) != 4 {
		log.Printf("%s: %s: invalid pipeline id\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := t[3]

	if r.Method == http.MethodGet {
		if v, ok := h.pipelines[id]; ok {
			w.WriteHeader(http.StatusOK)
			write([]byte(v), &w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else /* MethodPut */ {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("%s: %s: unable to read request: %s\n", r.RequestURI, GetIp(r), err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		h.pipelines[id] = string(b)
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/_template/") {
		log.Printf("%s: %s: not found\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPut {
		log.Printf("%s: %s: unsupported method (%s)\n", r.RequestURI, GetIp(r), r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	t := strings.Split(r.URL.Path, "/")
	if len(t) != 3 {
		log.Printf("%s: %s: invalid template id\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := t[2]

	if r.Method == http.MethodHead || r.Method == http.MethodGet {
		if v, ok := h.templates[id]; ok {
			w.WriteHeader(http.StatusOK)
			if r.Method == http.MethodGet {
				write([]byte(v), &w, r)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else /* MethodPut */ {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("%s: %s: unable to read request: %s\n", r.RequestURI, GetIp(r), err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		h.templates[id] = string(b)
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *BulkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/_bulk") {
		log.Printf("%s: %s: not found\n", r.RequestURI, GetIp(r))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("%s: %s: unsupported method (%s)\n", r.RequestURI, GetIp(r), r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get the index from the url if present
	var index string
	if r.URL.Path != "/_bulk" {
		m := regexp.MustCompile(`^/([^/]+)/_bulk$`).FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Printf("%s: %s: invalid index name\n", r.RequestURI, GetIp(r))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		index = m[1]
	}


	// Only supporting a subset of response fields (TODO How to know which ones the beat agents care about?)
	// Thankfully it seems most/all? beat agents are lenient and will use the http status only.

	items := make([]string, 0)

	// TODO How to structure the following code more clearly?  Need an easy way to bail on a bad line/iteration.
	//      Consider this is a temporary first draft until a better approach is worked out.
	//      Need careful though on handling errors as well, eg if the action can't be unmarshalled how to respond?

	consecErrors := 0
	iterate := true

	handleError := func(i string) {
		consecErrors++
		if consecErrors == 5 {
			log.Printf("%s: %s: too many consecutive errors\n", r.RequestURI, GetIp(r))
			w.WriteHeader(http.StatusPartialContent)
			iterate = false
		}

		if i != "" {
			items = append(items, i)
		}
	}

	scanner := bufio.NewScanner(r.Body)
	for iterate && scanner.Scan() {
		actionJson := scanner.Text()

		var action map[string]interface{}
		err := json.Unmarshal([]byte(actionJson), &action)
		if err != nil {
			log.Printf("%s: %s: unable to unmarshal json: %s\n", r.RequestURI, GetIp(r), err)
			handleError("")
			continue
		}

		// Use the index specified in the action metadata if present
		metadataIndex := func(v interface{}) (string, error) {
			if m, ok := v.(map[string]interface{}); ok {
				if i, ok := m["_index"]; ok {
					if i2, ok := i.(string); ok {
						return i2, nil
					}
				}
				return "", nil
			} else {
				return "", errors.New(fmt.Sprintf("%s: %s: unexpected action format", r.RequestURI, GetIp(r)))
			}
		}

		if v, ok := action["index"]; ok {
			index2 := index
			if i, err := metadataIndex(v); err != nil {
				log.Printf("%s: %s: unexpected action format\n", r.RequestURI, GetIp(r))
				handleError(`{"index":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			} else if i != "" {
				index2 = i
			}

			if !scanner.Scan() {
				break
			}
			recordJson := scanner.Text()

			var record map[string]interface{}
			err := json.Unmarshal([]byte(recordJson), &record)
			if err != nil {
				log.Printf("%s: %s: unable to unmarshal json: %s\n", r.RequestURI, GetIp(r), err)
				handleError(`{"index":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			}

			// *** Here index record(Json) to index2 as appropriate ***

			items = append(items, fmt.Sprintf(`{"index":{"_index":"%s","_type":"_doc","result":"created","status":200}}`, index2))

		} else if v, ok := action["delete"]; ok {
			index2 := index
			if i, err := metadataIndex(v); err != nil {
				log.Printf("%s: %s: unexpected action format\n", r.RequestURI, GetIp(r))
				handleError(`{"delete":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			} else if i != "" {
				index2 = i
			}

			// *** Here delete index2 as appropriate ***

			items = append(items, fmt.Sprintf(`{"delete":{"_index":"%s","_type":"_doc","result":"deleted","status":200}}`, index2))

		} else if v, ok := action["create"]; ok {
			index2 := index
			if i, err := metadataIndex(v); err != nil {
				log.Printf("%s: %s: unexpected action format\n", r.RequestURI, GetIp(r))
				handleError(`{"create":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			} else if i != "" {
				index2 = i
			}

			if !scanner.Scan() {
				break
			}
			recordJson := scanner.Text()

			var record map[string]interface{}
			err := json.Unmarshal([]byte(recordJson), &record)
			if err != nil {
				log.Printf("%s: %s: unable to unmarshal json: %s\n", r.RequestURI, GetIp(r), err)
				handleError(`{"create":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			}

			// *** Here create record(Json) in index2 as appropriate ***
			// Create does the same thing as index but should fail if the document already exists.

			items = append(items, fmt.Sprintf(`{"create":{"_index":"%s","_type":"_doc","result":"created","status":201}}`, index2))

		} else if v, ok := action["update"]; ok {
			index2 := index
			if i, err := metadataIndex(v); err != nil {
				log.Printf("%s: %s: unexpected action format\n", r.RequestURI, GetIp(r))
				handleError(`{"update":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			} else if i != "" {
				index2 = i
			}

			if !scanner.Scan() {
				break
			}
			recordJson := scanner.Text()

			var record map[string]interface{}
			err := json.Unmarshal([]byte(recordJson), &record)
			if err != nil {
				log.Printf("%s: %s: unable to unmarshal json: %s\n", r.RequestURI, GetIp(r), err)
				handleError(`{"update":{"_index":"","_type":"_doc","result":"failed","status":400}`)
				continue
			}

			// *** Here upsert record(Json)["doc"] in index2 as appropriate ***

			items = append(items, fmt.Sprintf(`{"update":{"_index":"%s","_type":"_doc","result":"updated","status":201}}`, index2))

		} else {
			log.Printf("%s: %s: missing or unexpected action\n", r.RequestURI, GetIp(r), err)
			handleError("")
			continue
		}

		consecErrors = 0
	}
	if err := scanner.Err(); err != nil {
		log.Printf("%s: %s: unable to read request: %s\n", r.RequestURI, GetIp(r), err)
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	var resp bytes.Buffer
	resp.WriteString(`{"errors":false,"items":[`)
	for i, v := range items {
		if i > 0 {
			resp.WriteString(",")
		}
		resp.WriteString(v)
	}
	resp.WriteString(`]}`)

	write(resp.Bytes(), &w, r)
}
