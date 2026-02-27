package flex

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL   = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"
	userAgent = "ibkr-finance-cli/1.0"

	// 错误码
	errStillGenerating = 1019
	errRateLimit       = 1018
)

type Client struct {
	token      string
	httpClient *http.Client
	lastReq    time.Time
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendRequest 发起 Flex Query 请求，返回 ReferenceCode
func (c *Client) SendRequest(queryID string) (string, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/SendRequest?t=%s&q=%s&v=3", baseURL, c.token, queryID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result SendRequestResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Status != "Success" {
		return "", fmt.Errorf("API 错误 [%d]: %s", result.ErrorCode, result.ErrorMessage)
	}

	return result.ReferenceCode, nil
}

// GetStatement 轮询获取报表数据
func (c *Client) GetStatement(referenceCode string) (*FlexQueryResponse, []byte, error) {
	maxRetries := 20
	for i := 0; i < maxRetries; i++ {
		c.rateLimit()

		url := fmt.Sprintf("%s/GetStatement?t=%s&q=%s&v=3", baseURL, c.token, referenceCode)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("发送请求失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("读取响应失败: %w", err)
		}

		// 先尝试解析为错误响应（FlexStatementResponse）
		var errResp SendRequestResponse
		if xml.Unmarshal(body, &errResp) == nil && errResp.XMLName.Local == "FlexStatementResponse" {
			if errResp.ErrorCode == errStillGenerating {
				fmt.Printf("  报表生成中，等待重试 (%d/%d)...\n", i+1, maxRetries)
				time.Sleep(2 * time.Second)
				continue
			}
			if errResp.ErrorCode == errRateLimit {
				fmt.Println("  触发速率限制，等待 10 秒...")
				time.Sleep(10 * time.Second)
				continue
			}
			if errResp.Status != "Success" {
				return nil, nil, fmt.Errorf("API 错误 [%d]: %s", errResp.ErrorCode, errResp.ErrorMessage)
			}
		}

		// 解析为完整的 FlexQueryResponse
		var result FlexQueryResponse
		if err := xml.Unmarshal(body, &result); err != nil {
			return nil, nil, fmt.Errorf("解析报表失败: %w", err)
		}

		return &result, body, nil
	}

	return nil, nil, fmt.Errorf("获取报表超时，已重试 %d 次", maxRetries)
}

// FetchQuery 完整的获取流程：SendRequest + GetStatement
func (c *Client) FetchQuery(queryID string) (*FlexQueryResponse, []byte, error) {
	fmt.Printf("发送 Flex Query 请求 (QueryID: %s)...\n", queryID)

	refCode, err := c.SendRequest(queryID)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("获取到 ReferenceCode: %s\n", refCode)

	fmt.Println("正在获取报表数据...")
	return c.GetStatement(refCode)
}

// rateLimit 确保请求间隔至少 1 秒
func (c *Client) rateLimit() {
	elapsed := time.Since(c.lastReq)
	if elapsed < time.Second {
		time.Sleep(time.Second - elapsed)
	}
	c.lastReq = time.Now()
}
