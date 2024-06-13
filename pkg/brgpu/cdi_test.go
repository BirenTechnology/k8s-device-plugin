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
package brgpu

import (
	"testing"

	"github.com/BirenTechnology/go-brml/brml"
	log "github.com/sirupsen/logrus"
)

func TestGenerateFile(t *testing.T) {
	err := brml.Init()
	if err != nil {
		log.Error(err)
	}
	defer brml.Shutdown()

	err = generateConfigCdiFile(RuntimeRunc)
	if err != nil {
		t.Error(err)
	}
}
