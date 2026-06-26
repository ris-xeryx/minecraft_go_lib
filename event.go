package mcgo

// ── Event types ──────────────────────────────────────────────────────────

type EventType int

const (
	EvtInstallStarted EventType = iota
	EvtFileChecked
	EvtDownloadStarted
	EvtDownloadProgress
	EvtFileDownloaded
	EvtInstallCompleted
	EvtLaunchStarted
	EvtProcessStarted
	EvtError
)

type Event struct {
	Type        EventType
	TotalBytes  int64
	BytesLoaded int64
	Message     string
	Error       error
}

// ── EventBus ─────────────────────────────────────────────────────────────

type EventBus struct{ subs []func(Event) }

func NewEventBus() *EventBus { return &EventBus{} }

func (b *EventBus) On(fn func(Event)) { b.subs = append(b.subs, fn) }

func (b *EventBus) Emit(e Event) {
	for _, s := range b.subs {
		s(e)
	}
}

func (b *EventBus) emitInstallStarted(total int64) {
	b.Emit(Event{Type: EvtInstallStarted, TotalBytes: total})
}
func (b *EventBus) emitFileChecked(path string) {
	b.Emit(Event{Type: EvtFileChecked, Message: path})
}
func (b *EventBus) emitDownloadStarted(url string) {
	b.Emit(Event{Type: EvtDownloadStarted, Message: url})
}
func (b *EventBus) emitProgress(loaded, total int64) {
	b.Emit(Event{Type: EvtDownloadProgress, BytesLoaded: loaded, TotalBytes: total})
}
func (b *EventBus) emitFileDownloaded(path string) {
	b.Emit(Event{Type: EvtFileDownloaded, Message: path})
}
func (b *EventBus) emitInstallComplete() {
	b.Emit(Event{Type: EvtInstallCompleted})
}
func (b *EventBus) emitError(err error) {
	b.Emit(Event{Type: EvtError, Error: err})
}
