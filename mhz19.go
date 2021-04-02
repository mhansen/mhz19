// Datasheet: https://www.winsen-sensor.com/d/files/PDF/Infrared%20Gas%20Sensor/NDIR%20CO2%20SENSOR/MH-Z19%20CO2%20Ver1.0.pdf
package mhz19

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// GasConcentrationRequest is documented at https://www.winsen-sensor.com/d/files/PDF/Infrared%20Gas%20Sensor/NDIR%20CO2%20SENSOR/MH-Z19%20CO2%20Ver1.0.pdf
type GasConcentrationRequest struct {
	Start    byte
	SensorNo byte
	Command  byte
	Byte3    byte
	Byte4    byte
	Byte5    byte
	Byte6    byte
	Byte7    byte
	Checksum byte
}

func NewGasConcentrationRequest() *GasConcentrationRequest {
	return &GasConcentrationRequest{
		Start:    0xFF,
		SensorNo: 0x01,
		Command:  0x86,
		Checksum: 0x79,
	}
}

func (r *GasConcentrationRequest) Write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, r)
}

// GasConcentrationResponse is documented at https://www.winsen-sensor.com/d/files/PDF/Infrared%20Gas%20Sensor/NDIR%20CO2%20SENSOR/MH-Z19%20CO2%20Ver1.0.pdf
type GasConcentrationResponse struct {
	Start             byte // should always be 0xFF
	Command           byte // should always be 0x86
	Concentration     uint16
	OffsetTemperature byte // Temperature in Celsius + 40
	_                 byte
	_                 byte
	_                 byte
	Checksum          byte
}

// Temperature returns sensor reading in Celsius
func (r *GasConcentrationResponse) Temperature() int {
	return int(r.OffsetTemperature) - 40
}

func ReadGasConcentrationResponse(r io.Reader) (*GasConcentrationResponse, error) {
	len := 9
	buf := make([]byte, len)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	if n != len {
		return nil, fmt.Errorf("too few bytes read: want=%d got=%d", len, n)
	}

	var sum byte = 0
	for i := 0; i < 8; i++ {
		sum += buf[i]
	}
	sum = 0xff - sum

	var resp GasConcentrationResponse
	binary.Read(bytes.NewReader(buf), binary.BigEndian, &resp)

	if sum != resp.Checksum {
		return &resp, &ChecksumError{fmt.Sprintf("checksum failed: got %v want %v", sum, resp.Checksum)}
	}
	return &resp, nil
}

type ChecksumError struct {
	err string
}

func (e *ChecksumError) Error() string {
	return e.err
}
