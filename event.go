// Package mcgo provides functions for installing and launching Minecraft.

package mcgo

import "fmt"

// EventType代表安装/启动过程中可能发生的事件类型
type EventType int

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

func (e EventType) String() string {
	switch e {
	case EventInstallStarted:
		return "InstallStarted"
	case EventFileChecked:
		return "FileChecked"
	case EventDownloadStarted:
		return "DownloadStarted"
	case EventDownloadProgress:
		return "DownloadProgress"
	case EventFileDownloaded:
		return "FileDownloaded"
	case EventInstallCompleted:
		return "InstallCompleted"
	case EventLaunchStarted:
		return "LaunchStarted"
	case EventProcessStarted:
		return "ProcessStarted"
	case EventProcessOutput:
		return "ProcessOutput"
	case EventProcessExited:
		return "ProcessExited"
	case EventError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", int(e))
	}
}

// Event es un evento del flujo de instalación/lanzamiento.
type Event struct {
	Type        EventType
	TotalBytes  int64
	Bytes       int64
	BytesLoaded int64
	Message     string
	Error       error
}

// EventCallback es la firma de la función que recibe eventos.
type EventCallback func(Event)

// EventBus distribuye eventos a suscriptores.
type EventBus struct {
	subs []EventCallback
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (b *EventBus) Subscribe(cb EventCallback) {
	b.subs = append(b.subs, cb)
}

func (b *EventBus) Emit(e Event) {
	for _, s := range b.subs {
		s(e)
	}
}

// Helpers para emitir eventos comunes.
func (b *EventBus) emitInstallStarted(total int64) {
	b.Emit(Event{Type: EventInstallStarted, TotalBytes: total})
}

func (b *EventBus) emitFileChecked(path string) {
	b.Emit(Event{Type: EventFileChecked, Message: path})
}

func (b *EventBus) emitDownloadStarted(url string) {
	b.Emit(Event{Type: EventDownloadStarted, Message: url})
}

func (b *EventBus) emitProgress(loaded, total int64) {
	b.Emit(Event{
		Type:        EventDownloadProgress,
		Bytes:       loaded,
		TotalBytes:  total,
		BytesLoaded: loaded,
	})
}

func (b *EventBus) emitFileDownloaded(path string) {
	b.Emit(Event{Type: EventFileDownloaded, Message: path})
}

func (b *EventBus) emitInstallComplete() {
	b.Emit(Event{Type: EventInstallCompleted})
}

func (b *EventBus) emitError(err error) {
	b.Emit(Event{Type: EventError, Error: err})
}
