# API

## How to use it

```bash
curl 200.17.212.244:4000/v1/healthcheck
```

To check if the api is running.

```bash
curl 200.17.212.244:4000/v1/analysis -XPOST -F "template=105" -F "file=@fff351ad66140a5e49eb323c4bf53700.exe"
```

This will start an analysis using the template 105 and the malware
fff351ad66140a5e49eb323c4bf53700.exe and it will return an ID (in this case was
1139bb22-ca7e-44c1-9995-ad0908d3f609, but this is random) that will be used to
get information about the analysis.

```bash
websocat ws://200.17.212.244:4000/v1/status/1139bb22-ca7e-44c1-9995-ad0908d3f609
```

This will open a WebSocket connection to get the status in real time.

```bash
curl 200.17.212.244:4000/v1/report/1139bb22-ca7e-44c1-9995-ad0908d3f609
```

After the status is marked as "Completed", you can retrieve the report using
this command.

## Build instructions

1. Create api.env

```
PROXMOX_NODE=
PROXMOX_TOKEN_ID=
PROXMOX_TOKEN_SECRET=
PROXMOX_URL=
```

2. Build

```
mage BuildAndDeploy
```

or `go run mage.go BuildAndDeploy` if you don't have mage installed.

You can install mage with `go run mage.go EnsureMage`.
