package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BirenTechnology/go-brml/brml"
	"github.com/BirenTechnology/k8s-device-plugin/pkg/brgpu"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := brml.Init()
	if err != nil {
		log.Errorf("brml init failed %v", err)
	}
	defer brml.Shutdown()

	cardsFiles, err := os.ReadDir("/dev/biren")
	if err != nil {
		log.Errorf("read dir /dev/biren failed %v", err)
		panic(err)
	}
	cards := []string{}
	for _, f := range cardsFiles {
		if strings.Contains(f.Name(), "card_") {
			cards = append(cards, f.Name())
		}
	}
	devices, err := brgpu.DeviceDiscover()
	if err != nil {
		log.Errorf("discover devices failed %v", err)
		panic(err)
	}
	log.Info("discover devices:")
	for _, d := range devices {
		fmt.Println(d.PhysicalNum, d.Instances)
	}

	_, err = brgpu.Device2Graph(cards)
	if err != nil {
		log.Errorf("device %v to graph failed %v", cards, err)
		panic(err)
	}

	log.Info("/dev/biren/card_x -> gpu hw:")
	for _, c := range cards {
		gpu_id, err := os.ReadFile(fmt.Sprintf("/sys/class/biren/%s/device/physical_id", c))
		if err != nil {
			log.Errorf("read sys/class/biren/%s/device/physical_id failed %v", c, err)
		} else {
			fmt.Printf("%s -> %v", c, string(gpu_id))
		}
	}

	log.Info("brsmi gpu list:")
	cmd := exec.Command("brsmi", "gpu", "list")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Errorf("exec `brsmi gpu list` failed %v", err)
	}

	log.Info("brsmi gpu topo:")
	cmd = exec.Command("brsmi", "topo", "-m")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Errorf("exec `brsmi gpu list` failed %v", err)
	}
}
