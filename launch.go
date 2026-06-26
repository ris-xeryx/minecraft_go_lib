package mcgo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Launch builds the JVM command and starts Minecraft. Returns the running *exec.Cmd.
func Launch(opts LaunchOpts) (*exec.Cmd, error) {
	if opts.Bus == nil {
		opts.Bus = NewEventBus()
	}
	java, err := FindJava()
	if err != nil {
		return nil, err
	}
	info, err := fetchVersion(opts.Version)
	if err != nil {
		return nil, err
	}
	p := Host()

	cp := classpath(info, opts.Dir, p)
	natives := filepath.Join(opts.Dir, "natives", info.ID)
	args := buildJVM(opts, info, cp, natives, p)
	args = append(args, info.MainClass)
	args = append(args, buildGameArgs(info, opts, cp, p)...)

	opts.Bus.Emit(Event{Type: EvtLaunchStarted, Message: "Launching"})

	cmd := exec.Command(java, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		opts.Bus.emitError(err)
		return nil, err
	}
	opts.Bus.Emit(Event{Type: EvtProcessStarted, Message: fmt.Sprintf("pid=%d", cmd.Process.Pid)})
	return cmd, nil
}

func buildJVM(opts LaunchOpts, info *VersionInfo, cp, natives string, p Platform) []string {
	args := []string{
		"-Xms" + opts.MinRAM,
		"-Xmx" + opts.MaxRAM,
	}
	if info.Arguments != nil {
		for _, arg := range info.Arguments.JVM {
			if Allowed(arg.Rules, p, nil) {
				args = append(args, renderArgs(arg.Value, info, opts, cp)...)
			}
		}
	} else {
		args = append(args, "-Djava.library.path="+natives, "-cp", cp)
	}
	args = append(args, opts.JVMArgs...)
	return args
}

func buildGameArgs(info *VersionInfo, opts LaunchOpts, cp string, p Platform) []string {
	var args []string
	if info.Arguments != nil {
		for _, arg := range info.Arguments.Game {
			if Allowed(arg.Rules, p, nil) {
				args = append(args, renderArgs(arg.Value, info, opts, cp)...)
			}
		}
	} else if len(info.GameArguments) > 0 {
		args = append(args, renderArgs(strings.Fields(strings.Join(info.GameArguments, " ")), info, opts, cp)...)
	}
	return args
}

func renderArgs(values []string, info *VersionInfo, opts LaunchOpts, cp string) []string {
	var out []string
	for _, v := range values {
		out = append(out, tmpl(v, info, opts, cp))
	}
	return out
}

func tmpl(s string, info *VersionInfo, opts LaunchOpts, cp string) string {
	r := strings.NewReplacer(
		"${auth_player_name}", opts.Profile.Username,
		"${auth_uuid}", opts.Profile.UUID,
		"${auth_access_token}", opts.Profile.AccessToken,
		"${auth_session}", opts.Profile.AccessToken,
		"${version_name}", info.ID,
		"${game_directory}", opts.Dir,
		"${assets_root}", filepath.Join(opts.Dir, "assets"),
		"${assets_index_name}", info.Assets,
		"${game_assets}", filepath.Join(opts.Dir, "assets"),
		"${user_type}", "mojang",
		"${version_type}", info.Type,
		"${natives_directory}", filepath.Join(opts.Dir, "natives", info.ID),
		"${launcher_name}", "mcgo",
		"${launcher_version}", "0.1",
		"${classpath}", cp,
	)
	return r.Replace(s)
}

func classpath(info *VersionInfo, dir string, p Platform) string {
	var paths []string
	base := filepath.Join(dir, "libraries")
	for i := range info.Libraries {
		lib := &info.Libraries[i]
		if lib.ok(p) && lib.Downloads != nil && lib.Downloads.Artifact != nil {
			paths = append(paths, filepath.Join(base, lib.Downloads.Artifact.Path))
		}
	}
	jarID := info.ID
	if info.JAR != "" {
		jarID = info.JAR
	}
	paths = append(paths, filepath.Join(dir, "versions", jarID, jarID+".jar"))
	if p.OS == "windows" {
		return strings.Join(paths, ";")
	}
	return strings.Join(paths, ":")
}
