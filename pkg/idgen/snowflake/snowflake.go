package snowflake

import (
	"errors"
	"sync"
	"time"

	"github.com/luojh/wallet/pkg/idgen"
)

const (
	// 雪花ID位数配置
	timestampBits = 42 // 时间戳位数
	machineIDBits = 6  // 机器码位数
	sequenceBits  = 12 // 序列号位数

	// 最大值
	maxMachineID = -1 ^ (-1 << machineIDBits) // 63
	maxSequence  = -1 ^ (-1 << sequenceBits)  // 4095

	// 位移
	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits
	sequenceMask   = maxSequence
	timestampMask  = -1 ^ (-1 << timestampBits)
)

var (
	ErrInvalidMachineID = errors.New("machine ID must be between 0 and 63")
	ErrSystemClock      = errors.New("system clock moved backwards")
)

// SnowflakeGenerator 雪花算法ID生成器
// ID结构: 1位符号位 + 42位时间戳 + 6位机器码 + 12位序列号
type SnowflakeGenerator struct {
	mu          sync.Mutex
	machineID   int64
	sequence    int64
	lastTime    int64
	epoch       time.Time
	maxSequence int64
}

// NewSnowflakeGenerator 创建雪花算法ID生成器
// machineID: 机器码 (0-63)
func NewSnowflakeGenerator(machineID int64) (idgen.IDGenerator, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, ErrInvalidMachineID
	}

	// 使用 2024-01-01 00:00:00 UTC 作为起始时间
	epoch := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	return &SnowflakeGenerator{
		machineID:   machineID,
		sequence:    0,
		lastTime:    0,
		epoch:       epoch,
		maxSequence: maxSequence,
	}, nil
}

func (g *SnowflakeGenerator) Clone() idgen.IDGenerator {
	return &SnowflakeGenerator{
		machineID:   g.machineID,
		sequence:    g.sequence,
		lastTime:    g.lastTime,
		epoch:       g.epoch,
		maxSequence: g.maxSequence,
	}
}

// Generate 生成唯一的int64 ID
func (g *SnowflakeGenerator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli() - g.epoch.UnixMilli()

	if now < g.lastTime {
		return 0, ErrSystemClock
	}

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & sequenceMask
		if g.sequence == 0 {
			// 当前毫秒的序列号用完，等待下一毫秒
			for now <= g.lastTime {
				now = time.Now().UnixMilli() - g.epoch.UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	// 组装ID: 时间戳(42位) + 机器码(6位) + 序列号(12位)
	id := ((now & timestampMask) << timestampShift) |
		(g.machineID << machineIDShift) |
		g.sequence

	return id, nil
}

// GenerateString 生成字符串格式的ID
func (g *SnowflakeGenerator) GenerateString() (string, error) {
	id, err := g.Generate()
	if err != nil {
		return "", err
	}
	return Int64ToString(id), nil
}

// Parse 解析ID
func (g *SnowflakeGenerator) Parse(id int64) idgen.ParsedID {
	sequence := id & sequenceMask
	machineID := (id >> machineIDShift) & maxMachineID
	timestamp := (id >> timestampShift) & timestampMask

	return idgen.ParsedID{
		Timestamp: g.epoch.Add(time.Duration(timestamp) * time.Millisecond),
		MachineID: machineID,
		Sequence:  sequence,
	}
}

// Int64ToString 将int64转换为字符串
func Int64ToString(id int64) string {
	const digits = "0123456789"
	var buf [20]byte
	i := len(buf)
	neg := id < 0
	if neg {
		id = -id
	}
	for id >= 10 {
		i--
		buf[i] = digits[id%10]
		id /= 10
	}
	i--
	buf[i] = digits[id]
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
