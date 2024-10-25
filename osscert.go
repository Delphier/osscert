package osscert

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

func Run() (rst string, err error) {
	var OSS_REGION = os.Getenv("OSS_REGION")
	var OSS_BUCKET = os.Getenv("OSS_BUCKET")
	var ACME_EMAIL = os.Getenv("ACME_EMAIL")

	ossConfig := oss.LoadDefaultConfig().WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).WithRegion(OSS_REGION)
	ossClient := oss.NewClient(ossConfig)
	response, err := ossClient.ListCname(context.TODO(), &oss.ListCnameRequest{Bucket: &OSS_BUCKET})
	if err != nil {
		return
	}
	cname := response.Cnames[0]
	domain := *cname.Domain
	expiry, err := time.Parse("Jan _2 15:04:05 2006 GMT", *cname.Certificate.ValidEndDate) // "Dec  5 23:59:59 2024 GMT"
	if err != nil {
		return
	}
	days := time.Until(expiry).Hours() / 24
	summary := fmt.Sprintf("Domain: %s\nExpiry: %s (%v days left)\n", domain, expiry, days)
	if days > 31 {
		return summary + "Certificate is OK!", nil
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return
	}

	user := User{
		Email: ACME_EMAIL,
		key:   privateKey,
	}

	acmeConfig := lego.NewConfig(&user)
	//acmeConfig.CADirURL = lego.LEDirectoryStaging
	acmeConfig.Certificate.KeyType = certcrypto.RSA2048

	acmeClient, err := lego.NewClient(acmeConfig)
	if err != nil {
		return
	}

	err = acmeClient.Challenge.SetHTTP01Provider(NewOSSFileChallengeProvider(ossClient, &OSS_BUCKET))
	if err != nil {
		return
	}

	reg, err := acmeClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	cert, err := acmeClient.Certificate.Obtain(request)
	if err != nil {
		return
	}

	certConfig := &oss.CertificateConfiguration{
		Certificate:       oss.Ptr(string(cert.Certificate)),
		PrivateKey:        oss.Ptr(string(cert.PrivateKey)),
		Force:             oss.Ptr(true),
		DeleteCertificate: oss.Ptr(false),
	}
	_, err = ossClient.PutCname(context.TODO(), &oss.PutCnameRequest{Bucket: &OSS_BUCKET, BucketCnameConfiguration: &oss.BucketCnameConfiguration{Domain: &domain, CertificateConfiguration: certConfig}})
	if err != nil {
		return
	}

	return summary + "Certificate Updated!", nil
}

type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type OSSFileChallengeProvider struct {
	Client *oss.Client
	Bucket *string
}

func NewOSSFileChallengeProvider(client *oss.Client, bucket *string) *OSSFileChallengeProvider {
	return &OSSFileChallengeProvider{Client: client, Bucket: bucket}

}

func buildChallengeFilePath(token string) *string {
	return oss.Ptr(strings.TrimLeft(http01.ChallengePath(token), "/"))
}
func (p *OSSFileChallengeProvider) Present(domain, token, keyAuth string) error {
	_, err := p.Client.PutObject(context.TODO(), &oss.PutObjectRequest{
		Bucket: p.Bucket,
		Key:    buildChallengeFilePath(token),
		Body:   strings.NewReader(keyAuth),
		Acl:    oss.ObjectACLPublicRead,
	})
	return err
}

func (p *OSSFileChallengeProvider) CleanUp(domain, token, keyAuth string) error {
	_, err := p.Client.DeleteObject(context.TODO(), &oss.DeleteObjectRequest{
		Bucket: p.Bucket,
		Key:    buildChallengeFilePath(token),
	})
	return err
}
