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

// Exec runs the plugin
func (p *Plugin) Exec() error {
	// create the client

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
	}).Info("attempting to check for version deploy")

	timeout := time.After(p.Timeout)
	tick := time.Tick(1 * time.Second)

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
			version, err := client.DescribeApplicationVersions(
				&elasticbeanstalk.DescribeApplicationVersionsInput{
					ApplicationName: aws.String(p.Application),
					VersionLabels:   aws.StringSlice([]string{p.VersionLabel}),
				},
			)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("problem retrieving application version information")
				return err
			}

			status := aws.StringValue(version.ApplicationVersions[0].Status)

			switch status {
			case elasticbeanstalk.ApplicationVersionStatusProcessed:
				return nil
			case elasticbeanstalk.ApplicationVersionStatusFailed:
				return errors.New("application version deploy failed")
			default:
				log.WithFields(log.Fields{
					"status": status,
				}).Error("waiting for application deploy to finish")
			}
		}
	}
}
