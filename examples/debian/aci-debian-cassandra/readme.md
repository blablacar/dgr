

# volume mount is mandatory to start until rkt support non root on empty-volumes :

```
# mkdir /tmp/data
# chmod 777 /tmp/data
# sudo rkt run --net=host --insecure-options=image target/image.aci --volume=data,kind=host,source=/tmp/data
```
