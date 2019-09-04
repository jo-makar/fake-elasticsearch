package main

import (
	"io/ioutil"
	"log"
	"net/http"
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
type PipelineHandler struct { *State }
type XpackHandler struct { *State }

func write(b []byte, w *http.ResponseWriter, r *http.Request) {
	if _, err := (*w).Write(b); err != nil {
		log.Printf("%s: %s: unable to write response: %s\n", r.RequestURI, GetIp(r), err)
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
