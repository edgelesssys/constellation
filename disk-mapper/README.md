# disk-mapper

The disk-mapper is a binary that runs during the initramfs of a Constellation node.

If running on a new node, it handles setting up the node's state disk by creating an integrity protected encrypted partition.

On a rebooting node, the disk-mapper handles recovery of the node by requesting a decryption key for its state disk.
Once the disk is decrypted, the measurement salt is read from disk and used to extend a PCR to mark the node as initialized.

## Testing

Integration test is available in `disk-mapper/test/integration_test.go`.
The integration test requires root privileges since it uses dm-crypt.
Build and run the test:

```bash
go test -c -tags=integration ./disk-mapper/internal/test/
sudo ./test.test
```
