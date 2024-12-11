package configuration

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"

	"verification-service/internal/library/server/telemetry"
)

// https://aws.github.io/aws-sdk-go-v2/docs/handling-errors/

// Region is an overload for Configuration -- changing the default region.
func Region(ctx context.Context, region string) aws.Config {
	if region == "" {
		slog.ErrorContext(ctx, "Fatal - Invalid, Empty Region")

		panic("unexpected, empty & invalid region provided as input")
	}

	configuration, e := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithDefaultsMode(aws.DefaultsModeStandard),
		config.WithLogConfigurationWarnings(true),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequestEventMessage),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(10),
		config.WithCredentialsCacheOptions(func(o *aws.CredentialsCacheOptions) { o.ExpiryWindow = 10 * time.Minute }),
		config.WithHTTPClient(telemetry.Client(map[string]string{})),
	)

	if e != nil {
		var ae smithy.APIError
		var oe *smithy.OperationError

		switch {
		case errors.As(e, &ae):
			slog.ErrorContext(ctx, "Failed Defining AWS Configuration for Region (AE)", slog.String("region", region), slog.String("code", ae.ErrorCode()), slog.Any("fault", ae.ErrorFault()), slog.String("message", ae.ErrorMessage()), slog.String("error", ae.Error()))
			panic(e)
		case errors.As(e, &oe):
			slog.ErrorContext(ctx, "Failed Defining AWS Configuration for Region (OE)", slog.String("region", region), slog.String("operation", oe.Operation()), slog.String("service", oe.Service()), slog.String("error", oe.Error()), slog.Any("unwrap", oe.Unwrap()))
			panic(e)
		default:
			slog.ErrorContext(ctx, "Failed Defining AWS Configuration for Region (Unknown)", slog.String("region", region), slog.String("error", e.Error()))
			panic(e)
		}
	}

	return configuration
}

// Configuration - AWS-related Authentication settings
func Configuration(ctx context.Context) aws.Config {
	configuration, e := config.LoadDefaultConfig(ctx,
		config.WithDefaultsMode(aws.DefaultsModeStandard),
		config.WithLogConfigurationWarnings(true),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequestEventMessage),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(10),
		config.WithCredentialsCacheOptions(func(o *aws.CredentialsCacheOptions) { o.ExpiryWindow = 10 * time.Minute }),
		config.WithHTTPClient(telemetry.Client(map[string]string{})),
	)

	if e != nil {
		var ae smithy.APIError
		var oe *smithy.OperationError

		switch {
		case errors.As(e, &ae):
			slog.ErrorContext(ctx, "Failed Defining AWS Configuration (AE)", slog.String("code", ae.ErrorCode()), slog.Any("fault", ae.ErrorFault()), slog.String("message", ae.ErrorMessage()), slog.String("error", ae.Error()))
			panic(e)
		case errors.As(e, &oe):
			slog.ErrorContext(ctx, "Failed Defining AWS Configuration (OE)", slog.String("operation", oe.Operation()), slog.String("service", oe.Service()), slog.String("error", oe.Error()), slog.Any("unwrap", oe.Unwrap()))
			panic(e)
		default:
			slog.ErrorContext(ctx, "Failed Defining AWS Configuration (Unknown)", slog.String("error", e.Error()))
			panic(e)
		}
	}

	return configuration
}
