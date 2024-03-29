package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("nsInitialisation", nsInitialisation)
	if reexec.Init() {
		os.Exit(0)
	}
}

// func extractFS() error {
// 	cmd := exec.Command("tar", "-x", "/assets/bionic-server-cloudimg-i386-root.tar.xz")
// 	err := cmd.Run()
// 	return err
// }

func nsInitialisation() {
	newrootPath := os.Args[1]

	// if err := mountProc(newrootPath); err != nil {
	// 	fmt.Printf("Error mounting /proc - %s\n", err)
	// 	os.Exit(1)
	// }

	if err := pivotRoot(newrootPath); err != nil {
		fmt.Printf("Error running pivot_root - %s\n", err)
		os.Exit(1)
	}

	nsRun()
}

func nsRun() {
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=-[ns-process]- # "}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}

func main() {
	// if err := extractFS(); err != nil{
	// 	fmt.Printf("Couldn't extract file system " + err.Error())
	// 	os.Exit(1)
	// }

	var rootfsPath string
	flag.StringVar(&rootfsPath, "rootfs", "/tmp/ns-process/rootfs", "Path to the root filesystem to use")
	flag.Parse()
	if err := syscall.Mkdir(rootfsPath, 0700); err != nil{
		if(err.Error() == "file exists"){
			fmt.Printf("Directory exists...Continuing...\n")
		}else{
			fmt.Printf("Error regarding MKDIR() - %s\n", err)
			os.Exit(1)
		}
	}

	cmd := reexec.Command("nsInitialisation", rootfsPath)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting the reexec.Command - %s\n", err)
		os.Exit(1)
	}


	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}
