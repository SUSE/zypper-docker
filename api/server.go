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
	"fmt"
	"log"
	"net/http"

	"github.com/SUSE/zypper-docker/backend"
	"github.com/SUSE/zypper-docker/backend/drivers"
	"github.com/SUSE/zypper-docker/logger"
	"github.com/SUSE/zypper-docker/utils"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

// Returns the "Not found" response.
func notFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//fmt.Fprint(w, lib.Response{Error: "Something wrong happened!"})
}

func route() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")
	return r
}

// Serve starts the API server by considering the given CLI context.
func Serve(ctx *cli.Context) {
	// Initialize the logger and the drivers for this command.
	logger.Initialize("apiserver", true)
	backend.Initialize(true)
	drivers.Initialize(ctx)

	// Routing.
	n := negroni.Classic()
	r := route()
	n.UseHandler(r)

	port := fmt.Sprintf(":%v", ctx.Int("port"))

	// Try to run on HTTPS.
	cert := ctx.String("cert-path")
	key := ctx.String("key-path")
	if cert != "" && key != "" {
		log.Printf("Running on port %s", port)
		if err := http.ListenAndServeTLS(port, cert, key, n); err != nil {
			logger.Fatalf("Could not start server: %v", err)
		}
		utils.ExitWithCode(0)
		return
	}

	// Falling back to normal HTTP.
	logger.Printf("Running with HTTP: use --cert-path and --key-path for running in HTTPS")
	n.Run(port)
}
