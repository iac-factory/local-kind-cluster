package mail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/smithy-go"

	"user-service/internal/library/mail/internal/configuration"
)

func Verification(ctx context.Context, recipient string, code string) error {
	const (
		sender  = "no-reply@ethr.gg"
		subject = "ETHR - Verify Email Address"
		set     = "ethr-email-verification-configuration-set"
	)

	var html, text bytes.Buffer

	settings := configuration.Region(ctx, "us-east-2")

	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)

	log := slog.Group("input",
		slog.String("sender", sender),
		slog.String("subject", subject),
		slog.String("timestamp", timestamp),
		slog.String("recipient", recipient),
		slog.String("region", settings.Region),
	)

	slog.DebugContext(ctx, "Email Verification Metadata", log)

	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}

	metadata := Metadata{24, "hours", fmt.Sprintf("%s/verify/%s", frontend, code)}

	if e := HTML.Execute(&html, metadata); e != nil {
		slog.ErrorContext(ctx, "Unable to Generate HTML Email Template", slog.String("error", e.Error()))

		return e
	}

	if e := Text.Execute(&text, metadata); e != nil {
		slog.ErrorContext(ctx, "Unable to Generate Text Email Template", slog.String("error", e.Error()))

		return e
	}

	// Create the send email input
	input := &ses.SendEmailInput{
		Source: aws.String(sender),
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		ReplyToAddresses:     []string{},
		ReturnPath:           nil,
		ConfigurationSetName: aws.String(set),
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(html.String()),
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(text.String()),
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Tags: []types.MessageTag{
			{
				Name:  aws.String("Type"),
				Value: aws.String("User-Email-Verification"),
			},
			{
				Name:  aws.String("Timestamp"),
				Value: aws.String(timestamp),
			},
		},
	}

	client := ses.NewFromConfig(settings)

	result, e := client.SendEmail(ctx, input)
	if e != nil {
		var ae smithy.APIError
		var oe *smithy.OperationError

		switch {
		case errors.As(e, &ae):
			slog.ErrorContext(ctx, "Failed Submitting Verification Email (AE)", slog.String("code", ae.ErrorCode()), slog.Any("fault", ae.ErrorFault()), slog.String("message", ae.ErrorMessage()), slog.String("error", ae.Error()))
			return e
		case errors.As(e, &oe):
			slog.ErrorContext(ctx, "Failed Submitting Verification Email (OE)", slog.String("operation", oe.Operation()), slog.String("service", oe.Service()), slog.String("error", oe.Error()), slog.Any("unwrap", oe.Unwrap()))
			return e
		default:
			slog.ErrorContext(ctx, "Failed Submitting Verification Email (Unknown)", slog.String("error", e.Error()))
			return e
		}
	}

	slog.InfoContext(ctx, "Email Successfully Submitted", slog.String("message-id", aws.ToString(result.MessageId)))

	return nil
}
