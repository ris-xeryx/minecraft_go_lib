# mcgo — Minecraft launcher library for Go

Create your own Minecraft launcher in Go. Inspired by
`minecraft-launcher-lib` (Python), `MCLC` (Node.js), `lighty-launcher` (Rust).
Built with [opencode](https://github.com/anomalyco/opencode) (DeepSeek V4 Pro).

```bash
go get github.com/ris-xeryx/minecraft_go_lib
```

---

## Quick start: build a launcher in 3 steps

### 1. List versions

```go
package main

import (
	"fmt"
	mcgo "github.com/ris-xeryx/minecraft_go_lib"
)

func main() {
	versions, _ := mcgo.Versions()
	latest, _ := mcgo.Latest("release")
	fmt.Printf("%d versions, latest is %s\n", len(versions), latest)
}
```

### 2. Authenticate

```go
// Offline (no password needed)
profile, _ := mcgo.OfflineAuth{"MyPlayer"}.Login()

// Microsoft (device code)
auth := mcgo.MicrosoftAuth{ClientID: "your-azure-app-client-id"}
auth.OnCode = func(code, url string) {
	fmt.Printf("Go to %s and enter: %s\n", url, code)
}
profile, _ := auth.Login()
```

### 3. Install and launch

```go
// Download Minecraft 1.21.4
bus := mcgo.NewEventBus()
bus.On(func(e mcgo.Event) {
	fmt.Println(e.Message)
})

mcgo.Install("1.21.4", "/home/user/.minecraft", bus)

// Launch it
cmd, _ := mcgo.Launch(mcgo.LaunchOpts{
	Version: "1.21.4",
	Dir:     "/home/user/.minecraft",
	Profile: profile,
	MinRAM:  "2G",
	MaxRAM:  "4G",
})
cmd.Wait()
```

---

## API reference

### Versions & manifests

| Function | Returns |
|----------|---------|
| `Versions()` | All available Minecraft versions |
| `Latest("release")` | Latest release ID |
| `Latest("snapshot")` | Latest snapshot ID |

### Authentication

| Function | Returns |
|----------|---------|
| `OfflineAuth{Username}.Login()` | `Profile` (no password) |
| `MicrosoftAuth{ClientID}.Login()` | `Profile` via device-code flow |
| `MicrosoftAuth.OnCode` | Callback to show user the code/URL |

### Installation

| Function | Returns |
|----------|---------|
| `Install(version, dir, bus)` | Downloads client, libs, assets, natives |

### Loaders

| Function | Returns |
|----------|---------|
| `LoaderVersion(Fabric, "1.21.4")` | Latest loader version string |
| `Loaders()` | `[Vanilla, Fabric, Quilt, NeoForge, Forge]` |

### Launch

| Function | Returns |
|----------|---------|
| `Launch(LaunchOpts{...})` | Running `*exec.Cmd` |

`LaunchOpts` fields: `Version`, `Dir`, `Profile`, `MinRAM`, `MaxRAM`, `JVMArgs`, `Bus`.

### Events

```go
bus.On(func(e mcgo.Event) {
	switch e.Type {
	case mcgo.EvtDownloadProgress:
		fmt.Printf("%d/%d\n", e.BytesLoaded, e.TotalBytes)
	case mcgo.EvtError:
		fmt.Println(e.Error)
	}
})
```

| Constant | When |
|----------|------|
| `EvtInstallStarted` | Install begins |
| `EvtFileChecked` | Checking SHA1 of existing file |
| `EvtDownloadStarted` | Download begins |
| `EvtDownloadProgress` | Bytes loaded (BytesLoaded / TotalBytes) |
| `EvtFileDownloaded` | Single file done |
| `EvtInstallCompleted` | All files downloaded |
| `EvtLaunchStarted` | JVM command built |
| `EvtProcessStarted` | Minecraft process running |
| `EvtError` | Error occurred |

### Utilities

| Function | Returns |
|----------|---------|
| `FindJava()` | Path to Java binary |
| `Host()` | Current `Platform{OS, Arch}` |

## License

GPL v3
