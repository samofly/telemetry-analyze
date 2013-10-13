package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
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
	Timestamp time.Duration
	Acc       Point3d
	Gyro      Point3d
}

func readFloat32(in []byte) (out []byte, res float32, err error) {
	if len(in) < 4 {
		err = io.ErrUnexpectedEOF
		return
	}
	out = in[4:]
	num := uint32(in[0]) + uint32(in[1])<<8 + uint32(in[2])<<16 + uint32(in[3])<<24
	res = math.Float32frombits(num)
	return
}

func readPoint3d(in []byte) (out []byte, p Point3d, err error) {
	out = in
	for i := range p {
		var v float32
		if out, v, err = readFloat32(out); err != nil {
			return
		}
		p[i] = float64(v)
	}
	return
}

func readLogPoint(num []byte) (res *LogPoint, err error) {
	res = new(LogPoint)
	if len(num) < 2+2*3*4 {
		return nil, io.ErrUnexpectedEOF
	}
	num = num[2:]
	if num, res.Gyro, err = readPoint3d(num); err != nil {
		return
	}
	if num, res.Acc, err = readPoint3d(num); err != nil {
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
	for i, line := range strings.Split(string(data), "\n") {
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "0 133 ") {
			continue
		}
		fmt.Println(line)
		num, err := readBytes(line)
		if err != nil {
			log.Fatalf("Failed to parse line #%d: %s, err: %v", i, line, err)
		}
		p, err := readLogPoint(num)
		if err != nil {
			log.Fatalf("Failed to read log point at line #%d: %s, err: %v", i, line, err)
		}
		fmt.Printf("%+v\n", p)
	}
}
