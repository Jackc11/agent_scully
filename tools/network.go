package tools

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type ResponseData struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// /hosts
func HostsDNS(domain, ip string) string {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			if host == domain {
				addr = net.JoinHostPort(ip, port)
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://%s", domain))
	if err != nil {
		slog.Error("Failed to connect to domain", "error", err)
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	data := ResponseData{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headers,
		Body:       string(bodyBytes),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}

	jsonString := string(jsonBytes)
	return jsonString
}

// /certs
func checkCertAndKey(certFile, keyFile, expectedDomain string) (string, error) {
	var report strings.Builder

	// 1. Verify the certificate and private key match and are valid PEM files
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		report.WriteString(fmt.Sprintf("1. Verify the certificate and private key match and are valid PEM files - failed (%v)\n", err))
		return report.String(), err
	}
	report.WriteString("1. Verify the certificate and private key match and are valid PEM files - ok\n")

	// 2. Parse the leaf certificate
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		report.WriteString(fmt.Sprintf("2. Parse the leaf certificate - failed (%v)\n", err))
		return report.String(), err
	}
	report.WriteString("2. Parse the leaf certificate - ok\n")

	// 3. Check expiration dates
	now := time.Now()
	if now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		err = fmt.Errorf("valid from %s to %s", leaf.NotBefore, leaf.NotAfter)
		report.WriteString(fmt.Sprintf("3. Check expiration dates - failed (%v)\n", err))
		return report.String(), err
	}
	report.WriteString("3. Check expiration dates - ok\n")

	// 4. Verify the certificate covers the expected domain name
	if expectedDomain != "" {
		if err := leaf.VerifyHostname(expectedDomain); err != nil {
			report.WriteString(fmt.Sprintf("4. Verify hostname (%s) - failed (%v)\n", expectedDomain, err))
			return report.String(), err
		}
		report.WriteString(fmt.Sprintf("4. Verify hostname (%s) - ok\n", expectedDomain))
	}

	// 5. Verify against system root CAs and intermediate certs
	intermediates := x509.NewCertPool()
	for _, rawCert := range cert.Certificate[1:] {
		if inter, err := x509.ParseCertificate(rawCert); err == nil {
			intermediates.AddCert(inter)
		}
	}

	opts := x509.VerifyOptions{
		DNSName:       expectedDomain,
		Intermediates: intermediates,
		CurrentTime:   now,
	}
	if _, err := leaf.Verify(opts); err != nil {
		report.WriteString(fmt.Sprintf("5. Verify trust chain - failed (%v)\n", err))
		return report.String(), err
	}
	report.WriteString("5. Verify trust chain - ok\n")

	return report.String(), nil
}
