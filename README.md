# dnspod-webhook

Cert manager acme dns01 webhook provider for DNSPod using DNSPod API 3.0.

## Usage

First create [dnspod apikey](https://console.dnspod.cn/account/token/apikey).

If you want to create clusterIssuer, please run.

```bash
helm install -n cert-manager cert-manager-webhook-dnspod ./deploy/dnspod-webhook -f values.yaml
or 
helm repo add dnspod-webhook https://charts.jmploop.com/dnspod-webhook
helm repo update
helm install -n cert-manager cert-manager-webhook-dnspod dnspod-webhook/cert-manager-webhook-dnspod -f values.yaml
```

values.yaml example, additional details [see](./deploy/dnspod-webhook/values.yaml).

```yaml
clusterIssuer:
  enabled: true
  name: webhook-dnspod # ClusterIssue name
  ttl: 600 # dns ttl
  staging: false #  Let’s Encrypt staging or prod
  secretId: '' # dnspod secredId
  secretKey: '' # dnspod secretKey
  email: ''  # Let’s Encrypt email remind expire
```

Then create certificate.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example.cert
  namespace: cert-manager
spec:
  secretName: example-secret # certificate secret name
  issuerRef:
    name: webhook-dnspod # ClusterIssuer.name
    kind: ClusterIssuer
  dnsNames:
    - "*.example.com"
```
