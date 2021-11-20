package dnspod

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var ttl = uint64(600)

var config = Config{
	SecretId:  "",
	SecretKey: "",
	TTL:       &ttl,
}

var txtRecord = TXTRecord{
	Domain:    "webhook-dnspod.cert-manager.com",
	SubDomain: "_acme-challenge",
	Value:     "xxx",
}

func TestNewClient(t *testing.T) {
	_, err := NewClient(config)
	assert.Nil(t, err)
}

func TestAddTxtRecord(t *testing.T) {
	client, err := NewClient(config)
	err = client.AddTxtRecord(txtRecord)
	assert.Nil(t, err)
}

func TestDeleteTxtRecord(t *testing.T) {
	client, err := NewClient(config)
	err = client.DeleteTxtRecord(txtRecord)
	assert.Nil(t, err)
}
