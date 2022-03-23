# PCR-updater

New images result in different PCR values for the image.
This utility program makes it simple to update the expected PCR values of the CLI.

## Usage

### Script

Run `fetch_pcrs.sh` to create Constellations on all supported cloud providers and read their PCR states.
```shell
./fetch_pcrs.sh
```

The result is printed to screen and written as Go code to files in `./pcrs`.
```bash
+ main
+ command -v constellation
+ command -v go
+ mkdir -p ./pcrs
+ constellation create azure 2 Standard_D4s_v3 --name pcr-fetch -y
Your Constellation was created successfully.
++ jq '.azurecoordinators | to_entries[] | select(.key|startswith("")) | .value.PublicIP' -rcM constellation-state.json
+ coord_ip=192.0.2.1
+ go run ../main.go -coord-ip 192.0.2.1 -o ./pcrs/azure_pcrs.go
connecting to Coordinator at 192.0.2.1:9000
PCRs:
{
  "0": "q27iAZeXGAiCPdu1bqRA2gAoyMO2KrXWY4YkTCQowc4=",
  ...
  "9": "dEGJtQe3h+SI0z42yO7TklzwPixtM3iMCUeJPGRozvg="
}
+ constellation terminate
Your Constellation was terminated successfully.
+ constellation create gcp 2 n2d-standard-2 --name pcr-fetch -y
Your Constellation was created successfully.
++ jq '.gcpcoordinators | to_entries[] | select(.key|startswith("")) | .value.PublicIP' -rcM constellation-state.json
+ coord_ip=192.0.2.2
+ go run ../main.go -coord-ip 192.0.2.2 -o ./pcrs/gcp_pcrs.go
connecting to Coordinator at 192.0.2.2:9000
PCRs:
{
  "0": "DzXCFGCNk8em5ornNZtKi+Wg6Z7qkQfs5CfE3qTkOc8=",
  ...
  "9": "gse53SjsqREEdOpImJH4KAb0b8PqIgwI+Ps/XSiFnN4="
}
+ constellation terminate
Your Constellation was terminated successfully.
```

### Manual

To read the PCR state of any running Constellation node, run the following:
```shell
go run main.go -coord-ip <NODE_IP> -coord-port <COORDINATOR_PORT>
```

The output is similar to the following:
```shell
$ go run main.go -coord-ip 192.0.2.3 -coord-port 12345
connecting to Coordinator at 192.0.2.3:12345
PCRs:
{
  "0": "DzXCFGCNk8em5ornNZtKi+Wg6Z7qkQfs5CfE3qTkOc8=",
  "1": "XBoRlWuQx6nIDr5vgUL0DlJHy6H6u1dPU3qK2NyToc8=",
  "10": "WLmYFRmDft/ajZJ056CAhpheU6Vbt73aR8eIQpLRGq0=",
  "11": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "12": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "13": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "14": "4tPyJd6A5g09KduV3+nWZQCiEzHAiRT5DulmAqlvpZU=",
  "15": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "16": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "17": "//////////////////////////////////////////8=",
  "18": "//////////////////////////////////////////8=",
  "19": "//////////////////////////////////////////8=",
  "2": "PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=",
  "20": "//////////////////////////////////////////8=",
  "21": "//////////////////////////////////////////8=",
  "22": "//////////////////////////////////////////8=",
  "23": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "3": "PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=",
  "4": "MmkueFj1rP2seH+bjeIRsO4dUnLnMdl7QgtGoAtQH7M=",
  "5": "ExaiapuIfo0KMBo8wj6kPDORLocgnH1C0G/KY8DcV3A=",
  "6": "PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=",
  "7": "UZcW+fhFRMpFkgU+EfKG2s3KdmgEA+TD2quLmthQHbo=",
  "8": "KLSMootYaHBjysWKq9CAYXkXpeYx9PUBimlSEZGJqUM=",
  "9": "gse53SjsqREEdOpImJH4KAb0b8PqIgwI+Ps/XSiFnN4="
}
```
