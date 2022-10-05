#!/bin/bash

# Start libvirt daemon
libvirtd --daemon --listen
virtlogd --daemon

sleep infinity
