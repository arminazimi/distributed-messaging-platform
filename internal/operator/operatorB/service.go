package operatorB

import (
	"context"
	"gateway/app"
	"gateway/internal/model"
)

type OB struct{}

func (o OB) Send(ctx context.Context, s model.Message) error {

	//for test refund
	//return errors.New("fall down")

	for _, v := range s.Recipients {
		app.Logger.Info("your message has sent ",
			"user id : ", s.CustomerID,
			"msg : ", s.Text,
			"number : ", v,
			"operator:", "B")
	}

	return nil
}
