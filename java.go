package mcgo

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindJava locates a java binary on the system.
func FindJava() (string, error) {
	// JAVA_HOME
	if home := envOr("JAVA_HOME", ""); home != "" {
		bin := filepath.Join(home, "bin", javaExe())
		if ok, _ := javaOk(bin); ok {
			return bin, nil
		}
	}
	// PATH
	if bin, err := exec.LookPath("java"); err == nil {
		return bin, nil
	}
	// Common locations
	for _, p := range commonJavaDirs() {
		bin := filepath.Join(p, javaExe())
		if ok, _ := javaOk(bin); ok {
			return bin, nil
		}
	}
	return "", fmt.Errorf("java not found. Set JAVA_HOME or install Java")
}

func javaExe() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func javaOk(path string) (bool, error) {
	_, err := exec.Command(path, "-version").CombinedOutput()
	return err == nil, err
}

func commonJavaDirs() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\Java\jre\bin`,
			`C:\Program Files\Java\jdk\bin`,
			`C:\Program Files\Eclipse Adoptium\jdk-17\bin`,
		}
	case "darwin":
		return []string{
			"/Library/Java/JavaVirtualMachines/*/Contents/Home/bin",
			"/opt/homebrew/opt/openjdk/bin",
		}
	default:
		return []string{
			"/usr/lib/jvm/default/bin",
			"/usr/lib/jvm/java-17-openjdk/bin",
			"/opt/java/bin",
		}
	}
}
