#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

###############################################################################
gentoo_mirror="http://de-mirror.org/gentoo"
###############################################################################

stage3=$(curl --silent ${gentoo_mirror}/releases/amd64/autobuilds/latest-stage3.txt | grep -v "^#" | egrep "stage3-amd64-[0-9]+" | cut -d' ' -f 1 )
stage3tarball=${stage3/*\//}
stage3url="${gentoo_mirror}/releases/amd64/autobuilds/${stage3}"


cd "${TARGET}"
for elem in .DIGESTS .DIGESTS.asc ""
do
    if [ -f "${stage3tarball}${elem}" ]; then
        echo_purple "${stage3tarball}${elem}: file already exists locally. Will not download it"
    else
        echo_green "Downloading ${stage3url}${elem}"
        trap "rm ${TARGET}/${stage3tarball}${elem}" EXIT
        wget ${stage3url}${elem}
        trap - EXIT
    fi
done

echo_green "Import gentoo stage3 gpg public key"
#gpg --recv-keys 0xBB572E0E2D182910
#gpg --keyserver subkeys.pgp.net --recv-keys 96D8BF6D 2D182910 17072058

echo_green "Checking sha512 of ${stage3tarball}"
grep "$(openssl dgst -r -sha512 "${stage3tarball}")" "${stage3tarball}.DIGESTS"
echo_green "Checking whirlpool of ${stage3tarball}"
grep "$(openssl dgst -r -whirlpool "${stage3tarball}")" "${stage3tarball}.DIGESTS"
echo_green "Checking signature of ${stage3tarball}"
gpg --verify "${stage3tarball}.DIGESTS.asc"


portageTarball="portage-latest.tar.bz2"
portageurl="${gentoo_mirror}/snapshots/portage-latest.tar.bz2"

cd "${TARGET}"
for elem in .md5sum .gpgsig ""
do
    if [ -f "${portageTarball}${elem}" ]; then
        echo_purple "${portageTarball}${elem}: file already exists locally. Will not download it"
    else
        echo_green "Downloading ${portageurl}${elem}"
        trap "rm ${TARGET}/${portageTarball}${elem}" EXIT
        wget ${portageurl}${elem}
        trap - EXIT
    fi
done

echo_green "Checking sha512 of ${portageTarball}"
grep "$(openssl dgst -r -md5 "${portageTarball}")" "${portageTarball}.md5sum"
echo_green "Checking signature of ${portageTarball}"
gpg --verify "${portageTarball}.gpgsig" "${portageTarball}"

echo_green "Cleanup of rootfs"
rm -Rf "${TARGET}/rootfs"
mkdir -p "${TARGET}/rootfs"
echo_green "Extracting stage3 to target/rootfs"
tar xjpf ${stage3tarball} -C "${TARGET}/rootfs" --xattrs
tar xjpf ${portageTarball} -C "${TARGET}/rootfs/usr/" --xattrs