package fcm

import (
	"context"
	"fmt"

	"fnpush/model"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"

	"google.golang.org/api/option"
)

const MaxAllowedRecipients = 500

type FcmClient interface {
	Push(p *model.PushRequest) (*model.PushResponse, error)
}

type fcmClient struct {
	messaging *messaging.Client
}

func NewClient(credentials []byte) (FcmClient, error) {

	options := option.WithCredentialsJSON(credentials)
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, nil, options)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize Firebase application: %v", err)
	}

	messaging, err := app.Messaging(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot create Firebase Cloud Messaging client: %v", err)
	}

	return &fcmClient{
		messaging: messaging,
	}, nil
}

func (c *fcmClient) Push(p *model.PushRequest) (*model.PushResponse, error) {

	messages := make([]*messaging.Message, 0)
	notifications := make(map[int]*model.Notification)

	for i, token := range p.Recipients {

		message := messaging.Message{
			Token: token,
			Data:  p.Data,
		}

		if p.Message != nil {

			message.Notification = &messaging.Notification{
				Title: p.Message.Title,
				Body:  p.Message.Body,
			}
		}

		messages = append(messages, &message)

		notifications[i] = &model.Notification{
			Recipient: token,
			Data:      p.Data,
			Message:   p.Message,
		}
	}

	batch, err := c.messaging.SendAll(context.Background(), messages)

	if err != nil {
		return nil, err
	}

	response := new(model.PushResponse)
	response.NotificationsCount = batch.SuccessCount
	response.ErrorsCount = batch.FailureCount

	for i, r := range batch.Responses {

		if r.Success {
			notifications[i].Id = r.MessageID
		} else {

			// Add current response error to notification errors list
			response.Errors = append(response.Errors, model.NotificationError{
				Recipient: notifications[i].Recipient,
				Message:   r.Error.Error(),
			})

			// Remove failed notification to successful notifications list
			delete(notifications, i)
		}
	}

	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, *notification)
	}

	return response, nil
}
