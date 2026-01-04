package util

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
)

// Encoder provides binary encoding utilities
type Encoder struct {
	buf *bytes.Buffer
}

// NewEncoder creates a new encoder
func NewEncoder() *Encoder {
	return &Encoder{buf: new(bytes.Buffer)}
}

// WriteUint8 writes a uint8
func (e *Encoder) WriteUint8(v uint8) error {
	return e.buf.WriteByte(v)
}

// WriteUint16 writes a uint16 in big-endian
func (e *Encoder) WriteUint16(v uint16) error {
	return binary.Write(e.buf, binary.BigEndian, v)
}

// WriteUint32 writes a uint32 in big-endian
func (e *Encoder) WriteUint32(v uint32) error {
	return binary.Write(e.buf, binary.BigEndian, v)
}

// WriteUint64 writes a uint64 in big-endian
func (e *Encoder) WriteUint64(v uint64) error {
	return binary.Write(e.buf, binary.BigEndian, v)
}

// WriteBytes writes a byte slice with length prefix
func (e *Encoder) WriteBytes(data []byte) error {
	if err := e.WriteUint32(uint32(len(data))); err != nil {
		return err
	}
	_, err := e.buf.Write(data)
	return err
}

// WriteString writes a string with length prefix
func (e *Encoder) WriteString(s string) error {
	return e.WriteBytes([]byte(s))
}

// WriteBigInt writes a big.Int with length prefix
func (e *Encoder) WriteBigInt(v *big.Int) error {
	if v == nil {
		return e.WriteBytes(nil)
	}
	return e.WriteBytes(v.Bytes())
}

// WriteFixedBytes writes a fixed-size byte slice
func (e *Encoder) WriteFixedBytes(data []byte, size int) error {
	if len(data) != size {
		return fmt.Errorf("expected %d bytes, got %d", size, len(data))
	}
	_, err := e.buf.Write(data)
	return err
}

// Bytes returns the encoded bytes
func (e *Encoder) Bytes() []byte {
	return e.buf.Bytes()
}

// Decoder provides binary decoding utilities
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a new decoder
func NewDecoder(data []byte) *Decoder {
	return &Decoder{r: bytes.NewReader(data)}
}

// ReadUint8 reads a uint8
func (d *Decoder) ReadUint8() (uint8, error) {
	var v uint8
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

// ReadUint16 reads a uint16
func (d *Decoder) ReadUint16() (uint16, error) {
	var v uint16
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

// ReadUint32 reads a uint32
func (d *Decoder) ReadUint32() (uint32, error) {
	var v uint32
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

// ReadUint64 reads a uint64
func (d *Decoder) ReadUint64() (uint64, error) {
	var v uint64
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

// ReadBytes reads a byte slice with length prefix
func (d *Decoder) ReadBytes() ([]byte, error) {
	length, err := d.ReadUint32()
	if err != nil {
		return nil, err
	}
	data := make([]byte, length)
	_, err = io.ReadFull(d.r, data)
	return data, err
}

// ReadString reads a string with length prefix
func (d *Decoder) ReadString() (string, error) {
	data, err := d.ReadBytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadBigInt reads a big.Int with length prefix
func (d *Decoder) ReadBigInt() (*big.Int, error) {
	data, err := d.ReadBytes()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return big.NewInt(0), nil
	}
	return new(big.Int).SetBytes(data), nil
}

// ReadFixedBytes reads a fixed-size byte slice
func (d *Decoder) ReadFixedBytes(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := io.ReadFull(d.r, data)
	return data, err
}

// Hex encoding utilities

// EncodeHex encodes bytes to hex string with 0x prefix
func EncodeHex(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}

// DecodeHex decodes hex string (with or without 0x prefix)
func DecodeHex(s string) ([]byte, error) {
	if len(s) >= 2 && s[0:2] == "0x" {
		s = s[2:]
	}
	return hex.DecodeString(s)
}

// MustDecodeHex decodes hex or panics
func MustDecodeHex(s string) []byte {
	data, err := DecodeHex(s)
	if err != nil {
		panic(err)
	}
	return data
}

// JSON utilities

// ToJSON converts a value to JSON bytes
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// ToJSONIndent converts a value to indented JSON bytes
func ToJSONIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// FromJSON parses JSON bytes into a value
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// CopyBytes returns a copy of a byte slice
func CopyBytes(data []byte) []byte {
	if data == nil {
		return nil
	}
	cpy := make([]byte, len(data))
	copy(cpy, data)
	return cpy
}

// PadBytes pads a byte slice to the specified length
func PadBytes(data []byte, length int) []byte {
	if len(data) >= length {
		return data
	}
	result := make([]byte, length)
	copy(result[length-len(data):], data)
	return result
}

// TrimBytes removes leading zero bytes
func TrimBytes(data []byte) []byte {
	for i, b := range data {
		if b != 0 {
			return data[i:]
		}
	}
	return []byte{0}
}
