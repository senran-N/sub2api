package service

import (
	"crypto/tls"
	"io"
	"net"
	"net/smtp"
	"strings"
	"testing"
)

type smtpClientStub struct {
	startTLSSupported bool
	startTLSCalled    bool
	startTLSConfig    *tls.Config
	authCalled        bool
	mailFrom          string
	rcptTo            string
	quitCalled        bool
	closeCalled       bool
	dataBuffer        strings.Builder
}

func (s *smtpClientStub) Extension(ext string) (bool, string) {
	if strings.EqualFold(ext, "STARTTLS") && s.startTLSSupported {
		return true, ""
	}
	return false, ""
}

func (s *smtpClientStub) StartTLS(config *tls.Config) error {
	s.startTLSCalled = true
	s.startTLSConfig = config
	return nil
}

func (s *smtpClientStub) Auth(_ smtp.Auth) error {
	s.authCalled = true
	return nil
}

func (s *smtpClientStub) Mail(from string) error {
	s.mailFrom = from
	return nil
}

func (s *smtpClientStub) Rcpt(to string) error {
	s.rcptTo = to
	return nil
}

func (s *smtpClientStub) Data() (io.WriteCloser, error) {
	return nopWriteCloser{Writer: &s.dataBuffer}, nil
}

func (s *smtpClientStub) Close() error {
	s.closeCalled = true
	return nil
}

func (s *smtpClientStub) Quit() error {
	s.quitCalled = true
	return nil
}

type nopWriteCloser struct {
	io.Writer
}

func (n nopWriteCloser) Close() error {
	return nil
}

func TestEmailServiceSendEmailWithConfigUsesSTARTTLSOnSubmissionPort(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	client := &smtpClientStub{startTLSSupported: true}
	smtpDialFunc = func(addr string) (smtpClient, error) {
		if addr != "smtp.example.com:587" {
			t.Fatalf("unexpected SMTP addr: %s", addr)
		}
		return client, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		t.Fatalf("unexpected implicit TLS dial: %s %s", network, addr)
		return nil, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		t.Fatalf("unexpected smtp.NewClient for STARTTLS path: %s", host)
		return nil, nil
	}

	svc := &EmailService{}
	err := svc.SendEmailWithConfig(&SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "demo",
		Password: "secret",
		From:     "from@example.com",
		UseTLS:   true,
	}, "to@example.com", "subject", "<p>body</p>")
	if err != nil {
		t.Fatalf("SendEmailWithConfig returned error: %v", err)
	}

	if !client.startTLSCalled {
		t.Fatal("expected STARTTLS to be used for secure submission port")
	}
	if client.startTLSConfig == nil || client.startTLSConfig.ServerName != "smtp.example.com" {
		t.Fatalf("unexpected STARTTLS config: %+v", client.startTLSConfig)
	}
	if !client.authCalled {
		t.Fatal("expected smtp auth to be called")
	}
	if client.mailFrom != "from@example.com" {
		t.Fatalf("unexpected MAIL FROM: %s", client.mailFrom)
	}
	if client.rcptTo != "to@example.com" {
		t.Fatalf("unexpected RCPT TO: %s", client.rcptTo)
	}
	if !client.quitCalled {
		t.Fatal("expected Quit to be called after send")
	}
}

func TestEmailServiceSendEmailWithConfigRejectsPlainAuthWithoutTLS(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	smtpDialFunc = func(addr string) (smtpClient, error) {
		t.Fatalf("unexpected plain smtp.Dial call: %s", addr)
		return nil, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		t.Fatalf("unexpected implicit TLS dial: %s %s", network, addr)
		return nil, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		t.Fatalf("unexpected smtp.NewClient call: %s", host)
		return nil, nil
	}

	svc := &EmailService{}
	err := svc.SendEmailWithConfig(&SMTPConfig{
		Host:     "smtp.example.com",
		Port:     25,
		Username: "demo",
		Password: "secret",
		From:     "from@example.com",
		UseTLS:   false,
	}, "to@example.com", "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected error when plaintext SMTP authentication is requested")
	}
	if !strings.Contains(err.Error(), "requires TLS or STARTTLS") {
		t.Fatalf("expected TLS requirement error, got: %v", err)
	}
}

func TestEmailServiceSendEmailWithConfigAllowsPlainSendWithoutCredentials(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	client := &smtpClientStub{}
	smtpDialFunc = func(addr string) (smtpClient, error) {
		if addr != "smtp.example.com:25" {
			t.Fatalf("unexpected SMTP addr: %s", addr)
		}
		return client, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		t.Fatalf("unexpected implicit TLS dial: %s %s", network, addr)
		return nil, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		t.Fatalf("unexpected smtp.NewClient call: %s", host)
		return nil, nil
	}

	svc := &EmailService{}
	err := svc.SendEmailWithConfig(&SMTPConfig{
		Host:   "smtp.example.com",
		Port:   25,
		From:   "from@example.com",
		UseTLS: false,
	}, "to@example.com", "subject", "<p>body</p>")
	if err != nil {
		t.Fatalf("SendEmailWithConfig returned error: %v", err)
	}
	if client.authCalled {
		t.Fatal("did not expect smtp auth to be called without credentials")
	}
	if client.mailFrom != "from@example.com" {
		t.Fatalf("unexpected MAIL FROM: %s", client.mailFrom)
	}
	if client.rcptTo != "to@example.com" {
		t.Fatalf("unexpected RCPT TO: %s", client.rcptTo)
	}
	if !client.quitCalled {
		t.Fatal("expected Quit to be called after send")
	}
}

