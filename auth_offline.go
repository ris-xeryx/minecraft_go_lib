package mcgo

// Authenticator es una interfaz para todos los métodos de autenticación.
type Authenticator interface {
	Authenticate() (Profile, error)
}

// AuthOffline crea un perfil offline usando solo el username.
type OfflineAuth struct {
	Username string
}

func NewOfflineAuth(username string) *OfflineAuth {
	return &OfflineAuth{Username: username}
}

func (a *OfflineAuth) Authenticate() (Profile, error) {
	return Profile{
		Username:   a.Username,
		UUID:       offlineUUID(a.Username),
		PlayerName: a.Username,
	}, nil
}

// offlineUUID genera un UUID using FDCE (FNV) hash offline (igual que Mojang).
func offlineUUID(username string) string {
	// Java UUID offline: MD5("OfflinePlayer:" + username)
	// Aquí uso un hash simple para mantener offline sin dependencias extras.
	// NOTA: Para compatibilidad con servidores que esperan UUID específico,
	// usa MD5("OfflinePlayer:"+username) — implementado abajo.
	return javaOfflineUUID(username)
}

func javaOfflineUUID(username string) string {
	// MD5 Like-Java's UUID.nameUUIDFromBytes(("OfflinePlayer:"+username).getBytes(UTF_8))
	// Versión 3 UUID basado en MD5, igual que Minecraft offline.
	const (
		versionMask = 0x0f // bits 12-15 = 0x3 (version 3)
		variantMask = 0x3f // bits 6-7 = 0b10
	)
	hash := md5Bytes([]byte("OfflinePlayer:" + username))
	hash[6] = (hash[6] & ^byte(versionMask)) | 0x30
	hash[8] = (hash[8] & ^byte(variantMask)) | 0x80
	return formatUUID(hash)
}
