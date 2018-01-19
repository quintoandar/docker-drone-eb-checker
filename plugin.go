package main

import (
	"errors"
	"fmt"
	"strings"
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
	Environment  string
	VersionLabel string

	Debug   bool
	Timeout time.Duration
	Tick    time.Duration
}

type logger struct {
	env     string
	status  string
	health  string
	version string
}

func (l logger) Info(msg string) {
	l.fields().Info(msg)
}

func (l logger) Warn(msg string) {
	l.fields().Warn(msg)
}

func (l logger) Error(msg string) {
	l.fields().Error(msg)
}

func (l logger) fields() *log.Entry {
	return log.WithFields(log.Fields{
		"env":     l.env,
		"status":  l.status,
		"health":  l.health,
		"version": l.version,
	})
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	if p.Debug {
		log.SetLevel(log.DebugLevel)
	}

	conf := &aws.Config{
		Region: aws.String(p.Region),
	}

	// Use key and secret if provided otherwise fall back to ec2 instance profile
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	}

	client := elasticbeanstalk.New(session.New(), conf)

	fields := log.Fields{
		"region":  p.Region,
		"app":     p.Application,
		"label":   p.VersionLabel,
		"timeout": p.Timeout,
		"tick":    p.Tick,
	}

	if len(p.Environment) != 0 {
		fields["env"] = p.Environment
	}

	log.WithFields(fields).Info("attempting to check for a successful deploy")

	timeout := time.After(p.Timeout)
	tick := time.Tick(p.Tick)

	for {
		select {

		case <-timeout:
			err := errors.New("timed out")

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("could not identify a successful deploy in time")
				return err
			}

		case <-tick:

			log.WithFields(fields).Debug("ticking")

			envs, err := p.getEnvironments(client)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("problem retrieving environment information")
				return err
			}

			if len(envs.Environments) == 0 {
				err := fmt.Errorf("application %s environment [%s] not found", p.Application, p.Environment)
				log.WithFields(log.Fields{
					"error": err,
				}).Error("problem retrieving environment information")
				return err
			}

			for _, env := range envs.Environments {

				label := aws.StringValue(env.VersionLabel)
				status := aws.StringValue(env.Status)
				health := aws.StringValue(env.Health)

				l := logger{
					env:     aws.StringValue(env.EnvironmentName),
					status:  status,
					health:  health,
					version: label,
				}

				if !strings.HasPrefix(p.VersionLabel, label) {
					l.Info("environment is updating")
					continue
				}

				if status != elasticbeanstalk.EnvironmentStatusReady {
					l.Warn("environment is not ready")
					continue
				}

				if health != elasticbeanstalk.EnvironmentHealthGreen {
					l.Warn("environment health is not ok")
					continue
				}

				l.Info("environment deployment was successful")
				return nil
			}
		}
	}
}

func (p *Plugin) getEnvironments(client *elasticbeanstalk.ElasticBeanstalk) (*elasticbeanstalk.EnvironmentDescriptionsMessage, error) {
	if p.Environment != "" {
		return client.DescribeEnvironments(
			&elasticbeanstalk.DescribeEnvironmentsInput{
				ApplicationName:  aws.String(p.Application),
				EnvironmentNames: []*string{aws.String(p.Environment)},
			},
		)
	}

	return client.DescribeEnvironments(
		&elasticbeanstalk.DescribeEnvironmentsInput{
			ApplicationName: aws.String(p.Application),
		},
	)
}
