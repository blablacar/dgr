**Highly experimental and in development**

# cnt
Tool to build container


# cnt configuration file

cnt global conf is a `config.yml` file located in :
* windows >  $HOME + "/AppData/Local/Cnt";
* darwin > $HOME + "/Library/Cnt";
* linux > $HOME + "/.config/cnt";

with content :
```
push:
  type: maven
  url: https://localhost/nexus
  username: admin
  password: admin 

```

#Standard FileTree
```bash
├── attributes
│   └── attributes.yml # Attributes file for confd
├── confd
│   ├── conf.d 
│   │   └── config.toml # Confd template resource config
│   └── templates
│       └── config.tmpl # Confd source template
├── image-manifest.json # Image Manifest of the rocket Container
├── install.sh # Script ran by packer at build time
├── rootfs # Everything under this folder will be copied verbatim in the target rootfs.
│   └── usr
├── runlevels
│   ├── prestart-early # Scripts to be run before confd
│   │   └── 10.mkdir.sh
│   └── prestart-late # Scripts to be run after confd
│       └── 10.fetch.sh
   
```