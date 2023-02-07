/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgelesssys/constellation/v2/hack/image-measurement/server"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"libvirt.org/go/libvirt"
)

// Usage:
// go build
//./image-measurement --path=disk.raw --type=raw

type libvirtInstance struct {
	conn      *libvirt.Connect
	log       *logger.Logger
	imagePath string
}

func (l *libvirtInstance) uploadBaseImage(baseVolume *libvirt.StorageVol) (err error) {
	stream, err := l.conn.NewStream(libvirt.STREAM_NONBLOCK)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Free() }()
	file, err := os.Open(l.imagePath)
	if err != nil {
		return fmt.Errorf("error while opening %s: %s", l.imagePath, err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	fi, err := file.Stat()
	if err != nil {
		return err
	}
	if err := baseVolume.Upload(stream, 0, uint64(fi.Size()), 0); err != nil {
		return err
	}
	transferredBytes := 0
	buffer := make([]byte, 4*1024*1024)
	for {
		_, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		num, err := stream.Send(buffer)
		if err != nil {
			return err
		}
		transferredBytes += num

	}
	if transferredBytes < int(fi.Size()) {
		return fmt.Errorf("only send %d out of %d bytes", transferredBytes, fi.Size())
	}
	return nil
}

func (l *libvirtInstance) createLibvirtInstance() error {
	domainXMLString, err := domainXMLConfig.Marshal()
	if err != nil {
		return err
	}
	poolXMLString, err := poolXMLConfig.Marshal()
	if err != nil {
		return err
	}
	volumeBootXMLString, err := volumeBootXMLConfig.Marshal()
	if err != nil {
		return err
	}
	volumeBaseXMLString, err := volumeBaseXMLConfig.Marshal()
	if err != nil {
		return err
	}
	volumeStateXMLString, err := volumeStateXMLConfig.Marshal()
	if err != nil {
		return err
	}
	networkXMLString, err := networkXMLConfig.Marshal()
	if err != nil {
		return err
	}
	l.log.Infof("creating network")
	network, err := l.conn.NetworkCreateXML(networkXMLString)
	if err != nil {
		return err
	}
	defer func() { _ = network.Free() }()

	l.log.Infof("creating storage pool")
	poolObject, err := l.conn.StoragePoolDefineXML(poolXMLString, libvirt.STORAGE_POOL_DEFINE_VALIDATE)
	if err != nil {
		return fmt.Errorf("error defining libvirt storage pool: %s", err)
	}
	defer func() { _ = poolObject.Free() }()
	if err := poolObject.Build(libvirt.STORAGE_POOL_BUILD_NEW); err != nil {
		return fmt.Errorf("error building libvirt storage pool: %s", err)
	}
	if err := poolObject.Create(libvirt.STORAGE_POOL_CREATE_NORMAL); err != nil {
		return fmt.Errorf("error creating libvirt storage pool: %s", err)
	}
	volumeBaseObject, err := poolObject.StorageVolCreateXML(volumeBaseXMLString, 0)
	if err != nil {
		return fmt.Errorf("error creating libvirt storage volume 'base': %s", err)
	}
	defer func() { _ = volumeBaseObject.Free() }()

	l.log.Infof("uploading image to libvirt")
	if err := l.uploadBaseImage(volumeBaseObject); err != nil {
		return err
	}

	l.log.Infof("creating storage volume 'boot'")
	bootVol, err := poolObject.StorageVolCreateXML(volumeBootXMLString, 0)
	if err != nil {
		return fmt.Errorf("error creating libvirt storage volume 'boot': %s", err)
	}
	defer func() { _ = bootVol.Free() }()

	l.log.Infof("creating storage volume 'state'")
	stateVol, err := poolObject.StorageVolCreateXML(volumeStateXMLString, 0)
	if err != nil {
		return fmt.Errorf("error creating libvirt storage volume 'state': %s", err)
	}
	defer func() { _ = stateVol.Free() }()

	l.log.Infof("creating domain")
	domain, err := l.conn.DomainCreateXML(domainXMLString, libvirt.DOMAIN_NONE)
	if err != nil {
		return fmt.Errorf("error creating libvirt domain: %s", err)
	}
	defer func() { _ = domain.Free() }()
	return nil
}

func (l *libvirtInstance) deleteNetwork() error {
	nets, err := l.conn.ListAllNetworks(libvirt.CONNECT_LIST_NETWORKS_ACTIVE)
	if err != nil {
		return err
	}
	defer func() {
		for _, net := range nets {
			_ = net.Free()
		}
	}()
	for _, net := range nets {
		name, err := net.GetName()
		if err != nil {
			return err
		}
		if name != networkName {
			continue
		}
		if err := net.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func (l *libvirtInstance) deleteDomain() error {
	doms, err := l.conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return err
	}
	defer func() {
		for _, dom := range doms {
			_ = dom.Free()
		}
	}()
	for _, dom := range doms {
		name, err := dom.GetName()
		if err != nil {
			return err
		}
		if name != domainName {
			continue
		}
		if err := dom.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func (l *libvirtInstance) deleteVolume(pool *libvirt.StoragePool) error {
	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		return err
	}
	defer func() {
		for _, volume := range volumes {
			_ = volume.Free()
		}
	}()
	for _, volume := range volumes {
		name, err := volume.GetName()
		if err != nil {
			return err
		}
		if name != stateDiskName && name != bootDiskName && name != baseDiskName {
			continue
		}
		if err := volume.Delete(libvirt.STORAGE_VOL_DELETE_NORMAL); err != nil {
			return err
		}
	}
	return nil
}

func (l *libvirtInstance) deletePool() error {
	pools, err := l.conn.ListAllStoragePools(libvirt.CONNECT_LIST_STORAGE_POOLS_DIR)
	if err != nil {
		return err
	}
	defer func() {
		for _, pool := range pools {
			_ = pool.Free()
		}
	}()
	for _, pool := range pools {
		name, err := pool.GetName()
		if err != nil {
			return err
		}
		if name != diskPoolName {
			continue
		}
		active, err := pool.IsActive()
		if err != nil {
			return err
		}
		if active {
			if err := l.deleteVolume(&pool); err != nil {
				return err
			}
			if err := pool.Destroy(); err != nil {
				return err
			}
			if err := pool.Delete(libvirt.STORAGE_POOL_DELETE_NORMAL); err != nil {
				return err
			}
		}
		// If something fails and the Pool becomes inactive, we cannot delete/destroy it anymore.
		// We have to undefine it in this case
		if err := pool.Undefine(); err != nil {
			return err
		}
	}
	return nil
}

func (l *libvirtInstance) deleteLibvirtInstance() error {
	var err error
	err = errors.Join(err, l.deleteNetwork())
	err = errors.Join(err, l.deleteDomain())
	err = errors.Join(err, l.deletePool())
	return err
}

func (l *libvirtInstance) obtainMeasurements() (measurements measurements.M, err error) {
	// sanity check
	if err := l.deleteLibvirtInstance(); err != nil {
		return nil, err
	}
	done := make(chan struct{}, 1)
	serv := server.New(l.log, done)
	go func() {
		if err := serv.ListenAndServe("8080"); err != http.ErrServerClosed {
			l.log.With(zap.Error(err)).Fatalf("Failed to serve")
		}
	}()
	defer func() {
		err = errors.Join(err, l.deleteLibvirtInstance())
	}()
	if err := l.createLibvirtInstance(); err != nil {
		return nil, err
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	l.log.Infof("waiting for PCRs or CTRL+C")
	select {
	case <-done:
		break
	case <-sigs:
		break
	}
	signal.Stop(sigs)
	close(sigs)
	if err := serv.Shutdown(); err != nil {
		return nil, err
	}
	close(done)

	return serv.GetMeasurements(), nil
}

func main() {
	var imageLocation, imageType, outFile string
	var verboseLog bool
	flag.StringVar(&imageLocation, "path", "", "path to the image to measure (required)")
	flag.StringVar(&imageType, "type", "", "type of the image. One of 'qcow2' or 'raw' (required)")
	flag.StringVar(&outFile, "file", "-", "path to output file, or '-' for stdout")
	flag.BoolVar(&verboseLog, "v", false, "verbose logging")

	flag.Parse()
	log := logger.New(logger.JSONLog, zapcore.DebugLevel)
	if !verboseLog {
		log = log.WithIncreasedLevel(zapcore.FatalLevel) // Only print fatal errors in non-verbose mode
	}

	if imageLocation == "" || imageType == "" {
		flag.Usage()
		os.Exit(1)
	}
	volumeBootXMLConfig.BackingStore.Format.Type = imageType
	domainXMLConfig.Devices.Disks[1].Driver.Type = imageType

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to connect to libvirt")
	}
	defer conn.Close()

	lInstance := libvirtInstance{
		conn:      conn,
		log:       log,
		imagePath: imageLocation,
	}

	measurements, err := lInstance.obtainMeasurements()
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to obtain PCR measurements")
	}
	log.Infof("instances terminated successfully")

	output, err := yaml.Marshal(measurements)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to marshal measurements")
	}

	if outFile == "-" {
		fmt.Println(string(output))
	} else {
		if err := os.WriteFile(outFile, output, 0o644); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to write measurements to file")
		}
	}
}
