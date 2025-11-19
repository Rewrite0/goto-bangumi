// Package eventbus provides a simple event bus for decoupling communication between modules.
package eventbus

import (
	"fmt"
	"sort"
)

// handlerWrapper 包装事件处理器，添加优先级和唯一标识
type handlerWrapper struct {
	id       int          // 处理器的唯一标识
	handler  EventHandler // 事件处理函数
	priority int          // 优先级（数字越小优先级越高）
}

// EventBus 事件总线结构
type EventBus struct {
	handlers map[string][]handlerWrapper // 事件类型 -> 处理器列表的映射
	nextID   int                         // 下一个处理器 ID
}

// New 创建一个新的 EventBus 实例
func New() *EventBus {
	return &EventBus{
		handlers: make(map[string][]handlerWrapper),
		nextID:   1,
	}
}

// Subscribe 订阅事件
// eventType: 事件类型
// handler: 事件处理函数
// priority: 优先级（数字越小优先级越高，建议使用 0-100 范围）
// 返回: 处理器ID（用于取消订阅）和可能的错误
func (eb *EventBus) Subscribe(eventType string, handler EventHandler, priority int) (int, error) {
	if eventType == "" {
		return 0, fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return 0, fmt.Errorf("handler cannot be nil")
	}

	// 生成唯一的处理器 ID
	handlerID := eb.nextID
	eb.nextID++

	// 创建处理器包装
	wrapper := handlerWrapper{
		id:       handlerID,
		handler:  handler,
		priority: priority,
	}

	// 添加到处理器列表
	eb.handlers[eventType] = append(eb.handlers[eventType], wrapper)

	// 按优先级排序（优先级小的排在前面）
	eb.sortHandlers(eventType)

	return handlerID, nil
}

// Unsubscribe 取消订阅
// eventType: 事件类型
// handlerID: 处理器ID（Subscribe 返回的）
func (eb *EventBus) Unsubscribe(eventType string, handlerID int) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handlerID <= 0 {
		return fmt.Errorf("handler ID must be positive")
	}

	handlers, exists := eb.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers registered for event type: %s", eventType)
	}

	// 查找并移除指定的处理器
	for i, wrapper := range handlers {
		if wrapper.id == handlerID {
			// 从切片中移除该处理器
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)

			// 如果该事件类型没有处理器了，删除该键
			if len(eb.handlers[eventType]) == 0 {
				delete(eb.handlers, eventType)
			}

			return nil
		}
	}

	return fmt.Errorf("handler ID not found: %d", handlerID)
}

// Publish 发布事件（同步调用所有处理器）
// eventType: 事件类型
// data: 事件数据
// 返回: 第一个处理器返回的错误（如果有）
func (eb *EventBus) Publish(eventType string, data any) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	handlers, exists := eb.handlers[eventType]
	if !exists || len(handlers) == 0 {
		// 没有订阅者，静默返回
		return nil
	}

	// 按优先级顺序同步调用所有处理器
	for _, wrapper := range handlers {
		if err := wrapper.handler(data); err != nil {
			// 返回第一个错误，但会继续执行后续处理器
			// 如果需要在错误时停止执行，可以改为 return err
			return err
		}
	}

	return nil
}

// PublishAsync 异步发布事件（不等待处理器执行完成）
// 注意：由于项目不需要线程安全，这个方法仅作为演示，实际使用需谨慎
// eventType: 事件类型
// data: 事件数据
func (eb *EventBus) PublishAsync(eventType string, data interface{}) {
	if eventType == "" {
		return
	}

	handlers, exists := eb.handlers[eventType]
	if !exists || len(handlers) == 0 {
		return
	}

	// 异步调用所有处理器
	go func() {
		for _, wrapper := range handlers {
			wrapper.handler(data)
		}
	}()
}


// Clear 清空所有订阅
func (eb *EventBus) Clear() {
	eb.handlers = make(map[string][]handlerWrapper)
}

// sortHandlers 按优先级对处理器进行排序
func (eb *EventBus) sortHandlers(eventType string) {
	handlers := eb.handlers[eventType]
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].priority < handlers[j].priority
	})
	eb.handlers[eventType] = handlers
}
