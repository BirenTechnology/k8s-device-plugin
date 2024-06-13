// Copyright 2024 Shanghai Biren Technology Co., Ltd.
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package utils

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	K8s *kubernetes.Clientset
}

func NewClient(inCluster bool) (Client, error) {
	var config *rest.Config
	var err error
	switch inCluster {
	case false:
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	case true:
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return Client{}, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Client{}, err
	}
	return Client{
		K8s: clientset,
	}, nil
}

func InCluster() bool {
	ic := false
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		ic = true
	}
	return ic
}
