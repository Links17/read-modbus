package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/goburrow/modbus"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	START_ADDRESS = 0x8007
	BATCH_SIZE    = 100
)

var lastExecutionTime time.Time

func readModbus(w http.ResponseWriter, r *http.Request) {
	// 1. 创建 Modbus RTU 句柄
	currentTime := time.Now()
	handler := modbus.NewRTUClientHandler("/dev/ttyAMA3")
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 10 * time.Second
	newBase64Data := ""
	//var cacheInt = 0
	// 打开连接
	if err := handler.Connect(); err != nil {
		log.Fatalf("无法创建 Modbus RTU 连接: %v\n", err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	// 4. 读取寄存器 0x8002
	// 读取寄存器 0x8002
	results, err := client.ReadHoldingRegisters(0x1000, 2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取置信度失败: %v\n", err)
		return
	}
	fmt.Printf("读取置信度成功")
	fmt.Printf(string(results))

	//cacheInt = int(results[0])

	results, err = client.ReadHoldingRegisters(0x2000, 2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取寄存器坐标失败: %v\n", err)
		return
	}
	fmt.Println(results)
	bigEndianBytes1 := []byte{results[2], results[3], results[0], results[1]}

	// 转换为十进制数
	value1 := binary.BigEndian.Uint32(bigEndianBytes1)
	fmt.Printf("坐标十进制数: %d\n", value1)

	results, err = client.ReadHoldingRegisters(0x3000, 2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取区域失败: %v\n", err)
		return
	}
	bigEndianBytes := []byte{results[2], results[3], results[0], results[1]}

	// 转换为十进制数
	value := binary.BigEndian.Uint32(bigEndianBytes)

	fmt.Printf("区域: %d\n", value)

	// 拍照,并且吐出图片
	if currentTime.Sub(lastExecutionTime) >= 20*time.Second {
		lastExecutionTime = currentTime
		results, err = client.ReadHoldingRegisters(0x8002, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "拍照触发失败: %v\n", err)
			return
		}
		fmt.Printf("拍照触发失败，值为: %d", results[0])

		time.Sleep(3 * time.Second)

		results, err = client.ReadHoldingRegisters(0x8006, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取保持寄存器失败: %v\n", err)
			return
		}

		// 将结果转换为 uint16
		var count_register uint16
		count_register = uint16(results[0])<<8 | uint16(results[1])

		// 存储要读取的寄存器数量
		reg_count := count_register
		fmt.Printf("读取保持寄存器 0x8006 成功，获取的寄存器数量为: %d\n", reg_count)
		// 分配内存以保存所有寄存器数据
		registers := make([]uint16, reg_count)

		// 分批读取保持寄存器
		startAddress := uint16(START_ADDRESS)
		remainingRegs := int(reg_count)
		totalRead := 0

		for remainingRegs > 0 {
			readSize := BATCH_SIZE
			if remainingRegs < BATCH_SIZE {
				readSize = remainingRegs
			}

			results, err = client.ReadHoldingRegisters(startAddress, uint16(readSize))
			if err != nil {
				fmt.Fprintf(os.Stderr, "读取保持寄存器失败: %v\n", err)
				return
			}

			for i := 0; i < readSize; i++ {
				registers[totalRead+i] = uint16(results[2*i])<<8 | uint16(results[2*i+1])
			}

			totalRead += readSize
			remainingRegs -= readSize
			startAddress += uint16(readSize)
		}

		var builder strings.Builder

		for _, regValue := range registers {
			flippedValue := (regValue << 8) | (regValue >> 8)
			highByte := byte(flippedValue >> 8)
			lowByte := byte(flippedValue & 0xFF)
			builder.WriteByte(highByte)
			builder.WriteByte(lowByte)
		}

		// 去除MIME前缀并解码base64数据
		//b64data := base64ImageData[strings.IndexByte(base64ImageData, ',')+1:]
		unbased, err := base64.StdEncoding.DecodeString(builder.String())
		if err != nil {
			fmt.Println("Error decoding base64:", err)
			return
		}

		// 解码图像
		imgReader := bytes.NewReader(unbased)
		img, err := imaging.Decode(imgReader)
		if err != nil {
			fmt.Println("Error decoding image with imaging library:", err)
			return
		}

		// Create a new RGBA image to draw on
		bounds := img.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, image.Point{0, 0}, draw.Src)

		xy, err := parseXY(strconv.Itoa(int(value1)))
		if err != nil {
			fmt.Println("Error decoding image with imaging library:", err)
			return
		}

		wh, err := parseWH(strconv.Itoa(int(value)))
		if err != nil {
			fmt.Println("Error decoding image with imaging library:", err)
			return
		}
		// Define the rectangle's properties
		bboxX := xy.X - (wh.W / 2)
		bboxY := xy.Y - (wh.H / 2)
		bboxWidth := wh.W
		bboxHeight := wh.H

		// Define the color (red) and line width for the rectangle
		col := color.RGBA{255, 0, 0, 255}
		lineWidth := 2

		// Draw the top line of the rectangle
		for x := bboxX; x < bboxX+bboxWidth; x++ {
			for l := 0; l < lineWidth; l++ {
				rgba.Set(x, bboxY+l, col)
				rgba.Set(x, bboxY+bboxHeight-l, col)
			}
		}

		// Draw the left and right sides of the rectangle
		for y := bboxY; y < bboxY+bboxHeight; y++ {
			for l := 0; l < lineWidth; l++ {
				rgba.Set(bboxX+l, y, col)
				rgba.Set(bboxX+bboxWidth-l, y, col)
			}
		}

		// Encode the image to a new base64 string
		var buff bytes.Buffer
		png.Encode(&buff, rgba)
		newBase64Data = base64.StdEncoding.EncodeToString(buff.Bytes())
	}

	m := map[string]interface{}{
		"img":         newBase64Data,
		"area":        value,
		"coordinates": value1,
	}
	// 序列化结果为JSON
	jsonData, err := json.Marshal(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("无法序列化结果: %v", err), http.StatusInternalServerError)
		return
	}

	// 写入响应
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

type Coordinate struct {
	X, Y int
}

type Dimension struct {
	W, H int
}

// 解析 "xy:118132" 形式的字符串
func parseXY(input string) (Coordinate, error) {
	re := regexp.MustCompile(`(\d+)(\d{3})$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 3 {
		return Coordinate{}, fmt.Errorf("invalid format")
	}

	x, err := strconv.Atoi(matches[1])
	if err != nil {
		return Coordinate{}, err
	}

	y, err := strconv.Atoi(matches[2])
	if err != nil {
		return Coordinate{}, err
	}

	return Coordinate{X: x, Y: y}, nil
}

// 解析 "wh:97104" 形式的字符串
func parseWH(input string) (Dimension, error) {
	re := regexp.MustCompile(`(\d+)(\d{3})$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 3 {
		return Dimension{}, fmt.Errorf("invalid format")
	}

	w, err := strconv.Atoi(matches[1])
	if err != nil {
		return Dimension{}, err
	}

	h, err := strconv.Atoi(matches[2])
	if err != nil {
		return Dimension{}, err
	}

	return Dimension{W: w, H: h}, nil
}

func main() {
	http.HandleFunc("/read-modbus", readModbus)

	fmt.Println("HTTP server started on :2005")
	err := http.ListenAndServe("0.0.0.0:2005", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
