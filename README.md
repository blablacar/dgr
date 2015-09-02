# cnt

[![GoDoc](https://godoc.org/blablacar/cnt?status.png)](https://godoc.org/blablacar/cnt) [![Build Status](https://travis-ci.org/blablacar/cnt.svg?branch=master)](https://travis-ci.org/blablacar/cnt)

**Highly experimental and in development**

Tool to build **RKT** ACI and POD in a mixup of Chef, Dockerfile and Packer logic

File templating will be resolved on container start using **confd** in env mode

# commands
```bash
$ cnt build         # build the image
$ cnt clean         # clean the build
$ cnt install       # store target image to rkt local store
$ cnt push          # push target image to remote storage
$ cnt init          # create the filetree
```

# cnt configuration file

cnt global conf is a yaml file located at
* windows >  $HOME + "/AppData/Local/Cnt/config.yml";
* darwin > $HOME + "/Library/Cnt/config.yml";
* linux > $HOME + "/.config/cnt/config.yml";

with content :
```yml
push:
  type: maven
  url: https://localhost/nexus
  username: admin
  password: admin 
```


Building an ACI
===============

#Standard FileTree for ACI
```bash
├── attributes
│   └── attributes.yml              # Attributes file for confd
├── confd
│   ├── conf.d 
│   │   └── config.toml             # Confd template resource config
│   └── templates
│       └── config.tmpl             # Confd source template
├── cnt-manifest.yml                # Manifest
├── files                           # Everything under this folder will be copied verbatim in the target rootfs.
│   └── usr
├── runlevels
│   ├── prestart-early              # Scripts to be run when starting ACI before confd templating
│   │   └── 10.mkdir.sh
│   ├── prestart-late               # Scripts to be run when starting ACI after confd templating
│   │   └── 10.fetch.sh
│   ├── build-setup                 # Scripts to be run directly on source host before building
│   │   └── 10.prepare-rootfs.sh
│   ├── build                       # Scripts to be run when building
│   │   └── 10.install-stuff.sh
│   ├── inherit-build-early         # Scripts stored in ACI and used when building from this image
│   │   └── 00.apt-get-update.sh
│   └── inherit-build-late
│       └── 99.purge.sh

```

#The cnt manifest look like
```yaml
from: aci.test.com/aci-base:5
name: aci.test.com/aci-elasticsearch:1
aci:
  app:
    eventHandlers:
      - { name: pre-start, exec: [ "/usr/local/bin/prestart" ] }
    exec: [
        "/usr/share/elasticsearch/bin/elasticsearch",
          "-p", "/var/run/elasticsearch.pid",
          "-Des.default.config=/etc/elasticsearch/elasticsearch.yml",
          "-Des.default.path.home=/usr/share/elasticsearch",
          "-Des.default.path.logs=/var/log/elasticsearch",
          "-Des.default.path.data=/var/lib/elasticsearch",
          "-Des.default.path.work=/tmp/elasticsearch",
          "-Des.default.path.conf=/etc/elasticsearch"
    ]
    mountPoints:
      - {name: es-data, path: /var/lib/elasticsearch, readOnly: false}
      - {name: es-log, path: /var/log/elasticsearch, readOnly: false}
    dependencies:
        - aci.test.com/aci-php:0.2
```


Building a POD
=============

#Standard FileTree for POD

```bash
├── aci-elasticsearch               # Directory that match the pod app shortname (or name)
│   ├── attributes
│   │   └── attributes.yml          # Attributes file for confd in this ACI
│   ├── files                       # Files to be inserted into this ACI
│   ...  
├── cnt-pod-manifest.yml            # Pod Manifest

```
