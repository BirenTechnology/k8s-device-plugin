// Copyright 2024 Shanghai Biren Technology Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/BirenTechnology/k8s-device-plugin/cmd/app"
	log "github.com/sirupsen/logrus"
)

func init() {
	formatter := &log.TextFormatter{}
	formatter.DisableQuote = true
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
}

var (
	Version string
	Commit  string
	Time    string
)

func main() {
	log.Info("Biren GPU Device Plugin Start")
	log.Infof("Version: %s;Commit: %s; Build At: %s", Version, Commit, Time)
	rand.Seed(time.Now().UnixNano())

	command := app.NewManagerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
