# Task: 实现环形缓冲区

## 目标
实现固定大小的 RingBuffer[T]。

## 变更范围
- ringbuffer.go: 实现 RingBuffer

## 具体要求
- REQ-1: NewRingBuffer[T](capacity int) *RingBuffer[T]
- REQ-2: Write(v T) 写入元素（满时覆盖最旧）
- REQ-3: Read() (T, bool) 读取最旧元素
- REQ-4: Len() int 当前元素数
- REQ-5: IsFull() bool, IsEmpty() bool
- REQ-6: ToSlice() []T 按顺序返回所有元素

## 约束
- 不使用外部依赖
- 固定容量，不自动扩容

## 测试策略
- 验证写入和读取
- 验证覆盖最旧元素
- 验证 Len/IsFull/IsEmpty

## 不做什么
- 不添加线程安全
- 不添加动态扩容
