package mcgo

import (
	"fmt"
)

// GetLoaderVersions obtiene las versiones disponibles de un loader para una versión de MC.
func GetLoaderVersions(loader LoaderType, mcVersion string) ([]string, error) {
	switch loader {
	case LoaderFabric:
		return getFabricVersions(mcVersion)
	case LoaderQuilt:
		return getQuiltVersions(mcVersion)
	case LoaderForge:
		return getForgeVersions(mcVersion)
	case LoaderNeoForge:
		return getNeoForgeVersions(mcVersion)
	case LoaderVanilla:
		return []string{""}, nil
	default:
		return nil, fmt.Errorf("unknown loader: %s", loader)
	}
}

// LoaderVersions返回 el primer (más reciente) loader version para una MC version.
func LatestLoaderVersion(loader LoaderType, mcVersion string) (string, error) {
	versions, err := GetLoaderVersions(loader, mcVersion)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no %s version for %s", loader, mcVersion)
	}
	return versions[0], nil
}

// ListLoaders devuelve todos los loaders soportados.
func ListLoaders() []LoaderType {
	return []LoaderType{LoaderVanilla, LoaderFabric, LoaderQuilt, LoaderNeoForge, LoaderForge}
}
