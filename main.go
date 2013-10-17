package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func readBytes(line string) (res []byte, err error) {
	for _, elem := range strings.Split(line, " ") {
		if len(elem) == 0 {
			continue
		}
		num, err := strconv.ParseUint(elem, 10, 8)
		if err != nil {
			return nil, err
		}
		res = append(res, byte(num))
	}
	return
}

type Point3d [3]float64

type LogPoint struct {
	Index     int
	Timestamp time.Duration
	Acc       Point3d
	Gyro      Point3d
}

func readInt16(in []byte) (out []byte, v int16, err error) {
	if len(in) < 2 {
		err = io.ErrUnexpectedEOF
		return
	}
	out = in[2:]
	v = int16(in[0]) + int16(in[1])<<8
	return
}

func readUint32(in []byte) (out []byte, v uint32, err error) {
	if len(in) < 4 {
		err = io.ErrUnexpectedEOF
		return
	}
	out = in[4:]
	v = uint32(in[0]) + uint32(in[1])<<8 + uint32(in[2])<<16 + uint32(in[3])<<24
	return
}

func readPoint3d(in []byte, k float64) (out []byte, p Point3d, err error) {
	out = in
	for i := range p {
		var v int16
		if out, v, err = readInt16(out); err != nil {
			return
		}
		p[i] = float64(v) * k
	}
	return
}

func readLogPoint(num []byte) (res *LogPoint, err error) {
	res = new(LogPoint)
	if len(num) < 2+2*2*3 {
		return nil, io.ErrUnexpectedEOF
	}
	num = num[2:]

	var index uint32
	if num, index, err = readUint32(num); err != nil {
		return
	}
	res.Index = int(index)

	if num, res.Acc, err = readPoint3d(num, 2*8.0/65536); err != nil {
		return
	}
	if num, res.Gyro, err = readPoint3d(num, 2*500.0/65536); err != nil {
		return
	}
	var t int64
	for i, v := range num {
		t += int64(v) << (8 * uint(i))
	}
	res.Timestamp = time.Duration(t) * time.Microsecond
	return
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: %s <path/to/telemetry/dump", os.Args[0])
	}
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("index\tusec\taccX\taccY\taccZ\tgyroX\tgyroY\tgyroZ")
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "0 133 ") {
			continue
		}
		num, err := readBytes(line)
		if err != nil {
			log.Fatalf("Failed to parse line #%d: %s, err: %v", i, line, err)
		}
		p, err := readLogPoint(num)
		if err != nil {
			log.Fatalf("Failed to read log point at line #%d: %s, err: %v", i, line, err)
		}
		fmt.Printf("%d\t%d\t%f\t%f\t%f\t%f\t%f\t%f\n", p.Index, int64(p.Timestamp.Nanoseconds()/1000),
			p.Acc[0], p.Acc[1], p.Acc[2], p.Gyro[0], p.Gyro[1], p.Gyro[2])
	}
}
