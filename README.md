[![Build Status](https://travis-ci.org/blablacar/cnt.svg?branch=master)](https://travis-ci.org/blablacar/cnt)

**Highly experimental and in development**

# cnt
Tool to build **RKT** ACI and POD in a mixup of Chef, Dockerfile and Packer logic

File templating will be resolved on container start using **confd** in env mode

# commands
```bash
$ cnt build         # build the image
$ cnt clean         # clean the build
$ cnt install       # store target image to rkt local store
$ cnt push          # push target image to remote storage
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

#Standard FileTree
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
