package scheduler

import (
	"time"

	"github.com/sirupsen/logrus"

	"bank-service/internal/service"
)

type CreditScheduler struct {
	creditService service.CreditService
	logger        *logrus.Logger
	stopCh        chan struct{}
}

func NewCreditScheduler(creditService service.CreditService, logger *logrus.Logger) *CreditScheduler {
	return &CreditScheduler{
		creditService: creditService,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

func (s *CreditScheduler) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Credit scheduler started")

	s.processPayments()

	for {
		select {
		case <-ticker.C:
			s.processPayments()
		case <-s.stopCh:
			s.logger.Info("Credit scheduler stopped")
			return
		}
	}
}

func (s *CreditScheduler) Stop() {
	close(s.stopCh)
}

func (s *CreditScheduler) processPayments() {
	s.logger.Info("Processing pending credit payments")

	if err := s.creditService.ProcessPendingPayments(); err != nil {
		s.logger.Errorf("Error processing pending payments: %v", err)
	} else {
		s.logger.Info("Pending payments processed successfully")
	}
}
