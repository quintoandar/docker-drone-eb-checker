package main

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

// Plugin defines the beanstalk plugin parameters.
type Plugin struct {
	Key    string
	Secret string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string

	Application  string
	VersionLabel string
	Timeout      time.Duration
}

type logger struct {
	env    string
	status string
	health string
}

func (l logger) Info(msg string) {
	log.WithFields(log.Fields{
		"environment": l.env,
		"status":      l.status,
		"health":      l.health,
	}).Info(msg)
}

func (l logger) Error(msg string) {
	log.WithFields(log.Fields{
		"environment": l.env,
		"status":      l.status,
		"health":      l.health,
	}).Error(msg)
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	conf := &aws.Config{
		Region: aws.String(p.Region),
	}

	// Use key and secret if provided otherwise fall back to ec2 instance profile
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	}

	client := elasticbeanstalk.New(session.New(), conf)

	log.WithFields(log.Fields{
		"region":        p.Region,
		"application":   p.Application,
		"version-label": p.VersionLabel,
		"timeout":       p.Timeout,
	}).Info("attempting to check for a successful deploy")

	timeout := time.After(p.Timeout)
	tick := time.Tick(10 * time.Second)

	for {
		select {

		case <-timeout:
			err := errors.New("timed out")

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("problem retrieving application version information")
				return err
			}

		case <-tick:
			envs, err := client.DescribeEnvironments(
				&elasticbeanstalk.DescribeEnvironmentsInput{
					ApplicationName: aws.String(p.Application),
				},
			)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("problem retrieving environment information")
				return err
			}

			for _, env := range envs.Environments {

				l := logger{
					env:    aws.StringValue(env.EnvironmentName),
					status: aws.StringValue(env.Status),
					health: aws.StringValue(env.HealthStatus),
				}

				label := aws.StringValue(env.VersionLabel)
				status := aws.StringValue(env.Status)
				health := aws.StringValue(env.HealthStatus)

				if label != p.VersionLabel {
					l.Info("environment is updating")
					continue
				}

				if status != elasticbeanstalk.EnvironmentStatusReady {
					l.Info("label is correct but environment is not ready yet")
					continue
				}

				if health != elasticbeanstalk.EnvironmentHealthStatusOk {
					l.Info("environment is ready but health status is not ok")
					continue
				}

				l.Info("environment deployment was successful")
				return nil
			}
		}
	}
}
