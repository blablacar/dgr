package utils
import "io/ioutil"

const (
	packer_filename = "/packer.json"
	packer_content = `{
"builders": [
  {
    "type": "docker",
    "image": "docker-registry.priv.blablacar.net/blablajessie",
    "export_path": "rootfs.tar"
  }
  ],
  "provisioners": [
    {
      "type": "shell",
      "script": "cnt-provision.sh"
    },
    {
      "type": "shell",
      "script": "../install.sh"
    },
    {
      "type": "shell",
      "script": "cnt-cleanup.sh"
    }
  ]
}
`
	pre_filename = "/cnt-provision.sh"
	pre_content = `#!/bin/bash

wget -qO- http://debmirror.priv.blablacar.net/debian/key.gpg | apt-key add - && \
echo 'deb http://debmirror.priv.blablacar.net/debian/ jessie main contrib non-free' > /etc/apt/sources.list

# package pinning
cat > /etc/apt/preferences.d/blablacar.pref <<EEOF
Package: *
Pin: release o=BlaBlaCar
Pin-Priority: 900
EEOF

apt-get update
`
	post_filename = "/cnt-cleanup.sh"
	post_content = `#!/bin/bash

apt-get purge curl chef -y
apt-get autoremove --purge -y
rm -rf /tmp/*
rm -rf /var/lib/apt/lists/*
`

)

func WritePackerFiles(target string) {
	// TODO check error
	ioutil.WriteFile(target + packer_filename, []byte(packer_content), 0644)
	ioutil.WriteFile(target + pre_filename, []byte(pre_content), 0644)
	ioutil.WriteFile(target + post_filename, []byte(post_content), 0644)
}
