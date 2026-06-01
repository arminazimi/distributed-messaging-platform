package operatorA

import (
	"context"
	"gateway/app"
	"gateway/internal/model"
)

type OA struct{}

func (o OA) Send(ctx context.Context, s model.Message) error {

	//for test circuit breaker
	//return errors.New("fall down")

	for _, v := range s.Recipients {
		app.Logger.Info("your message has sent ",
			"user id : ", s.CustomerID,
			"msg : ", s.Text,
			"number : ", v,
			"operator:", "A")
	}

	return nil
}
