package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type RootHandler struct {
	*State
}

// TODO Where is this documented?
//      Ideally would like to support the minimum required

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

	type Version struct {
		Number            string `json:"number"`
		BuildFlavor       string `json:"build_flavor"`
		BuildType         string `json:"build_type"`
		BuildHash         string `json:"build_hash"`
		BuildDate         string `json:"build_date"`
		BuildSnapshot     bool   `json:"build_snapshot"`
		LuceneVersion     string `json:"lucene_version"`
		MinWireCompatVer  string `json:"minimum_wire_compatibility_version"`
		MinIndexCompatVer string `json:"minimum_index_compatibility_version"`
	}

	type Resp struct {
		Name        string  `json:"name"`
		ClusterName string  `json:"cluster_name"`
		ClusterUuid string  `json:"cluster_uuid"`
		Version     Version `json:"version"`
		Tagline     string  `json:"tagline"`
	}

	resp := Resp{
		       Name: h.NodeName,
		ClusterName: h.ClusterName,
		ClusterUuid: h.ClusterUuid,
		    Version: Version{
			           Number: "7.3.0",
			      BuildFlavor: "default",
			        BuildType: "rpm",
			        BuildHash: "de777fa",
			        BuildDate: "2019-07-24T18:30:11.767338Z",
			    BuildSnapshot: false,
			    LuceneVersion: "8.1.0",
			 MinWireCompatVer: "6.8.0",
			MinIndexCompatVer: "6.0.0-beta1",
		},
		    Tagline: "You Know, for Search",
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("%s: %s: error serializing response: %s\n", r.RequestURI, GetIp(r), err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(respBytes); err != nil {
		log.Printf("%s: %s: unable to write response: %s\n", r.RequestURI, GetIp(r), err)
	}
}