func TestEmailServiceTestSMTPConnectionWithConfigUsesImplicitTLSOnPort465(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	client := &smtpClientStub{}
	serverConn, peerConn := net.Pipe()
	defer func() { _ = peerConn.Close() }()

	var tlsDialCalled bool
	smtpDialFunc = func(addr string) (smtpClient, error) {
		t.Fatalf("unexpected plain smtp.Dial call: %s", addr)
		return nil, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		tlsDialCalled = true
		if network != "tcp" || addr != "smtp.example.com:465" {
			t.Fatalf("unexpected tls dial target: %s %s", network, addr)
		}
		if config == nil || config.ServerName != "smtp.example.com" {
			t.Fatalf("unexpected tls config: %+v", config)
		}
		return serverConn, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		if conn != serverConn {
			t.Fatal("smtp.NewClient should receive the TLS connection")
		}
		if host != "smtp.example.com" {
			t.Fatalf("unexpected host: %s", host)
		}
		return client, nil
	}

	svc := &EmailService{}
	err := svc.TestSMTPConnectionWithConfig(&SMTPConfig{
		Host:     "smtp.example.com",
		Port:     465,
		Username: "demo",
		Password: "secret",
		UseTLS:   true,
	})
	if err != nil {
		t.Fatalf("TestSMTPConnectionWithConfig returned error: %v", err)
	}

	if !tlsDialCalled {
		t.Fatal("expected implicit TLS dial for port 465")
	}
	if !client.authCalled {
		t.Fatal("expected smtp auth to be called")
	}
	if !client.quitCalled {
		t.Fatal("expected Quit to be called")
	}
}

func TestEmailServiceTestSMTPConnectionWithConfigRejectsSecureSMTPWithoutSTARTTLS(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	client := &smtpClientStub{startTLSSupported: false}
	smtpDialFunc = func(addr string) (smtpClient, error) {
		return client, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		t.Fatalf("unexpected implicit TLS dial: %s %s", network, addr)
		return nil, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		t.Fatalf("unexpected smtp.NewClient call: %s", host)
		return nil, nil
	}

	svc := &EmailService{}
	err := svc.TestSMTPConnectionWithConfig(&SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "demo",
		Password: "secret",
		UseTLS:   true,
	})
	if err == nil {
		t.Fatal("expected error when secure SMTP server does not advertise STARTTLS")
	}
	if !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("expected STARTTLS error, got: %v", err)
	}
	if !client.closeCalled {
		t.Fatal("expected client to be closed after STARTTLS capability check fails")
	}
}

func TestEmailServiceTestSMTPConnectionWithConfigRejectsPlainAuthWithoutTLS(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	smtpDialFunc = func(addr string) (smtpClient, error) {
		t.Fatalf("unexpected plain smtp.Dial call: %s", addr)
		return nil, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		t.Fatalf("unexpected implicit TLS dial: %s %s", network, addr)
		return nil, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		t.Fatalf("unexpected smtp.NewClient call: %s", host)
		return nil, nil
	}

	svc := &EmailService{}
	err := svc.TestSMTPConnectionWithConfig(&SMTPConfig{
		Host:     "smtp.example.com",
		Port:     25,
		Username: "demo",
		Password: "secret",
		UseTLS:   false,
	})
	if err == nil {
		t.Fatal("expected error when plaintext SMTP authentication is requested")
	}
	if !strings.Contains(err.Error(), "requires TLS or STARTTLS") {
		t.Fatalf("expected TLS requirement error, got: %v", err)
	}
}

func TestEmailServiceTestSMTPConnectionWithConfigAllowsPlainConnectionWithoutCredentials(t *testing.T) {
	originalSMTPDial := smtpDialFunc
	originalSMTPTLSDial := smtpTLSDialFunc
	originalSMTPNewClient := smtpNewClientFunc
	defer func() {
		smtpDialFunc = originalSMTPDial
		smtpTLSDialFunc = originalSMTPTLSDial
		smtpNewClientFunc = originalSMTPNewClient
	}()

	client := &smtpClientStub{}
	smtpDialFunc = func(addr string) (smtpClient, error) {
		if addr != "smtp.example.com:25" {
			t.Fatalf("unexpected SMTP addr: %s", addr)
		}
		return client, nil
	}
	smtpTLSDialFunc = func(network, addr string, config *tls.Config) (net.Conn, error) {
		t.Fatalf("unexpected implicit TLS dial: %s %s", network, addr)
		return nil, nil
	}
	smtpNewClientFunc = func(conn net.Conn, host string) (smtpClient, error) {
		t.Fatalf("unexpected smtp.NewClient call: %s", host)
		return nil, nil
	}

	svc := &EmailService{}
	err := svc.TestSMTPConnectionWithConfig(&SMTPConfig{
		Host:   "smtp.example.com",
		Port:   25,
		UseTLS: false,
	})
	if err != nil {
		t.Fatalf("TestSMTPConnectionWithConfig returned error: %v", err)
	}
	if client.authCalled {
		t.Fatal("did not expect smtp auth to be called without credentials")
	}
	if !client.quitCalled {
		t.Fatal("expected Quit to be called")
	}
}
