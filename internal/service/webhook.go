package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/wesleiramos/mockpay/internal/domain"
	"github.com/wesleiramos/mockpay/internal/store"
	"github.com/wesleiramos/mockpay/internal/util"
)

type retryItem struct {
	event   domain.WebhookEvent
	url     string
	secret  string
	attempt int
}

type WebhookService struct {
	store    *store.MemoryStore
	url      string
	secret   string
	events   chan domain.WebhookEvent
	retries  chan retryItem
	client   *http.Client
}

func NewWebhookService(s *store.MemoryStore, url, secret string) *WebhookService {
	ws := &WebhookService{
		store:   s,
		url:     url,
		secret:  secret,
		events:  make(chan domain.WebhookEvent, 100),
		retries: make(chan retryItem, 100),
		client:  &http.Client{Timeout: 5 * time.Second},
	}

	go ws.worker()
	go ws.retryWorker()

	return ws
}

func (ws *WebhookService) Dispatch(eventType domain.WebhookEventType, payload any) {
	if ws.url == "" {
		return
	}

	event := domain.WebhookEvent{
		ID:        util.NewID(),
		Type:      eventType,
		Payload:   payload,
		CreatedAt: domain.NowTimestamp(),
	}

	ws.events <- event
}

func (ws *WebhookService) worker() {
	for event := range ws.events {
		ws.send(event, ws.url, ws.secret, 1)
	}
}

func (ws *WebhookService) retryWorker() {
	for item := range ws.retries {
		time.Sleep(time.Duration(item.attempt) * 5 * time.Second)
		ws.send(item.event, item.url, item.secret, item.attempt)
	}
}

func (ws *WebhookService) send(event domain.WebhookEvent, url, secret string, attempt int) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("webhook marshal error: %v", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		log.Printf("webhook request error: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-MockPay-Signature", util.SignPayload(secret, payload))
	}

	resp, err := ws.client.Do(req)
	delivery := &domain.WebhookDelivery{
		ID:        util.NewID(),
		EventID:   event.ID,
		URL:       url,
		Attempt:   attempt,
		CreatedAt: domain.NowTimestamp(),
	}

	if err != nil {
		delivery.Success = false
		log.Printf("webhook delivery failed (attempt %d): %v", attempt, err)
	} else {
		delivery.StatusCode = resp.StatusCode
		delivery.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
		resp.Body.Close()

		if !delivery.Success {
			log.Printf("webhook returned status %d (attempt %d)", resp.StatusCode, attempt)
		}
	}

	ws.store.CreateDelivery(delivery)

	if !delivery.Success && attempt < 3 {
		ws.retries <- retryItem{
			event:   event,
			url:     url,
			secret:  secret,
			attempt: attempt + 1,
		}
	}
}
