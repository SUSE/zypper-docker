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

// This is a thin wrapper on top of zypper that allows patching docker images
// in a safe way.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Listen to all signals, propagates SIGINT, SIGTSTP and SIGTERM
// to the killChannel channel.
func listenSignals() {
	killChannel = make(chan bool)
	c := make(chan os.Signal)
	signal.Notify(c)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTSTP,
				syscall.SIGTERM:
				log.Printf("Signal '%v' received: shutting down gracefully.", sig)
				killChannel <- true
			default:
				log.Printf("Signal '%v' not handled. Doing nothing...", sig)
			}
		}
	}()
}
