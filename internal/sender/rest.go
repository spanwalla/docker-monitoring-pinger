package sender

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spanwalla/docker-monitoring-pinger/internal/auth"
	"github.com/spanwalla/docker-monitoring-pinger/internal/entity"
)

type RestSender struct {
	client *resty.Client
	auth   *auth.Client
}

func NewRestSender(baseURL string, auth *auth.Client) *RestSender {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetAuthScheme("Bearer")
	return &RestSender{client: client, auth: auth}
}

func (r *RestSender) Send(reports []entity.Report) error {
	token, err := r.auth.GetToken()
	if err != nil {
		if errors.Is(err, auth.ErrTryAgain) {
			log.Info("RestSender.Send - auth.GetToken: registered, will try again")
			return nil
		}

		log.Errorf("RestSender.Send - auth.GetToken: %v", err)
		return err
	}

	resp, err := r.client.R().
		SetAuthToken(token).
		SetBody(map[string][]entity.Report{"report": reports}).
		Post("/api/v1/reports")
	if err != nil {
		log.Errorf("RestSender.Send - post: %v", err)
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to send data: status %v", resp.Status())
	}

	return nil
}
