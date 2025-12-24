package sh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSConn wraps a WebSocket connection with convenient methods.
type WSConn struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	messages chan WSMessage
	errors   chan error
	done     chan struct{}
	closed   bool
}

// WSMessage represents a received WebSocket message.
type WSMessage struct {
	Type int    // websocket.TextMessage or websocket.BinaryMessage
	Data []byte
	Err  error
}

// Text returns the message as string.
func (m *WSMessage) Text() string {
	return string(m.Data)
}

// JSON parses the message as JSON map.
func (m *WSMessage) JSON() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(m.Data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WSDialer configures WebSocket connection options.
type WSDialer struct {
	headers http.Header
	timeout time.Duration
}

// WS creates a new WebSocket dialer.
func WS() *WSDialer {
	return &WSDialer{
		headers: make(http.Header),
		timeout: 30 * time.Second,
	}
}

// Header sets a header for the WebSocket handshake.
func (d *WSDialer) Header(key, value string) *WSDialer {
	d.headers.Set(key, value)
	return d
}

// Auth sets Bearer token for the WebSocket handshake.
func (d *WSDialer) Auth(token string) *WSDialer {
	d.headers.Set("Authorization", "Bearer "+token)
	return d
}

// Timeout sets the handshake timeout.
func (d *WSDialer) Timeout(timeout time.Duration) *WSDialer {
	d.timeout = timeout
	return d
}

// Dial connects to a WebSocket endpoint.
func (d *WSDialer) Dial(url string) (*WSConn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: d.timeout,
	}

	conn, _, err := dialer.Dial(url, d.headers)
	if err != nil {
		return nil, err
	}

	ws := &WSConn{
		conn:     conn,
		messages: make(chan WSMessage, 100),
		errors:   make(chan error, 10),
		done:     make(chan struct{}),
	}

	// Start reader goroutine
	go ws.readLoop()

	return ws, nil
}

// Dial is a convenience function to dial a WebSocket.
func Dial(url string) (*WSConn, error) {
	return WS().Dial(url)
}

func (ws *WSConn) readLoop() {
	defer close(ws.messages)
	for {
		select {
		case <-ws.done:
			return
		default:
			msgType, data, err := ws.conn.ReadMessage()
			if err != nil {
				if !ws.closed {
					ws.messages <- WSMessage{Err: err}
				}
				return
			}
			ws.messages <- WSMessage{Type: msgType, Data: data}
		}
	}
}

// Send sends a text message.
func (ws *WSConn) Send(message string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// SendJSON sends a JSON message.
func (ws *WSConn) SendJSON(data interface{}) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.conn.WriteJSON(data)
}

// SendBinary sends a binary message.
func (ws *WSConn) SendBinary(data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Recv waits for and returns the next message.
func (ws *WSConn) Recv(timeout time.Duration) *WSMessage {
	select {
	case msg, ok := <-ws.messages:
		if !ok {
			return &WSMessage{Err: fmt.Errorf("connection closed")}
		}
		return &msg
	case <-time.After(timeout):
		return &WSMessage{Err: fmt.Errorf("receive timeout after %v", timeout)}
	}
}

// RecvText waits for a text message and returns it as string.
func (ws *WSConn) RecvText(timeout time.Duration) (string, error) {
	msg := ws.Recv(timeout)
	if msg.Err != nil {
		return "", msg.Err
	}
	return msg.Text(), nil
}

// RecvJSON waits for a message and parses it as JSON.
func (ws *WSConn) RecvJSON(timeout time.Duration) (map[string]interface{}, error) {
	msg := ws.Recv(timeout)
	if msg.Err != nil {
		return nil, msg.Err
	}
	return msg.JSON()
}

// RecvUntil receives messages until a condition is met or timeout.
func (ws *WSConn) RecvUntil(timeout time.Duration, condition func(*WSMessage) bool) ([]*WSMessage, error) {
	deadline := time.Now().Add(timeout)
	var messages []*WSMessage

	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		msg := ws.Recv(remaining)
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

// Expect sends a message and expects a response matching a condition.
func (ws *WSConn) Expect(send string, timeout time.Duration, check func(string) bool) (string, error) {
	if err := ws.Send(send); err != nil {
		return "", fmt.Errorf("send failed: %w", err)
	}

	msg := ws.Recv(timeout)
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
func (ws *WSConn) ExpectJSON(send string, timeout time.Duration) (map[string]interface{}, error) {
	if err := ws.Send(send); err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}
	return ws.RecvJSON(timeout)
}

// Ping sends a ping and waits for pong.
func (ws *WSConn) Ping(timeout time.Duration) error {
	ws.mu.Lock()
	err := ws.conn.WriteMessage(websocket.PingMessage, []byte{})
	ws.mu.Unlock()
	if err != nil {
		return err
	}

	// Set pong handler with channel
	pongCh := make(chan struct{}, 1)
	ws.conn.SetPongHandler(func(string) error {
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
		return fmt.Errorf("pong timeout")
	}
}

// Close closes the WebSocket connection.
func (ws *WSConn) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}
	ws.closed = true
	close(ws.done)

	// Send close message
	ws.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	return ws.conn.Close()
}

// Messages returns the message channel for custom handling.
func (ws *WSConn) Messages() <-chan WSMessage {
	return ws.messages
}
