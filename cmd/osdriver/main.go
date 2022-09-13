package main

import (
	"flag"
	"os"

	"github.com/fimreal/goutils/ezap"
	"github.com/fimreal/tencent-cos-csi-driver/pkg/driver"
)

var (
	endpoint = flag.String("endpoint", "unix://csi/csi.sock", "CSI endpoint")
	nodeID   = flag.String("nodeid", "", "node id")
)

func main() {
	flag.Parse()
	if *nodeID == "" {
		ezap.Fatal("nodeID is empty")
	}

	driver, err := driver.NewDriver(*endpoint, *nodeID)
	if err != nil {
		ezap.Fatal(err)
	}
	driver.Start()
	os.Exit(0)
}
