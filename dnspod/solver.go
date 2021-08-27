package dnspod

import (
	"context"
	"fmt"
	"github.com/jetstack/cert-manager/pkg/acme/webhook"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// DnsPodSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
type DnsPodSolver struct {
	client *kubernetes.Clientset
}

func NewSolver() webhook.Solver {
	return &DnsPodSolver{}
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
func (c *DnsPodSolver) Name() string {
	return "dnspod"
}

func (c *DnsPodSolver) newDnsPodClient(ch *v1alpha1.ChallengeRequest) (*Client, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return nil, err
	}

	klog.Infof("Decoded config: %v", cfg)

	ctx := context.Background()
	secretId, err := c.getSecret(ctx, cfg.APISecretIdRef, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}
	secretKey, err := c.getSecret(ctx, cfg.APISecretKeyRef, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}

	client, err := newClient(string(secretId), string(secretKey))
	if err != nil {
		return nil, fmt.Errorf("new dns client error: %v", err)
	}

	return client, nil
}

func (c *DnsPodSolver) getSecret(ctx context.Context, selector cmmeta.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := c.client.CoreV1().Secrets(ns).Get(
		ctx,
		selector.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %q", ns+"/"+selector.Name)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}

	return nil, errors.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *DnsPodSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Presenting txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)

	client, err := c.newDnsPodClient(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}

	txtRecord := TXTRecord{
		Domain:    "",
		SubDomain: "",
		Value:     ch.Key,
	}

	err = client.addTxtRecord(txtRecord)
	if err != nil {
		klog.Errorf("Add txt record %q error: %v", ch.ResolvedFQDN, err)
		return err
	}

	klog.Infof("Presented txt record %v", ch.ResolvedFQDN)
	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *DnsPodSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Cleaning up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)

	client, err := c.newDnsPodClient(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}

	txtRecord := TXTRecord{
		Domain:    "",
		SubDomain: "",
		Value:     ch.Key,
	}

	err = client.deleteTxtRecord(txtRecord)
	if err != nil {
		klog.Errorf("Delete domain record %v error: %v", ch.ResolvedFQDN, err)
		return err
	}

	klog.Infof("Cleaned up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *DnsPodSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.client = cl
	return nil
}
