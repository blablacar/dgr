
# aci-base

```bash
cd aci-base
sudo dgr -c install
```

# aci-grafana

```bash
cd ../aci-grafana
sudo dgr -c build
sudo rkt --net=host --insecure-options=image run --interactive target/image.aci
```

# aci-prometheus

```bash
cd ../aci-prometheus
sudo dgr -L trace -c build
mkdir /tmp/prom
sudo rkt --set-env=LOG_LEVEL=trace  --net=host --insecure-options=image run --interactive target/image.aci --volume=data,kind=host,source=/tmp/prom
```
