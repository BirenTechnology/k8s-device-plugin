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
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BirenTechnology/k8s-device-plugin/pkg/brgpu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Options struct {
	mode                  string
	pluginMountPath       string
	pulse                 int
	initModeTolerateLevel int
	mountAllDevice        bool
	mountDriDevice        bool
	runtime               string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.pulse, "pulse", o.pulse, "heart beating every seconds")
	fs.StringVar(&o.runtime, "container-runtime", o.runtime, "the container runtime;runc or kata, default is runc")
	fs.BoolVar(&brgpu.CdiFeature, "cdi-feature", brgpu.CdiFeature, "enable cdi feature")
	fs.BoolVar(&brgpu.OverwriteCdiConfig, "overwrite-cdi-config", brgpu.OverwriteCdiConfig, "overwrite cdi config")
	fs.BoolVar(&brgpu.MountHostPath, "mount-host-path", brgpu.MountHostPath, "mount lib and bin folder in host to container, default is false")
}

func (o *Options) Run() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var gpuConfig brgpu.GPUConfig
	bgm := brgpu.NewBrGPUManager(o.pluginMountPath, gpuConfig)

	go func() {
		sig := <-sigs
		log.Infof("Get the signal %s", sig)
		bgm.Stop <- true
	}()

	bgm.Serve(o.pulse, o.mountAllDevice, o.mountDriDevice, o.runtime)
	return nil
}

func NewManagerCommand() *cobra.Command {
	opts := NewOptions()

	cmd := &cobra.Command{
		Use:  "br-gpu-device-plugin",
		Long: "Biren gpu device plugin",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}
