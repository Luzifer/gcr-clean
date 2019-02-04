[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/gcr-clean)](https://goreportcard.com/report/github.com/Luzifer/gcr-clean)
![](https://badges.fyi/github/license/Luzifer/gcr-clean)
![](https://badges.fyi/github/downloads/Luzifer/gcr-clean)
![](https://badges.fyi/github/latest-release/Luzifer/gcr-clean)

# Luzifer / gcr-clean

`gcr-clean` is a small helper to clean unused manifests from the GCR inside a Google Cloud Project. It takes all manifests not anymore tagged and deletes them.

It supports reading authentication information from Google Application Default Credentials (`account.json`) or the Docker configuration.

## Usage

```console
$ gcr-clean --help
Usage of gcr-clean:
      --account string     Path to account.json file with GCR access
      --listen string      Port/IP to listen on (default ":3000")
      --log-level string   Log level (debug, info, warn, error, fatal) (default "info")
  -n, --noop               Do not execute destructive DELETE operation (default true)
  -p, --parallel int       How many deletions to execute in parallel (default 10)
      --registry string    The registry used (gcr.io, eu.gcr.io, us.gcr.io, ...) (default "gcr.io")
      --version            Prints current version and exits

$ gcr-clean luzifer-registry
INFO[0000] Fetching repositories...
INFO[0001] Manifest deleted          manifest="sha256:a411[...]" noop=true repo=luzifer-registry/eventsys
```
