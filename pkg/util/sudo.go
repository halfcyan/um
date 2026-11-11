package util

import (
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func isInteractive() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdin.Fd()), unix.TCGETS)
	return err == nil
}

// SudoIfNeeded re-executes um as root via sudo, or pkexec when there's no
// terminal (e.g. from a KCM). pkexec drops forwardEnv because wrapping um in
// `env` would break the polkit action match. it exits 126 or 127 on auth failure.
func SudoIfNeeded(forwardEnv []string) {
	if os.Getuid() == 0 {
		return
	}

	current, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if !isInteractive() {
		if pkexec, err := exec.LookPath("pkexec"); err == nil {
			args := append([]string{"pkexec", current}, os.Args[1:]...)
			if err := unix.Exec(pkexec, args, os.Environ()); err != nil {
				panic(err)
			}
		}
	}

	env := []string{"sudo"}
	for _, v := range forwardEnv {
		val, found := os.LookupEnv(v)
		if !found {
			continue
		}

		env = append(env, v+"="+val)
	}

	args := append(env, current)
	args = append(args, os.Args[1:]...)

	if err := unix.Exec("/usr/bin/sudo", args, os.Environ()); err != nil {
		panic(err)
	}
}
