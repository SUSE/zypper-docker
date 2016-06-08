// Copyright (c) 2015 SUSE LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/SUSE/zypper-docker/backend"
	"github.com/SUSE/zypper-docker/logger"
	"github.com/mssola/capture"
)

type imagesShowResponse struct {
	Updates  bool
	Security bool
	Log      string
	Error    string
}

func errorResponse(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := imagesShowResponse{Error: msg}
	js, _ := json.Marshal(&resp)
	w.Write(js)
}

func imagesShow(w http.ResponseWriter, req *http.Request) {
	var err error
	var image string
	resp := imagesShowResponse{Updates: false, Security: false}

	// Evaluating the given parameter.
	query := req.URL.Query()
	if val, ok := query["image"]; ok {
		if len(val) != 1 {
			errorResponse(w, http.StatusNotFound, "expecting an 'image' query parameter")
			return
		}
		image = val[0]
	} else {
		errorResponse(w, http.StatusNotFound, "expecting an 'image' query parameter")
		return
	}

	logger.Printf("evaluating image: %v", image)
	res := capture.All(func() {
		resp.Updates, resp.Security, err = backend.HasPatches(image)
	})
	if err != nil || res.Error != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	resp.Log = string(res.Stdout)
	js, _ := json.Marshal(&resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
