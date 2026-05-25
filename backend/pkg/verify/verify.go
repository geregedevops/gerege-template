// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package verify нь GeregeCloud Verify (https://verify.gecloud.mn/) API-руу
// OTP илгээх / шалгах хоёр алхамыг нимгэн боож HTTP клиент болгоно. Энэ API
// нь кодыг өөрөө үүсгэж, бэлэн hash-аар хадгалж, brute-force хамгаалалттай
// баталгаажуулдаг тул template нь дотооддоо OTP үүсгэж SMTP-ээр явуулах
// үүргээ түүнд шилжүүлэхэд боломжтой. Үр дүнд нь Redis-д зөвхөн `request_id`
// л хадгалагдана; OTP-ийн bcrypt comparator / attempts counter нь алсаас байна.
package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL нь GeregeCloud Verify API-н production endpoint юм. Operator
// config-ээс өөр URL-ээр overrride хийж болно (жишээ нь staging-д).
const DefaultBaseURL = "https://api.gecloud.mn/v1/verify"

// ChannelEmail / ChannelSMS — verify.gecloud.mn-ийн дэмждэг суваг сонголтууд.
// channel нь Send хүсэлт бүрт орох тул operator config-оор сонгуулж байна.
const (
	ChannelEmail = "email"
	ChannelSMS   = "sms"
)

// ErrNotApproved нь /check хариунд status != "approved" үед буцаагдана —
// дуудагч (auth usecase) үүнийг "буруу код" гэж тайлбарлана.
var ErrNotApproved = errors.New("verify: otp not approved")

// Sender нь auth usecase-ийн харьцах хил (boundary): Send нь алсаас илгээж
// request_id буцаана, Check нь хэрэглэгчийн оруулсан кодыг тэр request_id-тай
// тулгана. Интерфейс байгаа нь тестүүдэд бодит HTTP алхалгүй mockery-р
// мокчлох боломжийг олгоно.
type Sender interface {
	// Send нь destination (email эсвэл утасны дугаар) руу OTP илгээнэ.
	// Channel нь клиент үүсгэх үед тогтоогддог. Амжилттай үед сервер талаас
	// олгосон request_id-г буцаана — дуудагч үүнийг Check-д ашиглахын тулд
	// хадгалах ёстой.
	Send(ctx context.Context, destination string) (requestID string, err error)
	// Check нь хэрэглэгчийн оруулсан кодыг request_id-тай тулгаж, амжилттай
	// бол nil буцаана. Код буруу эсвэл хугацаа дуусаагүй бол ErrNotApproved
	// эсвэл сүлжээ/сервер алдаа буцна.
	Check(ctx context.Context, requestID, code string) error
}

// Client нь HTTP-аар verify.gecloud.mn-руу хүсэлт хийнэ. Аюулгүй өгөгдмөл
// timeout (10с) суурьтай тул аль нэг хүсэлт worker goroutine-ийг хязгааргүй
// блоклохгүй.
type Client struct {
	baseURL string
	apiKey  string
	channel string
	http    *http.Client
}

// NewClient нь өгсөн baseURL, apiKey, channel-тай шинэ Client үүсгэнэ. baseURL
// нь sufix "/" ашиглавал хасагдана. apiKey хоосон үед Client нь call бүр дээр
// "missing api key" алдаа буцаана — операторт чимээгүй буруу тохиргоо хийхээс
// сэргийлнэ.
func NewClient(baseURL, apiKey, channel string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if channel == "" {
		channel = ChannelEmail
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		channel: channel,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type sendReq struct {
	To      string `json:"to"`
	Channel string `json:"channel"`
}

type sendResp struct {
	RequestID string `json:"request_id"`
}

type checkReq struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
}

type checkResp struct {
	Status string `json:"status"`
}

// errResp нь сервер JSON алдаа буцаасан тохиолдолд боломжит мессежийг ил болгоно.
// Үл мэдэгдэх формат алдаагүй задлан унших боломжтой тул log/wrap-ийн чанарыг
// сайжруулна.
type errResp struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Send нь /send endpoint руу хүсэлт явуулна.
func (c *Client) Send(ctx context.Context, destination string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("verify: api key is not configured")
	}
	var out sendResp
	if err := c.do(ctx, "/send", sendReq{To: destination, Channel: c.channel}, &out); err != nil {
		return "", err
	}
	if out.RequestID == "" {
		return "", errors.New("verify: server returned empty request_id")
	}
	return out.RequestID, nil
}

// Check нь /check endpoint руу хүсэлт явуулж, "approved" төлвийг шалгана.
func (c *Client) Check(ctx context.Context, requestID, code string) error {
	if c.apiKey == "" {
		return errors.New("verify: api key is not configured")
	}
	var out checkResp
	if err := c.do(ctx, "/check", checkReq{RequestID: requestID, Code: code}, &out); err != nil {
		return err
	}
	if !strings.EqualFold(out.Status, "approved") {
		return ErrNotApproved
	}
	return nil
}

// do нь хүсэлтийг JSON-р marshal хийж, X-API-Key толгойтойгоор POST хийгээд
// хариуг dst руу декодлоно. HTTP 2xx бус хариуг алдаа болгож, серверийн
// мессежийг (байгаа бол) ил болгоно.
func (c *Client) do(ctx context.Context, path string, body any, dst any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("verify: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("verify: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("verify: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if readErr != nil {
		return fmt.Errorf("verify: read response: %w", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Серверийн өгсөн мессежийг (байгаа бол) гарга — дотоод stack trace биш.
		var er errResp
		_ = json.Unmarshal(raw, &er)
		msg := strings.TrimSpace(er.Message)
		if msg == "" {
			msg = strings.TrimSpace(er.Error)
		}
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		return fmt.Errorf("verify: %s returned %d: %s", path, resp.StatusCode, msg)
	}

	if dst == nil {
		return nil
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("verify: decode response: %w", err)
	}
	return nil
}
