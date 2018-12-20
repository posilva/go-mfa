package session

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"sync"
)

// HandlerFunc defines a function to handle session map foreach
type HandlerFunc func(string, string, *AWSSession) error

// AWSSession a aws session created from the MFA session
type AWSSession struct {
	session *session.Session
}

// Get returns the AWS SDK session
func (as *AWSSession) Get() *session.Session {
	return as.session
}

// MFASession the main mfa session data
type MFASession struct {
	session        *session.Session
	params         *Params
	cachedSessions session.Session
}

// NewMFASession creates a new session using mfa
func NewMFASession(p Params) *MFASession {
	ss := MFASession{
		params: &p,
	}
	ss.initWithToken()
	return &ss
}

// Assume creates a custom session using the provided role
func (s *MFASession) Assume(role string, region string) (*AWSSession, error) {
	creds, err := assumeRole(s.session, role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}
	config := aws.NewConfig().WithRegion(region).WithCredentials(fromStsCredentials(creds))
	return &AWSSession{
		session: session.Must(session.NewSessionWithOptions(session.Options{Config: *config})),
	}, nil
}

// AssumeBulk creates several custom sessions from the MFA Session
func (s *MFASession) AssumeBulk(roleName string, accounts []string) (*Map, error) {
	return s.AssumeBulkWithRegions(roleName, accounts, DefaultRegions())
}

// AssumeBulkWithRegions creates a custom session from the main session
func (s *MFASession) AssumeBulkWithRegions(roleName string, accounts []string, regions []string) (*Map, error) {
	var sm Map
	for _, account := range accounts {
		role := "arn:aws:iam::" + account + ":role/" + roleName
		sm.Ensure(account)
		var rwg sync.WaitGroup
		rwg.Add(len(regions))
		for _, region := range regions {
			errs := make(chan error, 1)
			go func(a string, r string, errs chan<- error) {
				s, err := s.Assume(role, r)
				if err != nil {
					errs <- err
				}
				sm.Put(a, r, s)
				rwg.Done()
			}(account, region, errs)
		}
		rwg.Wait()

	}
	return &sm, nil
}

func (s *MFASession) initWithToken() error {
	creds, err := s.getMFACredentials()
	if err != nil {
		return fmt.Errorf("failed to create a session token: %v", err)
	}
	s.session = session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String("us-west-2"),
			Credentials: fromStsCredentials(creds),
		}}))
	return nil
}

func (s *MFASession) getMFACredentials() (*sts.Credentials, error) {
	sess := session.Must(session.NewSessionWithOptions(s.params.ToOptions()))

	svc := sts.New(sess)
	input := &sts.GetSessionTokenInput{
		TokenCode:       aws.String(s.params.MFAToken),
		DurationSeconds: aws.Int64(s.params.MFADuration),
		SerialNumber:    aws.String(s.params.SerialDevice),
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if err != nil {
				return nil, errors.Wrap(aerr, "failed to get a session token")
			}
		} else {
			if err != nil {
				return nil, errors.Wrap(err, "failed to get a session token")
			}
		}
	}
	return result.Credentials, nil
}

func fromStsCredentials(c *sts.Credentials) *credentials.Credentials {
	return credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     *c.AccessKeyId,
		SecretAccessKey: *c.SecretAccessKey,
		SessionToken:    *c.SessionToken,
	})
}

func assumeRole(session *session.Session, roleArn string) (*sts.Credentials, error) {
	svc := sts.New(session)
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(3600),
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("Gomonit"),
	}
	result, err := svc.AssumeRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeMalformedPolicyDocumentException:
				return nil, fmt.Errorf("failed to assume role [ErrCodeMalformedPolicyDocumentException]: %v", aerr.Error())
			case sts.ErrCodePackedPolicyTooLargeException:
				return nil, fmt.Errorf("failed to assume role [ErrCodePackedPolicyTooLargeException]: %v", aerr.Error())
			case sts.ErrCodeRegionDisabledException:
				return nil, fmt.Errorf("failed to assume role [ErrCodeRegionDisabledException]: %v", aerr.Error())
			default:
				return nil, fmt.Errorf("failed to assume role: %v", aerr.Error())
			}
		}
		return nil, fmt.Errorf("failed to assume role: %v", err.Error())
	}
	return result.Credentials, nil
}
