package at

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tarm/serial"
)

// Modem 配置
type Config struct {
	PortName        string           // 串口名称，如 '/dev/ttyUSB0' 或 'COM3'
	BaudRate        int              // 波特率，如 115200
	DataBits        int              // 数据位，如 8
	StopBits        int              // 停止位，如 1
	Parity          string           // 校验位，如 'N'
	ReadTimeout     time.Duration    // 读取超时时间
	WriteTimeout    time.Duration    // 写入超时时间
	CommandSet      *CommandSet      // 自定义 AT 命令集，如果为 nil 则使用默认命令集
	NotificationSet *NotificationSet // 自定义通知类型集，如果为 nil 则使用默认通知集
	ResponseSet     *ResponseSet     // 自定义响应类型集，如果为 nil 则使用默认响应集
}

// Modem 连接实现
type Connection struct {
	port          *serial.Port
	config        Config
	commands      CommandSet      // 使用的 AT 命令集
	notifications NotificationSet // 使用的通知类型集
	responses     ResponseSet     // 使用的响应类型集
	isClosed      atomic.Bool     // 连接是否已关闭（原子操作保证并发安全）

	// 统一读取相关字段
	reader       *bufio.Reader       // 统一的读取器
	responseChan chan string         // 命令响应通道
	urcHandler   NotificationHandler // 通知处理函数
	urcMu        sync.RWMutex        // 保护通知函数的读写锁
	mu           sync.Mutex          // 保护命令发送的互斥锁
}

// newConnection 创建一个新的 AT 连接
func newConnection(config Config) (AT, error) {
	port, err := serial.OpenPort(&serial.Config{
		Name:        config.PortName,
		Baud:        config.BaudRate,
		ReadTimeout: config.ReadTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	// 如果没有指定命令集，使用默认命令集
	commands := DefaultCommandSet()
	if config.CommandSet != nil {
		commands = *config.CommandSet
	}

	// 如果没有指定通知集，使用默认通知集
	notifications := DefaultNotificationSet()
	if config.NotificationSet != nil {
		notifications = *config.NotificationSet
	}

	// 如果没有指定响应集，使用默认响应集
	responses := DefaultResponseSet()
	if config.ResponseSet != nil {
		responses = *config.ResponseSet
	}

	conn := &Connection{
		port:          port,
		config:        config,
		commands:      commands,
		notifications: notifications,
		responses:     responses,
		reader:        bufio.NewReader(port),
		responseChan:  make(chan string, 100), // 带缓冲的通道
	}

	// 启动统一读取循环
	go conn.readLoop()

	return conn, nil
}

// IsConnected 检查连接状态
func (m *Connection) IsConnected() bool {
	return !m.isClosed.Load()
}

// Close 关闭连接
func (m *Connection) Close() error {
	if m.isClosed.Swap(true) {
		return nil // 已经关闭过了
	}

	// 关闭响应通道
	close(m.responseChan)

	return m.port.Close()
}

// readLoop 从串口读取数据并分发
func (m *Connection) readLoop() {
	for !m.isClosed.Load() {
		line, err := m.reader.ReadString('\n')
		if err != nil {
			// 连接已关闭，退出循环
			if m.isClosed.Load() {
				return
			}
			// EOF或超时错误，继续监听
			if err == io.EOF {
				continue
			}
			if strings.Contains(err.Error(), "timeout") {
				continue
			}
			// 严重错误，关闭连接
			_ = m.Close()
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue // 忽略空行
		}

		// 分发数据
		if m.notifications.IsNotification(line) {
			m.urcMu.RLock()
			handler := m.urcHandler
			m.urcMu.RUnlock()
			if handler != nil {
				go handler(line)
			}
		} else {
			select {
			case m.responseChan <- line:
			default:
				// 通道满了，丢弃数据（避免阻塞）
			}
		}
	}
}

// writeString 写入数据到串口
func (m *Connection) writeString(data string) error {
	if m.isClosed.Load() {
		return ErrConnectionClosed
	}

	// 防止并发写
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空响应通道中的旧数据
	for len(m.responseChan) > 0 {
		<-m.responseChan
	}

	n, err := m.port.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("failed to write to port: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("incomplete write: wrote %d of %d bytes", n, len(data))
	}

	return nil
}

// SendCommand 发送 AT 命令并等待响应
func (m *Connection) SendCommand(ctx context.Context, command string) ([]string, error) {
	if m.isClosed.Load() {
		return nil, ErrConnectionClosed
	}

	// 添加回车换行符并写入命令
	if err := m.writeString(command + "\r\n"); err != nil {
		return nil, err
	}

	return m.readResponse(ctx)
}

// SendCommandExpect 发送 AT 命令并期望特定响应
func (m *Connection) SendCommandExpect(ctx context.Context, command string, expected string) error {
	responses, err := m.SendCommand(ctx, command)
	if err != nil {
		return err
	}

	// 检查是否包含期望的响应
	for _, response := range responses {
		if strings.Contains(response, expected) {
			return nil
		}
	}

	return fmt.Errorf("expected response %q not found in %v", expected, responses)
}

// ListenNotifications 注册 modem 通知处理器
func (m *Connection) ListenNotifications(ctx context.Context, handler NotificationHandler) error {
	if m.isClosed.Load() {
		return ErrConnectionClosed
	}

	// 注册通知处理器
	m.urcMu.Lock()
	m.urcHandler = handler
	m.urcMu.Unlock()

	// 监听上下文取消，取消时移除处理器
	go func() {
		<-ctx.Done()
		m.urcMu.Lock()
		m.urcHandler = nil
		m.urcMu.Unlock()
	}()

	return nil
}
