package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/edgelesssys/constellation/mount/cryptmapper"
	"github.com/edgelesssys/constellation/mount/kms"
	"k8s.io/klog"
)

var (
	close     = flag.Bool("c", false, "close the crypt device")
	integrity = flag.Bool("integrity", false, "format the device with dm-integrity")
	source    = flag.String("source", "", "source volume")
	volumeID  = flag.String("target", "new_crypt_device", "mapped target")
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()
	flag.Parse()

	mapper := cryptmapper.New(kms.NewStaticKMS(), "", &cryptmapper.CryptDevice{})

	if *close {
		err := mapper.CloseCryptDevice(*volumeID)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if *source == "" {
			log.Fatal("missing require flag \"-source\"")
		}
		out, err := mapper.OpenCryptDevice(context.Background(), *source, *volumeID, *integrity)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Crypt device activate as: %q\n", out)
	}
}
