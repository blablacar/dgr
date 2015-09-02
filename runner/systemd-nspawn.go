package runner

//import "github.com/blablacar/cnt/utils"
//
//type SystemdNspawn struct {
//}
//
//func (s *SystemdNspawn) Prepare(target string) error {
//	return nil
//}
//
//func (s *SystemdNspawn) Run(target string) error {
//	return utils.ExecCmd("systemd-nspawn",
//		"--directory="+target+"/rootfs",
//		"--capability=all",
//		"--bind="+target+"/:/target",
//		"--share-system",
//		"target/build.sh")
//}
//
//func (s *SystemdNspawn) release(target string) error {
//	return nil
//}
