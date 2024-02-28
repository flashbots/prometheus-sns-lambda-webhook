package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/flashbots/prometheus-sns-lambda-webhook/config"
	"github.com/flashbots/prometheus-sns-lambda-webhook/processor"
	"github.com/flashbots/prometheus-sns-lambda-webhook/secret"
	"github.com/urfave/cli/v2"
)

var (
	allowedWebhookMethods = []string{
		http.MethodGet,
		http.MethodPost,
	}
)

var (
	ErrSecretMissingKey    = errors.New("secret manager misses key")
	ErrWebhookInvalidUrl   = errors.New("webhook url is invalid")
	ErrWebhookUndefinedUrl = errors.New("webhook url is not defined")

	ErrWebhookInvalidMethod = fmt.Errorf(
		"invalid webhook method (allowed methods are: %s)",
		strings.Join(allowedWebhookMethods, ","),
	)
)

func CommandLambda(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "lambda",
		Usage: "Run lambda handler (default)",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &cfg.Webhook.Method,
				EnvVars:     []string{"WEBHOOK_METHOD"},
				Name:        "webhook-method",
				Usage:       "the `method` of the webhook",
				Value:       http.MethodGet,
			},

			&cli.BoolFlag{
				Destination: &cfg.Webhook.IncludeBody,
				EnvVars:     []string{"WEBHOOK_INCLUDE_BODY"},
				Name:        "webhook-include-body",
				Usage:       "whether to relay the alert in webhook's request body",
				Value:       false,
			},

			&cli.StringFlag{
				Destination: &cfg.Webhook.Url,
				EnvVars:     []string{"WEBHOOK_URL"},
				Name:        "webhook-url",
				Usage:       "the `url` of the webhook",
			},
		},

		Before: func(_ *cli.Context) error {
			// read secrets (if applicable)
			if strings.HasPrefix(cfg.Webhook.Url, "arn:aws:secretsmanager:") {
				s, err := secret.AWS(cfg.Webhook.Url)
				if err != nil {
					return err
				}
				webhookUrl, exists := s["WEBHOOK_URL"]
				if !exists {
					return fmt.Errorf("%w: %s: %s",
						ErrSecretMissingKey, cfg.Webhook.Url, "WEBHOOK_URL",
					)
				}
				cfg.Webhook.Url = webhookUrl
			}

			// validate inputs
			if cfg.Webhook.Url == "" {
				return ErrWebhookUndefinedUrl
			}
			if url, err := url.ParseRequestURI(cfg.Webhook.Url); err != nil {
				return fmt.Errorf("%w: %w",
					ErrWebhookInvalidUrl, err,
				)
			} else {
				fmt.Println(url)

			}
			for _, allowedMethod := range allowedWebhookMethods {
				if cfg.Webhook.Method == allowedMethod {
					return nil
				}
			}
			return fmt.Errorf("%w: %s",
				ErrWebhookInvalidMethod, cfg.Webhook.Method,
			)
		},

		Action: func(_ *cli.Context) error {
			p, err := processor.New(cfg)
			if err != nil {
				return err
			}
			awslambda.Start(p.Lambda)
			return nil
		},
	}
}
