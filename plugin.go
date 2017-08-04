package main

import (
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

	Application     string
	EnvironmentName string
	VersionLabel    string
	Timeout         string
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
		"region":           p.Region,
		"application-name": p.Application,
		"environment":      p.EnvironmentName,
		"versionlabel":     p.VersionLabel,
		"timeout":          p.Timeout,
	}).Info("Attempting to check for deploy")

	// client
	return nil
}
