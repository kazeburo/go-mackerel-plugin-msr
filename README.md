# go-mackerel-plugin-msr

Mackerel metrics plugin for MySQL Multi Source Replication

This plugin fetches the seconds behind master of replication.

## Usage

```
Usage:
  mackerel-plugin-msr [OPTIONS]

Application Options:
  -H, --host=     Hostname (default: localhost)
  -p, --port=     Port (default: 3306)
  -u, --user=     Username (default: root)
  -P, --password= Password
  -v, --version   Show version

Help Options:
  -h, --help      Show this help message
```

Example

```
$ ./mackerel-plugin-msr
mysql-msr.behind.main-db\t0\ttime
mysql-msr.behind.user-db\t0\ttime0 
```


  ## Install

Please download release page or `mkr plugin install kazeburo/go-mackerel-plugin-msr`.

