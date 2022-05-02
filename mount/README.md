# constellation-mount-utils

## Dependencies

This package uses the C library [`libcryptsetup`](https://gitlab.com/cryptsetup/cryptsetup/) for device mapping.

To install the required dependencies on Ubuntu run:
```shell
sudo apt install libcryptsetup-dev
```


## Testing

A small test program is available in `test/main.go`.
To build the program run:
```shell
go build -o test/crypt ./test/
```

Create a new crypt device for `/dev/sdX` and map it to `/dev/mapper/volume01`:
```shell
sudo test/crypt -source /dev/sdX -target volume01 -v 4
```

You can now interact with the mapped volume as if it was an unformatted device:
```shell
sudo mkfs.ext4 /dev/mapper/volume01
sudo mount /dev/mapper/volume01 /mnt/volume01
```

Close the mapped volume:
```shell
sudo umount /mnt/volume01
sudo test/crypt -c -target volume01 -v 4
```
