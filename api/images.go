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
	"github.com/SUSE/zypper-docker/backend/drivers"
	"github.com/SUSE/zypper-docker/logger"
	"github.com/mssola/capture"
)

func errorResponse(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := drivers.Updates{Error: msg}
	js, _ := json.Marshal(&resp)
	w.Write(js)
}

func parseOutput(output []byte) drivers.Updates {
	up := drivers.Current().ParseUpdateOutput(output)
	if up.Error != "" {
		logger.Printf("could not parse output: %v", up.Error)
		up.Error = "something went wrong"
	}
	return up
}

func evaluateImage(w http.ResponseWriter, image string) {
	var err error

	logger.Printf("evaluating image: %v", image)

	res := capture.All(func() {
		err = backend.ListUpdates(backend.Security, image, true)
	})

	if err != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
	} else if res.Error != nil {
		logger.Printf("%v", res.Error)
		errorResponse(w, http.StatusInternalServerError, "something went wrong...")
	} else {
		resp := parseOutput(res.Stdout)

		js, _ := json.Marshal(&resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(js)
	}
}

// GET /images?image=<img>
func imagesShow(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	if val, ok := query["image"]; ok {
		if len(val) != 1 {
			errorResponse(w, http.StatusNotFound, "expecting an 'image' query parameter")
			return
		}
		evaluateImage(w, val[0])
	} else {
		errorResponse(w, http.StatusNotFound, "expecting an 'image' query parameter")
	}
}
