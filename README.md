# dgr - container build and runtime tool

[![Go Report Card](https://goreportcard.com/badge/github.com/blablacar/dgr)](https://goreportcard.com/report/github.com/blablacar/dgr)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5874B0.svg)](https://godoc.org/github.com/blablacar/dgr)
[![Build Status](https://img.shields.io/travis/blablacar/dgr/master.svg)](https://travis-ci.org/blablacar/dgr)

<img src="https://raw.githubusercontent.com/blablacar/dgr/gh-pages/logo.png" width="300">

**dgr** (pronounced "*digg-er*") is a command line utility designed to build and to configure at runtime App Containers Images ([ACI](https://github.com/appc/spec/blob/master/spec/aci.md)) and App Container Pods ([POD](https://github.com/appc/spec/blob/master/spec/pods.md)) based on convention over configuration.

dgr allows you to build generic container images for a service and to configure them at runtime. Therefore you can use the same image for different environments, clusters, or nodes by overriding the appropriate attributes when launching the container.

_dgr is actively used at blablacar to build and run more than an hundred different aci and pod to [run all platforms](http://blablatech.com/blog/why-and-how-blablacar-went-full-containers)._


## Build the ACI once, configure your app at runtime.

dgr provides various resources to build and configure an ACI:

- scripts at runlevels (build, prestart...)
- templates and attributes
- static files
- images dependencies

**Scripts** are executed at the image build, before your container is started and more. See [runlevels](#runlevels) for more information.

**Templates** and **attributes** are the way dgr deals with environment-specific configurations. **Templates** are stored in the image and resolved at runtime ; **attributes** are inherited from different contexts (aci -> pod -> environment).

**Static files** are copied to the same path in the container.

**Image dependencies** are used as defined in [APPC spec](https://github.com/appc/spec/blob/master/spec/aci.md#dependency-matching).


![demo](https://raw.githubusercontent.com/blablacar/dgr/gh-pages/aci-dummy.gif)


## Commands

```bash
$ dgr init          # init a sample project
$ dgr build         # build the image
$ dgr clean         # clean the build
$ dgr clean build   # just building, clean is always run before building
$ dgr clean install # clean, build and install aci in the local rkt
$ dgr clean push    # clean, build and push aci to remote storage
$ dgr clean test    # clean, build and test aci
$ dgr install       # use already built aci in target directory to install in rkt
$ dgr push          # use already built aci in target directory to push to remote storage
$ dgr test          # run tests on already built aci
$ dgr try           # run templating only to target/try (experimental)
```

There is a lot of different flags on each command. use the helper to see them :
```bash
$ dgr --help
...
$ dgr build --help
...
```

## Configuration file

dgr global configuration is a yaml file located at `~/.config/dgr/config.yml`. Home is the home of starting user (the caller user if running with sudo).

**targetWorkDir** is used to indicate the target work directory where dgr will work to build and create the ACI
**push*** contain informations on how to push the aci/pod to remote storage
**rkt** if you are not using rkt in your path, or want to create specif config

Example of configuration:

```yml
targetWorkDir: /tmp/target      # if you want to use another directory for all builds
push:
  type: maven
  url: https://localhost/nexus
  username: admin
  password: admin
rkt:                            # arguments to rkt. See rkt --help
  path:
  insecureOptions: [image]
  dir: /var/lib/rkt
  localConfig: /etc/rkt
  systemConfig: /usr/lib/rkt
  userConfig:
  trustKeysFromHttps: false
  noStore: false                # can be set by command line
  storeOnly: false              # can be set by command line
```


# Building an ACI

## Initializing a new project

Run the following commands to initialize a new complete sample project:

```bash
$ mkdir aci-myapp
$ cd aci-myapp
$ dgr init
```

It will generate the following file tree:

```text
.
|-- attributes
|   `-- attributes.yml                 # Attributes files that will be merged and used to resolve templates
|-- aci-manifest.yml                   # Manifest
|-- templates
|   |-- etc
|   |   |-- templated.tmpl             # template file that will end up at /etc/templated
|   |   `-- templated.tmpl.cfg         # configuration of the targeted file, like user and mode (optional file)
|   `-- header.partial                 # template part that can be included in template files
|-- files
|   `-- dummy                          # Files to be copied to the same location in the target rootfs
|-- runlevels
|   |-- builder
|   |   `-- 10.prepare.sh              # Scripts to be run inside the builder to prepare the aci for build
|   |-- build
|   |   `-- 10.install.sh              # Scripts to be run when building inside aci's rootfs
|   |-- build-late
|   |   `-- 10.build-late.sh           # Scripts to be run when building inside aci's rootfs after the copy of files
|   |-- inherit-build-early
|   |   `-- 10.inherit-build-early.sh  # Scripts stored in ACI and executed while used as a dependency
|   |-- inherit-build-late
|   |   `-- 10.inherit-build-late.sh   # Scripts stored in ACI and executed while used as a dependency
|   |-- prestart-early
|   |   `-- 10.prestart-early.sh       # Scripts to be run when starting ACI before templating
|   `-- prestart-late
|       `-- 10.prestart-late.sh        # Scripts to be run when starting ACI after templating
`-- tests
    |-- dummy.bats                     # Bats tests for this ACI
    `-- wait.sh                        # Script to wait until the service is up before running tests
```

This project is already valid which means that you can build it and it will result in a runnable ACI (dgr always adds busybox to the ACI). But you probably want to customize it at this point.

The only mandatory information is the `aci-manifest.yml`, with only the aci `name:`. You can remove everything else depending on you needs.  

## Nice other features

- builder runlevel with dependencies allow you build a project of any kind (java, php, go, node, ...) and release an aci without anything else than dgr and rkt on the host
- dgr will tell you if you are not using the latest version of a dependency and will tell you which version is the latest
- integrated test system that can be extended to support any kind of test system
- working with [pods](https://github.com/appc/spec/blob/master/spec/pods.md) as a unit during build too
- build application version based on container name
- extract aci version from the version of the software during installation

## How it's working
<img style="margin: 10px 30px 40px 0" src="https://docs.google.com/drawings/d/1bSP6Z2X79xkp6deSNaZ-ShrAPjAPa4bzyjL4df2HLwk/pub?w=850">

dgr uses the **builder** information from the **aci-manifest.yml** to construct a rkt stage1. dgr then start rkt with this stage1 on an empty container with the final manifest of your aci (to have dependencies during build).

Inside rkt, the builder isolate the build process inside a **systemd-nspawn** on the builder's rootfs (with mount point on the final aci's rootfs and aci's home) and run the following steps :
- using internal dgr filesystem (busybox, openssl, wget, curl) for the builder if no dependencies (nothing in /usr/bin)
- run **builder** runlevel
- copy **templater** and **inherit** runlevels
- isolate on final rootfs and run **build** runlevels
- copy **prestart**, **attributes**, **files**, **templates**
- isolate on final rootfs and run **build-late** runlevels


## Customizing

### The manifest

The dgr manifest looks like a light ACI manifest with extra builder and tester info. 
dgr will take the `aci` part and convert it to the format defined in the APPC spec.

Example of a *aci-manifest.yml*:

```yaml
name: example.com/myapp:0.1

builder:
  dependencies:
    - example.com/base:1

tester:
  dependencies:
    - example.com/base:1
    
aci:
  dependencies:
    - example.com/base:1
  app:
    exec:
      - /bin/myapp
      - -c
      - /etc/myapp/myapp.cfg
    mountPoints:
      - name: myapp-data
        path: /var/lib/myapp
        readOnly: false
```

The **name**, well, is the name of the ACI you are building.

**builder** node 

Under the **aci** key, you can add every key that is defined in the [APPC spec](https://github.com/appc/spec/blob/master/spec/aci.md) such as:

- **exec** which contains the absolute path to the executable your want to run at the start of the ACI and its args.
- **mountPoints** even though you can do it on the command line with recent versions of RKT.
- **isolators**...

Except **handlers** that are directly mapped to **prestart** runlevels 

### Runlevels

The scripts in `runlevels/build` dir are executed during the build to install in the ACI everything you need. For instance if your dependencies are based on debian, a build script could look like:

```bash
#!/bin/bash
apt-get update
apt-get install -y myapp
```

### Templates

You can create templates in your ACI. Templates are stored in the ACI as long as attributes and are resolved at start of the container.

Example:

*templates/etc/resolv.conf.tmpl*

```
{{ range .dns.nameservers -}}
nameserver {{ . }}
{{ end }}

{{ if .dns.search -}}
search {{ range .dns.search }} {{.}} {{end}}
{{end}}
```

*templates/etc/resolv.conf.tmpl.cfg*

```
uid: 0
gid: 0
mode: 0644
checkCmd: /dgr/bin/busybox true
```

`checkCmd` is a command to run after the templating to check that the configuration is valid or fail container start.

When you have to reuse the same part in multiple templates, you can create a partial template like defined in the [go templating](https://golang.org/pkg/text/template/#hdr-Nested_template_definitions).

*templates/header.partial*

```
{{define "header"}}
whatever
{{end}}
```

and include it in a template:

```
{{template "header" .}}
```

Templater provides functions to manipulate data inside the template. Here is the list:

| Tables    |      Function        |  Description                                                |
|-----------|:---------------------|:------------------------------------------------------------|
| base      | path.Base            |                                                             |
| split     | strings.Split        |                                                             |
| json      | UnmarshalJsonObject  |                                                             |
| jsonArray | UnmarshalJsonArray   |                                                             |
| dir       | path.Dir             |                                                             |
| getenv    | os.Getenv            |                                                             |
| join      | strings.Join         |                                                             |
| datetime  | time.Now             |                                                             |
| toUpper   | strings.ToUpper      |                                                             |
| toLower   | strings.ToLower      |                                                             |
| contains  | strings.Contains     |                                                             |
| replace   | strings.Replace      |                                                             |
| orDef     | orDef                | if first element is nil, use second as default              |
| orDefs    | orDefs               | if first array param is empty use second element to fill it |
| ifOrDef   | ifOrDef              | if first param is not nil, use second, else third           |

It also provide all function defined by [gtf project](https://github.com/leekchan/gtf)

*We can add functions on demand*

### Attributes

All the YAML files in the directory **attributes** are read by dgr. The first node of the YAML has to be "default" as it can be overridden in a POD or with a json in the env variable TEMPLATER_OVERRIDE in the cmd line.

*attributes/resolv.conf.yml*

```
default:
  dns:
    nameservers:
      - "8.8.8.8"
      - "8.8.4.4"
    search:
      - bla.com
```

### Prestart

dgr uses the "pre-start" eventHandler of the ACI to customize the ACI rootfs before the run depending on the instance or the environment.
It resolves at that time the templates so it has all the context needed to do that.
You can also run custom scripts before (prestart-early) or after (prestart-late) this template resolution. This is useful if you want to initialize a mountpoint with some data before running your app for instance.

*runlevels/prestart-late/init.sh*

```bash
#!/bin/bash
set -e
/usr/bin/myapp-init
```




## Running the aci

At this stage you should have a runnable aci. During build, dgr integrated into the aci a prestart that will take care of running templater using `templates` and `attributes`

### log level
Templates and default attribute values are integrated into the aci.
At start you can change log level of prestart scripts and the templater with the environment variable `--set-env=LOG_LEVEL=trace`.
default level is info. At `debug`, prestart shell script will activate debug (set -x). At level `trace`, templater will display the result of templating.


### Override template's attributes
Default attributes values integrated in the aci can be overridden by adding a json tree in the environment variable `TEMPLATER_OVERRIDE`


### example
```
# sudo rkt --set-env=LOG_LEVEL=trace  --net=host --insecure-options=image run --interactive target/image.aci '--set-env=TEMPLATER_OVERRIDE={"dns":{"nameservers":["10.11.254.253","10.11.254.254"]}}'
```

## Troubleshoot
dgr start by default with info log level. You can change this level with the `-L` command line argument.
The log level is also propagated to all runlevels with the environment variable: **LOG_LEVEL**.

You can activate debug on demand by including this code in your scripts:

```
#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x
```
Build it
```bash
$ dgr -L debug build
```

You can also debug the start of your container (prestart, templates) the same way
```bash
$ rkt run --set-env=LOG_LEVEL=debug example.com/my-app
```

**trace** loglevel, will tell the templater to display the result


## Building a POD

A pod is a group of aci that will build and run together as a single unit.

### Standard FileTree for POD
TODO

```bash
├── aci-elasticsearch               # Directory that match the pod app shortname (or name)
│   ├── attributes
│   │   └── attributes.yml          # Attributes file for templating in this ACI
│   ├── files                       # Files to be inserted into this ACI
│   ...
├── pod-manifest.yml            # Pod Manifest
```


## Ok, but concretely how should I use it?

*have a look at the examples/ directory where you can find aci for various distrib*

Depending on distrib, package manager and what you want to do, you will not work the same way. but globally there is 2 way of building an aci.

#### Building directly inside the aci
This is what you will see everywhere else in docker or rkt. You use the **build** and **build-late** runlevels and run commands on the the final rootfs (like apt-get install...)

#### Building outside of the aci
If you are using a package manager that support working outside of the target's rootfs or want to build a project, you will work outside of the stage1 directly inside the builder.
For example if you are buiding an aci for a go project from sources. you will prepare a **builder** with **go** to be able to build the project on the stage1 and put the binary on the aci's rootfs (go is not needed to run the aci).

*At this step, everybody can build any kind of project, since nothing on the host is used to build the project and the aci.*

Also, if you are using a package manager like `pacman` or `emerge`, you can build and install packages on the final **rootfs** without build dependencies nor the package manager.

#### Note About dependencies

Most package manager are not design for overlay and are working with a db file for installed software. this means than when your aci have multiple dependencies on the aci, the db files will overlap and the package manager will only see half of package installed.

As far as I know only `pacman`, that uses a file tree structure for install package, can support overlay.
If you are using a debian or similar. I recommand to limit the dependencies to only 2 layers. The base aci with debian minimal fs and one with the application you want.


## Comparison with alternatives

### dgr vs Dockerfile
A Dockerfile is purely configuration, describing the steps to build the container. It does not provide a common way of building containers across a team.
It does not provide scripts levels, ending with very long bash scripting for the run option in the dockerfile.
It does not handle configuration, nor at build time nor at runtime and does not support any kind of build outside of the container feature.

### dgr vs acbuild
acbuild is a command line tools to build ACIs. It is more flexible than Dockerfiles as it can be wrapped by other tools such as Makefiles but like Dockerfiles it doesn't provide a standard way of configuring the images.


## Requirement
- [rkt](https://github.com/coreos/rkt) in your `$PATH` or configured in dgr global conf
- being root is required to call rkt
- linux >= 3.18 with overlay filesystem


## I want to extend dgr
If you think your idea can be integrated directly in the core of dgr, please create an issue or a pull request.

If you want want to extend the way the **builder** is working (attributes, templates, files, ...), you can create a new **stage1 builder** and replace the internal one with : 
```
...
builder:
  image: dgrtool.com/aci-builder:1
...
```
You can do the same for the **tester**.




