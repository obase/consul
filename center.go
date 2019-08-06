package center

import (
	"errors"
	"time"
	"unsafe"
)

var ErrInvalidClient = errors.New("invalid consul client")

type Config struct {
	Address string              // 远程地址
	Timeout time.Duration       // 本地缓存过期时间
	Configs map[string][]string // 本地配置address
}

// 根据consul的服务项设计
type Check struct {
	Type     string `json:"type,omitempty"`
	Target   string `json:"target,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Interval string `json:"interval,omitempty"`
}

type Service struct {
	Id   string `json:"id,omitempty"` // 如果为空则自动生成
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

type Center interface {
	Register(service *Service, check *Check) (err error)
	Deregister(serviceId string) (err error)
	Discovery(name string) ([]*Service, error)
	Robin(name string) (*Service, error)
	Hash(name string, key string) (*Service, error)
}

var Default Center

func Setup(opt *Config) {
	if len(opt.Configs) > 0 {
		Default = newConfigClient(opt.Configs)
	} else {
		Default = newConsulCenter(opt)
	}
}

func Register(service *Service, check *Check) (err error) {
	if Default == nil {
		return ErrInvalidClient
	}
	return Default.Register(service, check)
}
func Deregister(serviceId string) (err error) {
	if Default != nil {
		return ErrInvalidClient
	}
	return Default.Deregister(serviceId)
}
func Discovery(name string) ([]*Service, error) {
	if Default == nil {
		return nil, ErrInvalidClient
	}
	return Default.Discovery(name)
}

func Robin(name string) (*Service, error) {
	if Default == nil {
		return nil, ErrInvalidClient
	}
	return Default.Robin(name)
}
func Hash(name string, key string) (*Service, error) {
	if Default == nil {
		return nil, ErrInvalidClient
	}
	return Default.Hash(name, key)
}

const (
	c1_32 uint32 = 0xcc9e2d51
	c2_32 uint32 = 0x1b873593
)

// GetHash returns a murmur32 hash for the data slice.
func mmhash(data []byte) uint32 {
	// Seed is set to 37, same as C# version of emitter
	var h1 uint32 = 37

	nblocks := len(data) / 4
	var p uintptr
	if len(data) > 0 {
		p = uintptr(unsafe.Pointer(&data[0]))
	}

	p1 := p + uintptr(4*nblocks)
	for ; p < p1; p += 4 {
		k1 := *(*uint32)(unsafe.Pointer(p))

		k1 *= c1_32
		k1 = (k1 << 15) | (k1 >> 17) // rotl32(k1, 15)
		k1 *= c2_32

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19) // rotl32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	tail := data[nblocks*4:]

	var k1 uint32
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1_32
		k1 = (k1 << 15) | (k1 >> 17) // rotl32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	h1 ^= uint32(len(data))

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return (h1 << 24) | (((h1 >> 8) << 16) & 0xFF0000) | (((h1 >> 16) << 8) & 0xFF00) | (h1 >> 24)
}
