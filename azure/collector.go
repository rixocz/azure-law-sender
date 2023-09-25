package azure

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"golang.org/x/text/encoding/unicode"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var workspaceRegex = regexp.MustCompile("(?i)^/Subscriptions/[^/]+/resourceGroups/(?P<rg_name>[^/]+)/providers/[^/]+/.+/(?P<name>.+)$")

type Collector interface {
	SendData(reader io.Reader) error
}

type azureCollector struct {
	workspaceId string
	sharedKey   string
	tableName   string
	timestamp   *time.Time
}

func (c azureCollector) SendData(reader io.Reader) error {
	err := sendDataToCollector(c.workspaceId, c.sharedKey, c.tableName, c.timestamp, reader)
	if err != nil {
		return err
	}
	return nil
}

func NewCollector(subscriptionId string, workspaceId string, tableName string) (Collector, error) {
	sharedKey, err := getSharedKey(subscriptionId, workspaceId)
	if err != nil {
		return nil, err
	}
	return azureCollector{
		workspaceId: workspaceId,
		sharedKey:   sharedKey,
		tableName:   tableName,
	}, nil
}

func getSharedKey(subscriptionId string, workspaceId string) (string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", err
	}
	workspaceName, workspaceRgName, err := getWorkspace(subscriptionId, workspaceId, cred)
	if err != nil {
		return "", err
	}
	client, err := armoperationalinsights.NewSharedKeysClient(subscriptionId, cred, nil)
	if err != nil {
		return "", err
	}
	keys, err := client.GetSharedKeys(context.Background(), workspaceRgName, workspaceName, nil)
	if err != nil {
		return "", err
	}
	return *keys.PrimarySharedKey, nil
}

func getWorkspace(subscriptionId string, workspaceId string, cred azcore.TokenCredential) (string, string, error) {
	workspacesClient, err := armoperationalinsights.NewWorkspacesClient(subscriptionId, cred, nil)
	if err != nil {
		return "", "", err
	}
	pager := workspacesClient.NewListPager(nil)
	if pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return "", "", err
		}
		for _, workspace := range page.Value {
			if *workspace.Properties.CustomerID == workspaceId {
				matches := workspaceRegex.FindStringSubmatch(*workspace.ID)
				workspaceName := matches[workspaceRegex.SubexpIndex("name")]
				workspaceRgName := matches[workspaceRegex.SubexpIndex("rg_name")]
				if workspaceName == "" {
					return "", "", fmt.Errorf("workspace name not found in workspace ID %q", *workspace.ID)
				}
				if workspaceRgName == "" {
					return "", "", fmt.Errorf("workspace resource group name not found in workspace ID %q", *workspace.ID)
				}
				return workspaceName, workspaceRgName, nil
			}
		}
	}

	return "", "", fmt.Errorf("workspace with customer ID %q not found", workspaceId)
}

type errResponse struct {
	Err string `json:"Error"`
	Msg string `json:"Message"`
}

func sendDataToCollector(workspaceId string, sharedKey string, tableName string, timestamp *time.Time, body io.Reader) error {
	url := fmt.Sprintf("https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01", workspaceId)

	if timestamp == nil {
		now := time.Now()
		timestamp = &now
	}
	gmtTimestamp := timestamp.In(time.FixedZone("GMT", 0))

	client := http.DefaultClient
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	if req.ContentLength == 0 { // req need to have content-length
		var buf bytes.Buffer
		_, err := io.Copy(&buf, body)
		if err != nil {
			return err
		}
		req, err = http.NewRequest("POST", url, &buf)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ms-Date", gmtTimestamp.Format(time.RFC1123))
	req.Header.Set("Log-Type", tableName)
	req.Header.Set("Time-Generated-Field", gmtTimestamp.Format(time.RFC3339))
	signature, err := createReqSignature(req, sharedKey)
	req.Header.Set("Authorization", fmt.Sprintf("SharedKey %s:%s", workspaceId, signature))
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode > 399 {
		resBody, _ := io.ReadAll(res.Body)
		var errRes errResponse
		err = json.Unmarshal(resBody, &errRes)
		return fmt.Errorf("response with status %s error: %q => %q", res.Status, errRes.Err, errRes.Msg)
	}
	return nil
}

func createReqSignature(req *http.Request, b64Key string) (string, error) {
	raw := strings.Join([]string{
		req.Method,
		strconv.FormatInt(req.ContentLength, 10),
		req.Header.Get("Content-Type"),
		"x-ms-date:" + req.Header.Get("X-Ms-Date"),
		"/api/logs",
	}, "\n")
	enc, err := hashHmac256b64enc(raw, b64Key)
	if err != nil {
		return "", err
	}
	return enc, nil
}

func hashHmac256b64enc(data string, b64Key string) (string, error) {
	uData, _ := unicode.UTF8.NewDecoder().String(data)
	uKey, _ := unicode.UTF8.NewDecoder().String(b64Key)
	rawKey, err := base64.StdEncoding.DecodeString(uKey)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, rawKey)
	_, err = mac.Write([]byte(uData))
	if err != nil {
		return "", err
	}
	hash := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash), nil
}
