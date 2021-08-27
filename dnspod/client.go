package dnspod

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type Client struct {
	dnsc *dnspod.Client
}

type TXTRecord struct {
	Domain    string
	SubDomain string
	Value     string
}

// newClient set retries and backoff
func newClient(secretId, secretKey string) (*Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	prof := profile.NewClientProfile()
	prof.RateLimitExceededMaxRetries = 3
	prof.RateLimitExceededRetryDuration = profile.ExponentialBackoff

	client, err := dnspod.NewClient(credential, regions.Guangzhou, prof)
	if err != nil {
		return nil, err
	}

	return &Client{dnsc: client}, nil
}

// addTxtRecord
func (c *Client) addTxtRecord(tr TXTRecord) error {
	// create txt record to domain
	req := dnspod.NewCreateRecordRequest()

	req.Domain = common.StringPtr(tr.Domain)
	req.SubDomain = common.StringPtr(tr.SubDomain)
	req.Value = common.StringPtr(tr.Value)

	req.RecordType = common.StringPtr("TXT")
	req.RecordLine = common.StringPtr("默认")

	_, err := c.dnsc.CreateRecord(req)
	return err
}

// deleteTxtRecord
func (c *Client) deleteTxtRecord(tr TXTRecord) error {
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

	// find equal txt record of domain then delete
	for _, record := range listResp.Response.RecordList {
		if record.Value == common.StringPtr(tr.Value) {
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
