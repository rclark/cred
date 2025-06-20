package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
)

var profile string

const (
	accessKeyID      = "AWS_ACCESS_KEY_ID"
	secretAccessKey  = "AWS_SECRET_ACCESS_KEY"
	sessionToken     = "AWS_SESSION_TOKEN"
	sessionExpiresAt = "AWS_SESSION_EXPIRES_AT"
	accountID        = "AWS_ACCOUNT_ID"
	defaultRegion    = "AWS_DEFAULT_REGION"
	region           = "AWS_REGION"
)

func allVars() []string {
	return []string{
		accessKeyID,
		secretAccessKey,
		sessionToken,
		sessionExpiresAt,
		accountID,
		defaultRegion,
		region,
	}
}

func set(key, val string) string {
	return fmt.Sprintf("%s=%s", key, val)
}

func unset(key string) string {
	return fmt.Sprintf("unset %s;", key)
}

func getCallerIdentity(ctx context.Context, cfg aws.Config) (*sts.GetCallerIdentityOutput, error) {
	data, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &apiErr) {
			return data, fmt.Errorf("Invalid credentials: %s: %s", apiErr.Code, apiErr.Message)
		}
		return data, fmt.Errorf("Invalid credentials: %w", err)
	}

	return data, nil
}

var rootCmd = &cobra.Command{
	Use:   "cred",
	Short: "Fetch AWS credentials and set them as environment variables",
	Long:  wordwrap.WrapString("Fetch AWS credentials and set them as environment variables.\n\nEvaluate the output of the command in order to export AWS credentials as environment variables, e.g. $(cred) or eval $(cred).", 80),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		opts := []func(*config.LoadOptions) error{}
		if profile != "" {
			opts = append(opts, config.WithSharedConfigProfile(profile))
		}

		for _, key := range allVars() {
			os.Setenv(key, "")
		}

		cfg, err := config.LoadDefaultConfig(ctx, opts...)
		if err != nil {
			return err
		}

		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return err
		}

		data, err := getCallerIdentity(ctx, cfg)
		if err != nil {
			return err
		}

		unsets := []string{}

		exports := []string{
			set(accessKeyID, creds.AccessKeyID),
			set(secretAccessKey, creds.SecretAccessKey),
		}

		if creds.AccountID != "" {
			exports = append(exports, set(accountID, creds.AccountID))
		} else {
			exports = append(exports, set(accountID, *data.Account))
		}

		if cfg.Region != "" {
			exports = append(exports, set(defaultRegion, cfg.Region))
			exports = append(exports, set(region, cfg.Region))
		} else {
			unsets = append(unsets, unset(defaultRegion))
			unsets = append(unsets, unset(region))
		}

		if creds.SessionToken != "" {
			exports = append(
				exports,
				set(sessionToken, creds.SessionToken),
				set(sessionExpiresAt, creds.Expires.Format(time.RFC3339)),
			)
		} else {
			unsets = append(
				unsets,
				unset(sessionToken),
				unset(sessionExpiresAt),
			)
		}

		output := fmt.Sprintf("export %s\n", strings.Join(exports, " "))
		if len(unsets) > 0 {
			output = fmt.Sprintf("%s\n%s", strings.Join(unsets, "\n"), output)
		}
		fmt.Print(output)

		return nil
	},
}

var expiryCmd = &cobra.Command{
	Use:     "expiry",
	Short:   "Print the time that explicit environment credentials will expire",
	Aliases: []string{"exp", "expires", "expire"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case os.Getenv("AWS_ACCESS_KEY_ID") == "":
			return fmt.Errorf("AWS credentials are not set as environment variables")
		case os.Getenv("AWS_SESSION_TOKEN") == "":
			return fmt.Errorf("AWS credentials in environment variables are not temporary")
		case os.Getenv("AWS_SESSION_EXPIRES_AT") == "":
			return fmt.Errorf("AWS credentials expiration time has not been recorded in your environment")
		default:
			expires, err := time.Parse(time.RFC3339, os.Getenv("AWS_SESSION_EXPIRES_AT"))
			if err != nil {
				return fmt.Errorf("AWS credentials expiration time has not been properly recorded in your environment")
			}
			fmt.Println(expires.Local().Format(time.RFC1123))
			return nil
		}
	},
}

var clearCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Clear AWS environment variables",
	Long:    wordwrap.WrapString("Clear AWS environment variables.\n\nEvaluate the output of the command in order to export AWS credentials as environment variables, e.g. $(cred clear) or eval $(cred clear).", 80),
	Aliases: []string{"unset", "rm", "none"},
	RunE: func(cmd *cobra.Command, args []string) error {
		unsets := []string{}
		for _, key := range allVars() {
			unsets = append(unsets, unset(key))
		}

		fmt.Println(strings.Join(unsets, "\n"))
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&profile, "profile", "", "AWS profile to use")

	rootCmd.AddCommand(expiryCmd)
	rootCmd.AddCommand(clearCmd)
}
