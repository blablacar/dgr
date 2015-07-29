[![build](https://img.shields.io/travis/blablacar/cnt.svg?style=flat)](https://travis-ci.org/blablacar/cnt)

**Highly experimental and in development**

# cnt
Tool to build **RKT** container in a mixup of Chef, Dockerfile and Packer logic

File templating will be resolved on container start using **confd** in local mode

# commands
* cnt build         # build the image
* cnt clean         # clean the build
* cnt install       # store target image to rkt local store
* cnt push          # push target image to remote storage

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
├── image-manifest.json             # Image Manifest of the rocket Container
├── install.sh                      # Script ran by packer at build time
├── files                           # Everything under this folder will be copied verbatim in the target rootfs.
│   └── usr
├── runlevels
│   ├── prestart-early              # Scripts to be run when starting ACI before confd templating
│   │   └── 10.mkdir.sh
│   ├── prestart-late               # Scripts to be run when starting ACI after confd templating
│   │   └── 10.fetch.sh
    └── build-setup                 # Scripts to be run directly on source host before building
        └── 10.prepare-rootfs.sh
```
