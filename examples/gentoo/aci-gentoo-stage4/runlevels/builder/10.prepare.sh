#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

###############################################################################
gentoo_mirror="http://gentoo.mirrors.ovh.net/gentoo-distfiles/"
###############################################################################

stage4=$(curl --silent ${gentoo_mirror}/releases/amd64/autobuilds/latest-stage4-amd64-minimal.txt | grep -v "^#" | egrep "stage4-amd64-minimal" | cut -d' ' -f 1 )
stage4tarball=${stage4/*\//}
stage4url="${gentoo_mirror}/releases/amd64/autobuilds/${stage4}"

for elem in .DIGESTS .DIGESTS.asc ""
do
    if [ -f "${stage4tarball}${elem}" ]; then
        echo_purple "${stage4tarball}${elem}: file already exists locally. Will not download it"
    else
        echo_green "Downloading ${stage4url}${elem}"
        trap "rm ${stage4tarball}${elem}" EXIT
        wget ${stage4url}${elem}
        trap - EXIT
    fi
done

#echo_green "Import gentoo stage3 gpg public key"
#gpg --recv-keys 0xBB572E0E2D182910
#gpg --keyserver subkeys.pgp.net --recv-keys 96D8BF6D 2D182910 17072058
#
#echo_green "Checking sha512 of ${stage4tarball}"
#grep "$(openssl dgst -r -sha512 "${stage4tarball}")" "${stage4tarball}.DIGESTS"
#echo_green "Checking whirlpool of ${stage4tarball}"
#grep "$(openssl dgst -r -whirlpool "${stage4tarball}")" "${stage4tarball}.DIGESTS"
#echo_green "Checking signature of ${stage4tarball}"
#gpg --verify "${stage4tarball}.DIGESTS.asc"

echo_green "Extracting stage4"
#TODO builder is missing libacl-devel, libattr-devel, and libselinux-devel for tar to support xattrs
tar xjpf ${stage4tarball} -C "${ROOTFS}" --xattrs
