# argo

*argo* is a really dumb and simple JSON log file forwarder to ElasticSearch.

* harversts multiple files
* sends events to the specified elasticsearch host (currently compatible with v7)
* registers file offset state in [boltdb](https://github.com/etcd-io/bbolt)

# Status

*argo* is just a personal project. There is absolutely no reason to try it in any kind of
environment :)

# Getting started

Assuming you have Go 1.11 or later you can build *argo* with:

```shell
$ make
```

Run the tiny test suite with:

```shell
make test
```

# Configuration

The following settings are currently supported:

| Setting              | Description                | Default  |
| -------------------- | -------------------------- | ----- |
| `host` (string)      | The elasticsearch host URL | "" |
| `paths` ([]string)   | The file paths to forward  | [] |
| `dispatch_interval` (int64) | Seconds to wait until next dispatch to the ES host | 5 |
| `timeout` (int64)    | Seconds to wait until closing the connection to the ES host | 10 |
| `dead_time` (string) | Duration to keep files alive after being inactive | "24h" |

For a sample configuration file refer to [`config.sample.json`](config.sample.json).

# Usage

To start *argo*, a configuration file is needed:

```shell
$ ./argo --config config.json
```
You can use [`config.sample.json`](config.sample.json) as a starting point.

You can also set the name of the boltdb file with the `--registry` flag:

```shell
$ ./argo --config config.json --registry ma.db
```

> It defaults to `argo.db`.
