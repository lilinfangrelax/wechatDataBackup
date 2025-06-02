package wechat

/**
这段代码是一个Go语言实现的并发文件处理程序，主要用于处理微信图片备份。我来逐步解释它的功能：

1. **错误处理**：
   - 首先检查前一个操作是否有错误(er)，如果有就直接返回错误

2. **读取目录内容**：
   - 使用`f.Readdir(0)`读取目录中的所有文件和子目录信息
   - 如果有错误则打印错误信息

3. **输出目录检查与创建**：
   - 检查输出目录(outDir)是否存在
   - 如果不存在则创建目录(权限0755)
   - 如果存在但不是目录则返回错误

4. **并发处理机制**：
   - 创建一个缓冲通道`taskChan`(容量100)用于传递文件信息
   - 启动一个goroutine将读取到的所有文件信息发送到通道
   - 发送完成后关闭通道

5. **工作池模式**：
   - 使用`sync.WaitGroup`来等待所有工作完成
   - 启动10个worker goroutine并发处理文件
   - 每个worker从通道获取文件信息并调用`handlerOne`处理

6. **完成处理**：
   - 等待所有worker完成(`wg.Wait()`)
   - 计算并打印总耗时
   - 返回nil表示成功

主要特点：
- 使用通道实现生产者-消费者模式
- 采用工作池(10个worker)控制并发度
- 使用WaitGroup同步goroutine
- 包含完善的错误处理
- 支持大目录处理(通道缓冲100)

这段代码的核心是将目录中的文件分发给多个worker并发处理，适合IO密集型任务。</think>
**/

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
	from: https://github.com/liuggchen/wechatDatDecode.git
*/

var imagePrefixBtsMap = make(map[string][]byte)

func DecryptDatByDir(inDir, outDir string) error {

	startTime := time.Now()
	f, er := os.Open(inDir)
	if er != nil {
		fmt.Println(er.Error())
		return er
	}
	readdir, er := f.Readdir(0)
	if er != nil {
		fmt.Println(er.Error())
	}

	if stat, er := os.Stat(outDir); os.IsNotExist(er) {
		er := os.MkdirAll(outDir, 0755)
		if er != nil {
			return er
		}
	} else if !stat.IsDir() {
		return errors.New(outDir + "is file")
	}

	var taskChan = make(chan os.FileInfo, 100)

	go func() {
		for _, info := range readdir {
			taskChan <- info
		}
		close(taskChan)
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for info := range taskChan {
				handlerOne(info, inDir, outDir)
			}
		}()
	}

	wg.Wait()
	t := time.Since(startTime).Seconds()
	log.Printf("\nfinished time= %v s\n", t)

	return nil
}

func DecryptDat(inFile string, outFile string) error {

	sourceFile, err := os.Open(inFile)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var preTenBts = make([]byte, 10)
	_, _ = sourceFile.Read(preTenBts)
	decodeByte, _, er := findDecodeByte(preTenBts)
	if er != nil {
		log.Println(er.Error())
		return err
	}

	distFile, er := os.Create(outFile)
	if er != nil {
		log.Println(er.Error())
		return err
	}
	writer := bufio.NewWriter(distFile)
	_, _ = sourceFile.Seek(0, 0)
	var rBts = make([]byte, 1024)
	for {
		n, er := sourceFile.Read(rBts)
		if er != nil {
			if er == io.EOF {
				break
			}
			log.Println("error: ", er.Error())
			return err
		}
		for i := 0; i < n; i++ {
			_ = writer.WriteByte(rBts[i] ^ decodeByte)
		}
	}
	_ = writer.Flush()
	_ = distFile.Close()
	_ = sourceFile.Close()
	// fmt.Println("output file：", distFile.Name())

	return nil
}

func handlerOne(info os.FileInfo, dir string, outputDir string) {
	if info.IsDir() || filepath.Ext(info.Name()) != ".dat" {
		return
	}
	fmt.Println("find file: ", info.Name())
	fPath := dir + "/" + info.Name()
	sourceFile, err := os.Open(fPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var preTenBts = make([]byte, 10)
	_, _ = sourceFile.Read(preTenBts)
	decodeByte, _, er := findDecodeByte(preTenBts)
	if er != nil {
		fmt.Println(er.Error())
		return
	}

	distFile, er := os.Create(outputDir + "/" + info.Name())
	if er != nil {
		fmt.Println(er.Error())
		return
	}
	writer := bufio.NewWriter(distFile)
	_, _ = sourceFile.Seek(0, 0)
	var rBts = make([]byte, 1024)
	for {
		n, er := sourceFile.Read(rBts)
		if er != nil {
			if er == io.EOF {
				break
			}
			fmt.Println("error: ", er.Error())
			return
		}
		for i := 0; i < n; i++ {
			_ = writer.WriteByte(rBts[i] ^ decodeByte)
		}
	}
	_ = writer.Flush()
	_ = distFile.Close()

	fmt.Println("output file：", distFile.Name())
}

func init() {
	//JPEG (jpg)，文件头：FFD8FF
	//PNG (png)，文件头：89504E47
	//GIF (gif)，文件头：47494638
	//TIFF (tif)，文件头：49492A00
	//Windows Bitmap (bmp)，文件头：424D
	const (
		Jpeg = "FFD8FF"
		Png  = "89504E47"
		Gif  = "47494638"
		Tif  = "49492A00"
		Bmp  = "424D"
	)
	JpegPrefixBytes, _ := hex.DecodeString(Jpeg)
	PngPrefixBytes, _ := hex.DecodeString(Png)
	GifPrefixBytes, _ := hex.DecodeString(Gif)
	TifPrefixBytes, _ := hex.DecodeString(Tif)
	BmpPrefixBytes, _ := hex.DecodeString(Bmp)

	imagePrefixBtsMap = map[string][]byte{
		".jpeg": JpegPrefixBytes,
		".png":  PngPrefixBytes,
		".gif":  GifPrefixBytes,
		".tif":  TifPrefixBytes,
		".bmp":  BmpPrefixBytes,
	}
}

func findDecodeByte(bts []byte) (byte, string, error) {
	for ext, prefixBytes := range imagePrefixBtsMap {
		deCodeByte, err := testPrefix(prefixBytes, bts)
		if err == nil {
			return deCodeByte, ext, err
		}
	}
	return 0, "", errors.New("decode fail")
}

func testPrefix(prefixBytes []byte, bts []byte) (deCodeByte byte, error error) {
	var initDecodeByte = prefixBytes[0] ^ bts[0]
	for i, prefixByte := range prefixBytes {
		if b := prefixByte ^ bts[i]; b != initDecodeByte {
			return 0, errors.New("no")
		}
	}
	return initDecodeByte, nil
}
