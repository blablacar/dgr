package runner

type ChrootRunner struct {
}

//func (r ChrootRunner) Run() {
//	// PREPARE blablabuild
//	// wget de blbablabuild sur nexus
//	// extraire dans /root/.cnt
//	buildChroot := UserHomeOrFatal() + "/.cnt/build-rootfs"
//	//	err := os.MkdirAll(target, 0755)
//	//	checkFatal(err)
//	// TODO extract files to rootfs
//
//
//	randomPath := randSeq(10)
//	tmpBuildPath := buildChroot + "/builds/" + randomPath
//	os.Mkdir(tmpBuildPath, 0755)
//	defer os.Remove(tmpBuildPath)
//
//	setupChroot(buildChroot)
//	ExecCmd("mount", "-o", "bind", cnt.path + target, tmpBuildPath)
//	defer ExecCmd("umount", tmpBuildPath)
//
//	ExecCmd("chroot", buildChroot, "/builds/" + randomPath + "/build.sh", randomPath)
//	defer releaseChrootIfNotUsed(buildChroot) //TODO use signal capture to cleanup in case of CTRL + C
//
//	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
//}
//
//func (r ChrootRunner) Release() {
//	log.Println("Unmounting elements in rootfs ", newRoot)
//	ExecCmd("umount", newRoot + "/dev/shm")
//	ExecCmd("umount", newRoot + "/dev/")
//	ExecCmd("umount", newRoot + "/sys/")
//	ExecCmd("umount", newRoot + "/proc/")
//}
//
//func (r ChrootRunner) Prepare() {
//	log.Println("Mounting elements in rootfs ", newRoot)
//	ExecCmd("mount", "-t", "proc", "proc", newRoot + "/proc/")
//	ExecCmd("mount", "-t", "sysfs", "sys", newRoot + "/sys/")
//	ExecCmd("mount", "-o", "bind", "/dev", newRoot + "/dev/")
//	ExecCmd("mount", "-o", "bind", "/dev/shm", newRoot + "/dev/shm")
//
//	ExecCmd("mount", "-o", "bind", "/etc/resolv.conf", newRoot + "/etc/resolv.conf")
//	//	ExecCmd("ln", "-sf", "/etc/resolv.conf", newRoot + "/etc/resolv.conf") //TODO replace with mount since network can change and also ln is not accessible from chroot
//}
