package yop

//func (b *Builder) copyInternalIfNeeded() error {
//	files := []string{
//		"/dgr/bin/templater",
//		"/dgr/bin/busybox",
//		"/dgr/bin/functions.sh",
//		"/dgr/bin/prestart",
//	}
//
//	for _, file := range files {
//		src := b.stage1Rootfs + file
//		dst := b.stage2Rootfs + file
//		_, err := os.Stat(dst)
//		if err != nil {
//			if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
//				return errs.WithEF(err, b.fields.WithField("path", filepath.Dir(dst)), "Failed to create directory")
//			}
//			if err := common.CopyFile(src, dst); err != nil {
//				return errs.WithEF(err, data.WithField("src", src).WithField("dst", dst), "Failed to copy file from stage1 to stage2")
//			}
//		} else {
//			srcChecksum, err := ChecksumFile(src)
//			if err != nil {
//				return errs.WithEF(err, b.fields.WithField("src", src), "Failed to checksum src file")
//			}
//			dstChecksum, err := ChecksumFile(dst)
//			if err != nil {
//				return errs.WithEF(err, b.fields.WithField("dst", dst), "Failed to checksum dst file")
//			}
//			if !bytes.Equal(srcChecksum, dstChecksum) {
//				logs.WithFields(b.fields).WithField("file", src).Debug("Src and dst files are not equals, overriding")
//				if err := common.CopyFile(src, dst); err != nil {
//					return errs.WithEF(err, data.WithField("src", src).WithField("dst", dst), "Failed to copy file from stage1 to stage2")
//				}
//			} else {
//				logs.WithFields(b.fields).WithField("file", src).Debug("Src and dst files are equals")
//			}
//		}
//	}
//	return nil
//}
//
