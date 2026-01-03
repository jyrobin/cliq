// Package ws provides WebSocket utilities for building client SDKs.
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Conn wraps a WebSocket connection with convenient methods.
type Conn struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	messages chan Message
	done     chan struct{}
	closed   bool
	url      string
}

// Message represents a received WebSocket message.
type Message struct {
	Type int    // websocket.TextMessage or websocket.BinaryMessage
	Data []byte
	Err  error
}

// Text returns the message data as string.
func (m *Message) Text() string {
	return string(m.Data)
}

// JSON parses the message data into v.
func (m *Message) JSON(v any) error {
	return json.Unmarshal(m.Data, v)
}

// JSONMap parses the message as a JSON map.
func (m *Message) JSONMap() (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(m.Data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MessageType extracts the "type" field from a JSON message.
func (m *Message) MessageType() string {
	var msg struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(m.Data, &msg); err != nil {
		return ""
	}
	return msg.Type
}

// Dialer configures WebSocket connection options.
type Dialer struct {
	headers http.Header
	timeout time.Duration
}

// NewDialer creates a new WebSocket dialer.
func NewDialer() *Dialer {
	return &Dialer{
		headers: make(http.Header),
		timeout: 30 * time.Second,
	}
}

// Header sets a header for the WebSocket handshake.
func (d *Dialer) Header(key, value string) *Dialer {
	d.headers.Set(key, value)
	return d
}

// Auth sets Bearer token for the WebSocket handshake.
func (d *Dialer) Auth(token string) *Dialer {
	d.headers.Set("Authorization", "Bearer "+token)
	return d
}

// Timeout sets the handshake timeout.
func (d *Dialer) Timeout(timeout time.Duration) *Dialer {
	d.timeout = timeout
	return d
}

// Dial connects to a WebSocket endpoint.
func (d *Dialer) Dial(url string) (*Conn, error) {
	return d.DialContext(context.Background(), url)
}

// DialContext connects to a WebSocket endpoint with context.
func (d *Dialer) DialContext(ctx context.Context, url string) (*Conn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: d.timeout,
	}

	conn, _, err := dialer.DialContext(ctx, url, d.headers)
	if err != nil {
		return nil, err
	}

	c := &Conn{
		conn:     conn,
		messages: make(chan Message, 100),
		done:     make(chan struct{}),
		url:      url,
	}

	go c.readLoop()
	return c, nil
}

// Dial connects to a WebSocket endpoint with default options.
func Dial(url string) (*Conn, error) {
	return NewDialer().Dial(url)
}

// DialContext connects with context and default options.
func DialContext(ctx context.Context, url string) (*Conn, error) {
	return NewDialer().DialContext(ctx, url)
}

func (c *Conn) readLoop() {
	defer close(c.messages)
	for {
		select {
		case <-c.done:
			return
		default:
			msgType, data, err := c.conn.ReadMessage()
			if err != nil {
				if !c.closed {
					c.messages <- Message{Err: err}
				}
				return
			}
			c.messages <- Message{Type: msgType, Data: data}
		}
	}
}

// Send sends a text message.
func (c *Conn) Send(message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// SendJSON sends a JSON message.
func (c *Conn) SendJSON(data any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(data)
}

// SendBinary sends a binary message.
func (c *Conn) SendBinary(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Recv waits for and returns the next message with timeout.
func (c *Conn) Recv(timeout time.Duration) *Message {
	select {
	case msg, ok := <-c.messages:
		if !ok {
			return &Message{Err: ErrConnectionClosed}
		}
		return &msg
	case <-time.After(timeout):
		return &Message{Err: fmt.Errorf("receive timeout after %v", timeout)}
	}
}

// RecvContext waits for the next message with context.
func (c *Conn) RecvContext(ctx context.Context) *Message {
	select {
	case msg, ok := <-c.messages:
		if !ok {
			return &Message{Err: ErrConnectionClosed}
		}
		return &msg
	case <-ctx.Done():
		return &Message{Err: fmt.Errorf("receive: %w", ctx.Err())}
	}
}

// RecvText waits for a text message and returns it as string.
func (c *Conn) RecvText(timeout time.Duration) (string, error) {
	msg := c.Recv(timeout)
	if msg.Err != nil {
		return "", msg.Err
	}
	return msg.Text(), nil
}

// RecvJSON waits for a message and parses it as JSON map.
func (c *Conn) RecvJSON(timeout time.Duration) (map[string]any, error) {
	msg := c.Recv(timeout)
	if msg.Err != nil {
		return nil, msg.Err
	}
	return msg.JSONMap()
}

// RecvJSONInto waits for a message and parses it into v.
func (c *Conn) RecvJSONInto(timeout time.Duration, v any) error {
	msg := c.Recv(timeout)
	if msg.Err != nil {
		return msg.Err
	}
	return msg.JSON(v)
}

// RecvUntil receives messages until condition is met or timeout.
func (c *Conn) RecvUntil(timeout time.Duration, condition func(*Message) bool) ([]*Message, error) {
	deadline := time.Now().Add(timeout)
	var messages []*Message

	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		msg := c.Recv(remaining)
		if msg.Err != nil {
			return messages, msg.Err
		}
		messages = append(messages, msg)
		if condition(msg) {
			return messages, nil
		}
	}
	return messages, fmt.Errorf("condition not met within %v", timeout)
}

// RecvUntilType receives messages until a message with the given type is received.
func (c *Conn) RecvUntilType(timeout time.Duration, msgType string) ([]*Message, error) {
	return c.RecvUntil(timeout, func(m *Message) bool {
		return m.MessageType() == msgType
	})
}

// Expect sends a message and expects a response matching a condition.
func (c *Conn) Expect(send string, timeout time.Duration, check func(string) bool) (string, error) {
	if err := c.Send(send); err != nil {
		return "", fmt.Errorf("send failed: %w", err)
	}

	msg := c.Recv(timeout)
	if msg.Err != nil {
		return "", msg.Err
	}

	text := msg.Text()
	if !check(text) {
		return text, fmt.Errorf("unexpected response: %s", text)
	}
	return text, nil
}

// ExpectJSON sends a message and expects a JSON response.
func (c *Conn) ExpectJSON(send string, timeout time.Duration) (map[string]any, error) {
	if err := c.Send(send); err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}
	return c.RecvJSON(timeout)
}

// ExpectType sends a JSON message and expects a response with the given type.
func (c *Conn) ExpectType(send any, timeout time.Duration, expectedType string) (*Message, error) {
	if err := c.SendJSON(send); err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	msg := c.Recv(timeout)
	if msg.Err != nil {
		return nil, msg.Err
	}

	if msg.MessageType() != expectedType {
		return msg, fmt.Errorf("expected type %q, got %q", expectedType, msg.MessageType())
	}
	return msg, nil
}

// Ping sends a ping and waits for pong.
func (c *Conn) Ping(timeout time.Duration) error {
	c.mu.Lock()
	err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
	c.mu.Unlock()
	if err != nil {
		return err
	}

	pongCh := make(chan struct{}, 1)
	c.conn.SetPongHandler(func(string) error {
		select {
		case pongCh <- struct{}{}:
		default:
		}
		return nil
	})

	select {
	case <-pongCh:
		return nil
	case <-time.After(timeout):
		return ErrPongTimeout
	}
}

// Close closes the WebSocket connection gracefully.
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	close(c.done)

	_ = c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	return c.conn.Close()
}

// IsClosed returns true if the connection is closed.
func (c *Conn) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// Messages returns the message channel for custom handling.
func (c *Conn) Messages() <-chan Message {
	return c.messages
}

// URL returns the connection URL.
func (c *Conn) URL() string {
	return c.url
}

// Errors
var (
	ErrConnectionClosed = fmt.Errorf("connection closed")
	ErrPongTimeout      = fmt.Errorf("pong timeout")
)
