package worker

import (
	"context"
	"log"
	"time"

	"github.com/biyonik/goshort/internal/domain"
)

// ClickProcessor click event'lerini işler
type ClickProcessor struct {
	channel   chan domain.ClickEvent
	batchSize int
	flushTime time.Duration
	saveFn    func(ctx context.Context, events []domain.ClickEvent) error
	stopCh    chan struct{}
}

// NewClickProcessor yeni bir click processor oluşturur
func NewClickProcessor(
	bufferSize int, // Channel kapasitesi (örn: 10000)
	batchSize int, // Kaç event'te bir DB'ye yaz (örn: 100)
	flushTime time.Duration, // Ne kadar sürede bir flush (örn: 5s)
	saveFn func(ctx context.Context, events []domain.ClickEvent) error,
) *ClickProcessor {
	return &ClickProcessor{
		channel:   make(chan domain.ClickEvent, bufferSize),
		batchSize: batchSize,
		flushTime: flushTime,
		saveFn:    saveFn,
		stopCh:    make(chan struct{}),
	}
}

// Push event'i channel'a gönderir (non-blocking)
func (p *ClickProcessor) Push(event domain.ClickEvent) {
	select {
	case p.channel <- event:
		// Başarıyla gönderildi
	default:
		// Channel dolu, event kayboldu (log'la)
		log.Printf("WARNING: Click channel full, dropping event for %s", event.URLCode)
	}
}

// Start worker'ı başlatır
func (p *ClickProcessor) Start(ctx context.Context) {
	go p.run(ctx)
	log.Println("Click processor started")
}

// Stop worker'ı durdurur
func (p *ClickProcessor) Stop() {
	close(p.stopCh)
	log.Println("Click processor stopped")
}

// run ana worker döngüsü
func (p *ClickProcessor) run(ctx context.Context) {
	batch := make([]domain.ClickEvent, 0, p.batchSize)
	ticker := time.NewTicker(p.flushTime)
	defer ticker.Stop()

	for {
		select {
		case event := <-p.channel:
			// Event geldi, batch'e ekle
			batch = append(batch, event)

			// Batch doldu mu?
			if len(batch) >= p.batchSize {
				p.flush(ctx, &batch)
			}

		case <-ticker.C:
			// Zaman doldu, batch'te ne varsa yaz
			if len(batch) > 0 {
				p.flush(ctx, &batch)
			}

		case <-p.stopCh:
			// Durduruluyor, kalan event'leri yaz
			if len(batch) > 0 {
				p.flush(ctx, &batch)
			}
			return

		case <-ctx.Done():
			// Context iptal edildi
			if len(batch) > 0 {
				p.flush(ctx, &batch)
			}
			return
		}
	}
}

// flush batch'i DB'ye yazar ve temizler
func (p *ClickProcessor) flush(ctx context.Context, batch *[]domain.ClickEvent) {
	if len(*batch) == 0 {
		return
	}

	if err := p.saveFn(ctx, *batch); err != nil {
		log.Printf("ERROR: Failed to save click batch: %v", err)
		// Hata olsa bile batch'i temizle (kayıp kabul edilebilir)
	} else {
		log.Printf("Flushed %d click events", len(*batch))
	}

	// Batch'i temizle (capacity'yi koru)
	*batch = (*batch)[:0]
}
