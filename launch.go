package mcgo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// JavaPathEncontrar java si no se especifica en options.
func (o *LaunchOptions) resolveJavaPath() (string, error) {
	if o.JavaPath != "" {
		return o.JavaPath, nil
	}
	return FindJava()
}

// ClasspathList reúne todos los jars del classpath.
func classpath(info *VersionInfo, dir string, platform Platform) []string {
	var paths []string
	libsDir := filepath.Join(dir, "libraries")
	for i := range info.Libraries {
		lib := &info.Libraries[i]
		if !lib.AllowedFor(platform) {
			continue
		}
		if lib.Downloads == nil || lib.Downloads.Artifact == nil {
			continue
		}
		paths = append(paths, filepath.Join(libsDir, lib.Downloads.Artifact.Path))
	}
	// client jar
	jarID := info.ID
	if info.JAR != "" {
		jarID = info.JAR
	}
	clientJar := filepath.Join(dir, "versions", jarID, jarID+".jar")
	paths = append(paths, clientJar)
	return paths
}

// classpathSeparator por SO.
func classpathSeparator(platform Platform) string {
	if platform.OS == "windows" {
		return ";"
	}
	return ":"
}

// buildGameArgs construye los argumentos del juego.
func buildGameArgs(info *VersionInfo, opts *LaunchOptions, platform Platform) []string {
	var args []string

	// Argumentos modernos (1.13+)
	if info.Arguments != nil {
		for _, arg := range info.Arguments.Game {
			if !RuleAllowed(arg.Rules, platform, nil) {
				continue
			}
			args = append(args, renderArgValues(arg.Value, info, opts)...)
		}
	} else if len(info.GameArguments) > 0 {
		// Legacy: pre-1.13, separados por espacio
		legacyArgs := strings.Join(info.GameArguments, " ")
		args = append(args, renderLegacyArgs(legacyArgs, info, opts)...)
	}

	return args
}

func renderArgValues(values []string, info *VersionInfo, opts *LaunchOptions) []string {
	var out []string
	for _, v := range values {
		out = append(out, renderTemplate(v, info, opts))
	}
	return out
}

// renderTemplate reemplaza placeholders como ${auth_player_name}.
func renderTemplate(s string, info *VersionInfo, opts *LaunchOptions) string {
	r := strings.NewReplacer(
		"${auth_player_name}", opts.Profile.Username,
		"${auth_uuid}", opts.Profile.UUID,
		"${auth_access_token}", opts.Profile.AccessToken,
		"${auth_session}", opts.Profile.AccessToken,
		"${version_name}", info.ID,
		"${game_directory}", opts.Instance.Directory,
		"${assets_root}", filepath.Join(opts.Instance.Directory, "assets"),
		"${assets_index_name}", info.Assets,
		"${game_assets}", filepath.Join(opts.Instance.Directory, "assets"),
		"${user_type}", "mojang",
		"${version_type}", info.Type,
		"${natives_directory}", opts.NativesPath,
		"${launcher_name}", "minecraft_go_lib",
		"${launcher_version}", "0.1",
		"${classpath}", strings.Join(classpath(info, opts.Instance.Directory, CurrentPlatform()),
			classpathSeparator(CurrentPlatform())),
	)
	return r.Replace(s)
}

// renderLegacyArgs aplica la misma lógica a args legacy separados por espacio.
func renderLegacyArgs(joined string, info *VersionInfo, opts *LaunchOptions) []string {
	parts := strings.Fields(joined)
	var out []string
	for _, p := range parts {
		out = append(out, renderTemplate(p, info, opts))
	}
	return out
}

// buildJVMArgs construye los argumentos de JVM (modernos o default).
func buildJVMArgs(info *VersionInfo, opts *LaunchOptions, platform Platform) []string {
	var args []string

	// Defaults comunes
	args = append(args,
		"-Xms"+opts.Memory.Min,
		"-Xmx"+opts.Memory.Max,
	)

	// Argumentos modernos de JVM
	if info.Arguments != nil {
		for _, arg := range info.Arguments.JVM {
			if !RuleAllowed(arg.Rules, platform, nil) {
				continue
			}
			args = append(args, renderArgValues(arg.Value, info, opts)...)
		}
	} else {
		// Defaults legacy
		args = append(args,
			"-Djava.library.path="+opts.NativesPath,
			"-cp",
			strings.Join(classpath(info, opts.Instance.Directory, platform),
				classpathSeparator(platform)),
		)
	}

	// User JVM args
	args = append(args, opts.JVMArgs...)

	return args
}

// Launch construye el commandline y lanza el proceso de Minecraft.
// Devuelve el *exec.Cmd para que el llamador pueda esperar/output.
func Launch(opts LaunchOptions) (*exec.Cmd, error) {
	if opts.Instance.Version == "" {
		return nil, fmt.Errorf("version is required")
	}
	if opts.EventBus == nil {
		opts.EventBus = NewEventBus()
	}

	// Encontrar Java
	javaPath, err := opts.resolveJavaPath()
	if err != nil {
		return nil, err
	}

	// Fetch info de la versión
	info, err := GetVersionInfo(opts.Instance.Version)
	if err != nil {
		return nil, err
	}

	// Natives path
	nativesPath := opts.NativesPath
	if nativesPath == "" {
		nativesPath = filepath.Join(opts.Instance.Directory, "natives", info.ID)
	}

	platform := CurrentPlatform()

	// Construir commandline
	cmdArgs := buildJVMArgs(info, &opts, platform)
	cmdArgs = append(cmdArgs, info.MainClass)
	cmdArgs = append(cmdArgs, buildGameArgs(info, &opts, platform)...)

	if opts.EventBus != nil {
		opts.EventBus.Emit(Event{Type: EventLaunchStarted, Message: "Launching Minecraft"})
	}

	cmd := exec.Command(javaPath, cmdArgs...)
	cmd.Dir = opts.Instance.Directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		opts.EventBus.emitError(fmt.Errorf("launch failed: %w", err))
		return nil, err
	}

	if opts.EventBus != nil {
		opts.EventBus.Emit(Event{Type: EventProcessStarted, Message: fmt.Sprintf("pid=%d", cmd.Process.Pid)})
	}

	return cmd, nil
}
