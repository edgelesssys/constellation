# State

Files and source code for mounting persistent state disks

## Testing

Integration test is available in `state/test/integration_test.go`.
The integration test requires root privileges since it uses dm-crypt.
Build and run the test:
```bash
go test -c ./state/test/
sudo ./test.test
```
