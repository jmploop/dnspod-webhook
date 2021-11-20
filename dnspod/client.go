package dnspod

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type Client struct {
	dnsc     *dnspod.Client
	region   string
	secretId string
	ttl      *uint64
}

type TXTRecord struct {
	Domain    string
	SubDomain string
	Value     string
}

// NewClient set retries and backoff
func NewClient(cfg Config) (*Client, error) {
	credential := common.NewCredential(cfg.SecretId, cfg.SecretKey)

	prof := profile.NewClientProfile()
	prof.RateLimitExceededMaxRetries = 3
	prof.RateLimitExceededRetryDuration = profile.ExponentialBackoff

	region := regions.Guangzhou
	client, err := dnspod.NewClient(credential, region, prof)
	if err != nil {
		return nil, err
	}

	return &Client{dnsc: client,
			region:   region,
			secretId: cfg.SecretId,
			ttl:      cfg.TTL},
		nil
}

// AddTxtRecord
func (c *Client) AddTxtRecord(tr TXTRecord) error {
	// create txt record to domain
	req := dnspod.NewCreateRecordRequest()

	req.Domain = common.StringPtr(tr.Domain)
	req.SubDomain = common.StringPtr(tr.SubDomain)
	req.Value = common.StringPtr(tr.Value)

	req.RecordType = common.StringPtr("TXT")
	req.RecordLine = common.StringPtr("默认")
	req.TTL = c.ttl

	_, err := c.dnsc.CreateRecord(req)
	return err
}

// DeleteTxtRecord
func (c *Client) DeleteTxtRecord(tr TXTRecord) error {
	// get all txt record for domain
	listReq := dnspod.NewDescribeRecordListRequest()

	listReq.Domain = common.StringPtr(tr.Domain)
	listReq.Subdomain = common.StringPtr(tr.SubDomain)

	listReq.RecordType = common.StringPtr("TXT")
	listReq.RecordLine = common.StringPtr("默认")

	listResp, err := c.dnsc.DescribeRecordList(listReq)
	if err != nil {
		return err
	}

	for _, record := range listResp.Response.RecordList {
		recordValue := *record.Value
		if recordValue == tr.Value {
			delReq := dnspod.NewDeleteRecordRequest()

			delReq.Domain = common.StringPtr(tr.Domain)
			delReq.RecordId = record.RecordId

			_, err := c.dnsc.DeleteRecord(delReq)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
