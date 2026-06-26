package mcgo

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FindJava intenta localizar un binario java en el sistema.
// Orden: JAVA_HOME/bin/java, java en PATH, ubicaciones comunes.
func FindJava() (string, error) {
	// 1. JAVA_HOME
	if home := envString("JAVA_HOME"); home != "" {
		bin := javaBin(home + "/bin")
		if javaBinExists(bin) {
			return bin, nil
		}
	}

	// 2. PATH
	if bin, err := exec.LookPath("java"); err == nil {
		return bin, nil
	}

	// 3. Ubicaciones comunes
	for _, p := range commonJavaPaths() {
		if javaBinExists(p) {
			return p, nil
		}
	}

	return "", fmt.Errorf("java not found; set JAVA_HOME or install Java")
}

// javaBinDevuelve la ruta a java ejecutable en un directorio bin.
func javaBin(dir string) string {
	name := "java"
	if runtime.GOOS == "windows" {
		name = "java.exe"
	}
	return filepath.Join(dir, name)
}

func javaBinExists(path string) bool {
	_, err := exec.Command(path, "-version").CombinedOutput()
	return err == nil
}

// commonJavaDevuelve las rutas comunes donde buscar Java por SO.
func commonJavaPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\Java\jre\bin\java.exe`,
			`C:\Program Files\Java\jdk\bin\java.exe`,
			`C:\Program Files\Eclipse Adoptium\jdk-17\bin\java.exe`,
		}
	case "darwin":
		return []string{
			"/Library/Java/JavaVirtualMachines/*/Contents/Home/bin/java",
			"/opt/homebrew/opt/openjdk/bin/java",
		}
	default: // linux/bsd
		return []string{
			"/usr/bin/java",
			"/usr/lib/jvm/default/bin/java",
			"/usr/lib/jvm/java-17-openjdk/bin/java",
			"/opt/java/bin/java",
		}
	}
}

// JavaMajorVersion ejecuta `java -version` y parsea la major version.
func JavaMajorVersion(javaPath string) (int, error) {
	out, err := exec.Command(javaPath, "-version").CombinedOutput()
	if err != nil {
		return 0, err
	}
	s := string(out)
	// Output: 'java version "17.0.1"' o 'openjdk version "17.0.1"' o 'version "1.8.0_291"'
	parts := strings.Split(s, "\"")
	if len(parts) < 2 {
		return 0, fmt.Errorf("cannot parse java version")
	}
	ver := parts[1]
	// "1.8.0_291" -> 8, "17.0.1" -> 17
	nums := strings.Split(ver, ".")
	if nums[0] == "1" && len(nums) > 1 {
		return atoiSafe(nums[1]), nil
	}
	return atoiSafe(nums[0]), nil
}

// envStringLee una variable de entorno.
func envString(key string) string {
	return getenv(key)
}
