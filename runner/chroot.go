package runner

type ChrootRunner struct {
}


//func (r *ChrootRunner) Prepare(target string) {
//	log.Println("Mounting elements in rootfs ", newRoot)
//	utils.ExecCmd("mount", "-t", "proc", "proc", newRoot + "/proc/")
//	utils.ExecCmd("mount", "-t", "sysfs", "sys", newRoot + "/sys/")
//	utils.ExecCmd("mount", "-o", "bind", "/dev", newRoot + "/dev/")
//	utils.ExecCmd("mount", "-o", "bind", "/dev/shm", newRoot + "/dev/shm")
//
//	utils.ExecCmd("mount", "-o", "bind", "/etc/resolv.conf", newRoot + "/etc/resolv.conf")
//	//	ExecCmd("ln", "-sf", "/etc/resolv.conf", newRoot + "/etc/resolv.conf") //TODO replace with mount since network can change and also ln is not accessible from chroot
//}
//
//func (r *ChrootRunner) Run(target string, command ...string) {
//		// PREPARE blablabuild
//		// wget de blbablabuild sur nexus
//		// extraire dans /root/.cnt
//		buildChroot := UserHomeOrFatal() + "/.cnt/build-rootfs"
//		//	err := os.MkdirAll(target, 0755)
//		//	checkFatal(err)
//		// TODO extract files to rootfs
//
//
//		randomPath := utils.RandSeq(10)
//		tmpBuildPath := buildChroot + "/builds/" + randomPath
//		os.Mkdir(tmpBuildPath, 0755)
//		defer os.Remove(tmpBuildPath)
//
//		setupChroot(buildChroot)
//		ExecCmd("mount", "-o", "bind", cnt.path + target, tmpBuildPath)
//		defer ExecCmd("umount", tmpBuildPath)
//
//		ExecCmd("chroot", buildChroot, "/builds/" + randomPath + "/build.sh", randomPath)
//		defer releaseChrootIfNotUsed(buildChroot) //TODO use signal capture to cleanup in case of CTRL + C
//}
//
//func (r *ChrootRunner) Release(/*target string, noBuildImage bool*/) {
//	log.Println("Unmounting elements in rootfs ", newRoot)
//	utils.ExecCmd("umount", newRoot + "/dev/shm")
//	utils.ExecCmd("umount", newRoot + "/dev/")
//	utils.ExecCmd("umount", newRoot + "/sys/")
//	utils.ExecCmd("umount", newRoot + "/proc/")
//}
