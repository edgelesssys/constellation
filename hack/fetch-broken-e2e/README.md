# Motivation

Sometimes E2E pipeline fails in a way that cleanup was not possible, but a state was stored. These scripts help with manual cleanup.

## Usage

```bash
# Downloads states of all recent (last 20) runs
./fetch.sh
# Find the UID of cluster in Azure/GCP you want to delete
./find.sh <UID>
# Follow the instructions
```
