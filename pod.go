package main

type Pod struct {
	path string
}

func OpenPod(path string, args BuildArgs) (*Pod, error) {
	pod := new(Pod)
	pod.path = path
	return pod, nil
}

func (* Pod) Build() {
	println("Building pod")
}
