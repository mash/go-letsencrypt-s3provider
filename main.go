package main

import (
	"github.com/xenolf/lego/acme"
	"log"
	"os"
)

const (
	stagingDirectoryURL = "https://acme-staging.api.letsencrypt.org/directory"
	directoryURL        = "https://acme-v01.api.letsencrypt.org/directory"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s {email} {domain} {production|staging}", os.Args[0])
	}
	email := os.Args[1]
	domain := os.Args[2]
	directory := os.Args[3]
	if directory == "production" {
		directory = directoryURL
	} else {
		directory = stagingDirectoryURL
	}
	log.Printf("Using directory: %s", directory)

	user, err := NewUser(email)
	if err != nil {
		log.Fatalf("Failed to NewUser, error: %s", err)
	}

	// A client facilitates communication with the CA server.
	client, err := acme.NewClient(directory, &user, acme.RSA2048)
	if err != nil {
		log.Fatalf("Failed to NewClient, error: %s", err)
	}

	// New users will need to register
	reg, err := client.Register()
	if err != nil {
		log.Fatalf("Failed to Register, error: %s", err)
	}
	user.Registration = reg

	// The client has a URL to the current Let's Encrypt Subscriber
	// Agreement. The user will need to agree to it.
	if err := client.AgreeToTOS(); err != nil {
		log.Fatalf("Failed to AgreeToTOS, error: %s", err)
	}

	// We only use HTTP01
	client.ExcludeChallenges([]acme.Challenge{acme.TLSSNI01, acme.DNS01})

	provider := NewS3UploadingProvider()
	if err := client.SetChallengeProvider(acme.HTTP01, provider); err != nil {
		log.Fatalf("Failed to SetChallengeProvider failed, error: %s", err)
	}

	// The acme library takes care of completing the challenges to obtain the certificate(s).
	// The domains must resolve to this machine or you have to use the DNS challenge.
	bundle := false
	certificates, failures := client.ObtainCertificate([]string{domain}, bundle, nil)
	if len(failures) > 0 {
		log.Printf("Failed to ObtainCertificate failed, failures: %s", failures)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	log.Printf("certificates: %#v\n", certificates)
}
