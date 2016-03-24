

Look at build.sh to see the dependency order of aci

here is a example of building/installing/running for prometheus with building dependencies

```bash
dgr -W aci-base clean install
dgr -W none/aci-libc clean install

cd none/aci-prometheus
sudo dgr -L debug build
mkdir /tmp/prom
sudo rkt --set-env=LOG_LEVEL=trace  --net=host --insecure-options=image run --interactive target/image.aci --volume=data,kind=host,source=/tmp/prom
```
