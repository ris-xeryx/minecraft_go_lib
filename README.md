# minecraft_go_lib

> Librería Go para crear launchers de Minecraft. Inspirada en `minecraft-launcher-lib` (Python), `MCLC` (Node.js), `lighty-launcher` (Rust), `PrismLauncher` (C++).

## Instalación

```bash
go get github.com/ris-xeryx/minecraft_go_lib
```

## Ejemplo

### Obtener versiones disponibles

```go
package main

import (
	"fmt"
	mcgo "github.com/ris-xeryx/minecraft_go_lib"
)

func main() {
	versions, _ := mcgo.GetAllVersions()
	fmt.Printf("Hay %d versiones disponibles\n", len(versions))
	latest, _ := mcgo.GetLatestRelease()
	fmt.Println("Release más reciente:", latest)
}
```

### Instalar una versión

```go
bus := mcgo.NewEventBus()
bus.Subscribe(func(e mcgo.Event) {
	if e.Type == mcgo.EventDownloadProgress {
		fmt.Printf("\r%d / %d bytes", e.BytesLoaded, e.TotalBytes)
	}
})

_, err := mcgo.Install(mcgo.InstallParams{
	Version:    "1.21.4",
	Directory:  "/path/to/.minecraft",
	EventBus:   bus,
})
```

### Auth offline

```go
auth := mcgo.NewOfflineAuth("Player123")
profile, _ := auth.Authenticate()
fmt.Println(profile.Username, profile.UUID)
```

### Auth Microsoft (device code)

```go
auth := mcgo.NewMicrosoftAuth("your-azure-app-client-id")
auth.DeviceCodeCallback = func(code, url string) {
	fmt.Printf("Visita %s e introduce el código %s\n", url, code)
}
profile, _ := auth.Authenticate()
fmt.Println(profile.Username, profile.UUID)
```

### Lanzar Minecraft

```go
cmd, err := mcgo.Launch(mcgo.LaunchOptions{
	Instance: mcgo.Instance{
		Name:      "my-instance",
		Version:   "1.21.4",
		Loader:    mcgo.LoaderVanilla,
		Directory: "/path/to/.minecraft",
	},
	Profile: profile,
	Memory:  mcgo.LaunchMemory{Min: "2G", Max: "4G"},
})
if err != nil {
	panic(err)
}
cmd.Wait()
```

### Resolver versión de un loader

```go
fabricVer, _ := mcgo.LatestLoaderVersion(mcgo.LoaderFabric, "1.21.4")
fmt.Println("Fabric:", fabricVer)
```

## API

| Función | Descripción |
|---------|-------------|
| `GetAllVersions()` | Lista de versiones disponibles (Mojang) |
| `GetLatestRelease()` / `GetLatestSnapshot()` | Última release/snapshot |
| `GetVersionInfo(id)` | JSON completo de una versión |
| `Install(params)` | Descarga client.jar, libs, assets, natives |
| `Launch(opts)` | Lanza Minecraft y devuelve `*exec.Cmd` |
| `LatestLoaderVersion(loader, mcVer)` | Resuelve la versión más reciente del loader |
| `NewOfflineAuth(username)` | Auth offline (solo username) |
| `NewMicrosoftAuth(clientID)` | Auth Microsoft device code |
| `FindJava()` | Encuentra binario Java en el sistema |
| `NewEventBus()` | Sistema de eventos para progreso |

## Eventos

```go
const (
	EventInstallStarted EventType = iota
	EventFileChecked
	EventDownloadStarted
	EventDownloadProgress
	EventFileDownloaded
	EventInstallCompleted
	EventLaunchStarted
	EventProcessStarted
	EventProcessOutput
	EventProcessExited
	EventError
)
```

## Características

- ✅ Fetch de versions manifest de Mojang
- ✅ Auth offline (UUID igual que Mojang: MD5("OfflinePlayer:username"))
- ✅ Auth Microsoft device code flow (Azure → Xbox Live → XSTS → Minecraft)
- ✅ Instalación: client.jar, libraries, assets, natives (con rules por-OS)
- ✅ Resolución de loaders: Fabric, Quilt, Forge, NeoForge
- ✅ Detección de Java (JAVA_HOME, PATH, ubicaciones comunes)
- ✅ Event system con progreso de descarga
- ✅ Multi-OS (Windows, macOS, Linux)
- ✅ Verificación SHA1 de archivos
- ✅ Soporte para versiones `inheritsFrom` (hereda de padre)

## Licencia

[MIT](LICENSE)