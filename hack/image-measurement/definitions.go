/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"libvirt.org/go/libvirtxml"
)

var (
	libvirtImagePath = "/var/lib/libvirt/images/"
	baseDiskName     = "constellation-measurement"
	stateDiskName    = "constellation-measurement-state"
	booteDiskName    = "constellation-measurement-boot"
	diskPoolName     = "constellation-measurement-pool"
	domainName       = "constellation-measurement-vm"
	networkName      = "constellation-measurement-net"

	networkXMLConfig = libvirtxml.Network{
		Name: networkName,
		Forward: &libvirtxml.NetworkForward{
			Mode: "nat",
			NAT: &libvirtxml.NetworkForwardNAT{
				Ports: []libvirtxml.NetworkForwardNATPort{
					{
						Start: 1024,
						End:   65535,
					},
				},
			},
		},
		Bridge: &libvirtxml.NetworkBridge{
			Name:  "virbr1",
			STP:   "on",
			Delay: "0",
		},
		DNS: &libvirtxml.NetworkDNS{
			Enable: "yes",
		},
		IPs: []libvirtxml.NetworkIP{
			{
				Family:  "ipv4",
				Address: "10.42.0.1",
				Prefix:  16,
				DHCP: &libvirtxml.NetworkDHCP{
					Ranges: []libvirtxml.NetworkDHCPRange{
						{
							Start: "10.42.0.2",
							End:   "10.42.255.254",
						},
					},
				},
			},
		},
	}
	poolXMLConfig = libvirtxml.StoragePool{
		Name:   diskPoolName,
		Type:   "dir",
		Source: &libvirtxml.StoragePoolSource{},
		Target: &libvirtxml.StoragePoolTarget{
			Path: libvirtImagePath,
			Permissions: &libvirtxml.StoragePoolTargetPermissions{
				Owner: "0",
				Group: "0",
				Mode:  "0711",
			},
		},
	}
	volumeBootXMLConfig = libvirtxml.StorageVolume{
		Type: "file",
		Name: booteDiskName,
		Target: &libvirtxml.StorageVolumeTarget{
			Path: libvirtImagePath + booteDiskName,
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
		BackingStore: &libvirtxml.StorageVolumeBackingStore{
			Path:   libvirtImagePath + baseDiskName,
			Format: &libvirtxml.StorageVolumeTargetFormat{},
		},
		Capacity: &libvirtxml.StorageVolumeSize{
			Unit:  "GiB",
			Value: uint64(10),
		},
	}

	volumeBaseXMLConfig = libvirtxml.StorageVolume{
		Type: "file",
		Name: baseDiskName,
		Target: &libvirtxml.StorageVolumeTarget{
			Path: libvirtImagePath + baseDiskName,
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
		Capacity: &libvirtxml.StorageVolumeSize{
			Unit:  "GiB",
			Value: uint64(10),
		},
	}

	volumeStateXMLConfig = libvirtxml.StorageVolume{
		Type: "file",
		Name: stateDiskName,
		Target: &libvirtxml.StorageVolumeTarget{
			Path: libvirtImagePath + stateDiskName,
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
		Capacity: &libvirtxml.StorageVolumeSize{
			Unit:  "GiB",
			Value: uint64(10),
		},
	}

	port            = uint(0)
	domainXMLConfig = libvirtxml.Domain{
		Title: "measurement-VM",
		Name:  domainName,
		Type:  "kvm",
		Memory: &libvirtxml.DomainMemory{
			Value: 2,
			Unit:  "GiB",
		},
		Resource: &libvirtxml.DomainResource{
			Partition: "/machine",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Current:   2,
			Value:     2,
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "custom",
			Model: &libvirtxml.DomainCPUModel{
				Fallback: "forbid",
				Value:    "qemu64",
			},
			Features: []libvirtxml.DomainCPUFeature{
				{
					Policy: "require",
					Name:   "x2apic",
				},
				{
					Policy: "require",
					Name:   "hypervisor",
				},
				{
					Policy: "require",
					Name:   "lahf_lm",
				},
				{
					Policy: "disable",
					Name:   "svm",
				},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			PAE:  &libvirtxml.DomainFeature{},
			SMM: &libvirtxml.DomainFeatureSMM{
				State: "on",
			},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},

		OS: &libvirtxml.DomainOS{
			// If we set firmware to efi, Loader and NVRam will be chosen
			// automatically
			Firmware: "efi",
			/* Loader: &libvirtxml.DomainLoader{
				Readonly: "yes",
				Secure:   "yes",
				Type:     "pflash",
				// Path:     "/usr/share/edk2-ovmf/x64/OVMF_CODE.secboot.fd",
			}, */
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "q35",
				Type:    "hvm",
			},
			/* NVRam: &libvirtxml.DomainNVRam{
				// Template: "/usr/share/edk2-ovmf/x64/OVMF_VARS.fd",
				NVRam: "/var/lib/libvirt/qemu/nvram/control-plane-image-measurement_VARS.fd",
			}, */
			BootDevices: []libvirtxml.DomainBootDevice{
				{
					Dev: "hd",
				},
			},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sda",
						Bus: "scsi",
					},
					Source: &libvirtxml.DomainDiskSource{
						Index: 2,
						Volume: &libvirtxml.DomainDiskSourceVolume{
							Pool:   diskPoolName,
							Volume: booteDiskName,
						},
					},
				},
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
					Source: &libvirtxml.DomainDiskSource{
						Index: 1,
						Volume: &libvirtxml.DomainDiskSourceVolume{
							Pool:   diskPoolName,
							Volume: stateDiskName,
						},
					},
					Alias: &libvirtxml.DomainAlias{
						Name: "virtio-disk1",
					},
				},
			},
			Controllers: []libvirtxml.DomainController{
				{
					Type:  "scsi",
					Model: "virtio-scsi",
				},
			},
			TPMs: []libvirtxml.DomainTPM{
				{
					Model: "tpm-tis",
					Backend: &libvirtxml.DomainTPMBackend{
						Emulator: &libvirtxml.DomainTPMBackendEmulator{
							Version: "2.0",
							ActivePCRBanks: &libvirtxml.DomainTPMBackendPCRBanks{
								SHA1:   &libvirtxml.DomainTPMBackendPCRBank{},
								SHA256: &libvirtxml.DomainTPMBackendPCRBank{},
								SHA384: &libvirtxml.DomainTPMBackendPCRBank{},
								SHA512: &libvirtxml.DomainTPMBackendPCRBank{},
							},
						},
					},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: networkName,
							Bridge:  "virbr1",
						},
					},
					Alias: &libvirtxml.DomainAlias{
						Name: "net0",
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				{
					Source: &libvirtxml.DomainChardevSource{
						Pty: &libvirtxml.DomainChardevSourcePty{
							Path: "/dev/pts/4",
						},
					},
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Port: &port,
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
					},
					Log: &libvirtxml.DomainChardevLog{
						File: "/tmp/libvirt.log",
					},
				},
			},
			Consoles: []libvirtxml.DomainConsole{
				{
					TTY: "/dev/pts/4",
					Source: &libvirtxml.DomainChardevSource{
						Pty: &libvirtxml.DomainChardevSourcePty{
							Path: "/dev/pts/4",
						},
					},
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "serial",
						Port: &port,
					},
				},
			},
			RNGs: []libvirtxml.DomainRNG{
				{
					Model: "virtio",
					Backend: &libvirtxml.DomainRNGBackend{
						Random: &libvirtxml.DomainRNGBackendRandom{
							Device: "/dev/urandom",
						},
					},
					Alias: &libvirtxml.DomainAlias{
						Name: "rng0",
					},
				},
			},
		},
		OnPoweroff: "destroy",
		OnCrash:    "destroy",
		OnReboot:   "restart",
	}
)
