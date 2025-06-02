package wechat // 定义包名为 wechat，包含微信数据处理相关的功能。
/**
这段代码是一个Go语言函数的一部分，主要用于处理微信数据的备份，特别是微信头像的导出。我来逐步解释这段代码：

1. `info := WeChatInfo{}`  
   创建一个空的`WeChatInfo`结构体实例，可能用于存储微信相关信息。

2. 检查Misc.db文件是否存在：
   ```go
   miscDBPath := fmt.Sprintf("%s\\Msg\\Misc.db", exportPath)
   _, err := os.Stat(miscDBPath)
   if err != nil {
       log.Println("no exist:", miscDBPath)
       return
   }
   ```
   - 构建Misc.db文件的路径
   - 检查该文件是否存在
   - 如果不存在则记录日志并返回

3. 检查HeadImage目录是否存在：
   ```go
   headImgPath := fmt.Sprintf("%s\\FileStorage\\HeadImage", exportPath)
   if _, err := os.Stat(headImgPath); err == nil {
       log.Println("has HeadImage")
       return
   }
   ```
   - 构建头像存储目录路径
   - 如果目录已存在则记录日志并返回（可能是为了避免重复导出）

4. 启动goroutine导出头像：
   ```go
   go func() {
       exportWeChatHeadImage(info, exportPath, progress)
       close(progress)
   }()
   ```
   - 启动一个新的goroutine异步执行头像导出
   - 调用`exportWeChatHeadImage`函数进行实际导出
   - 导出完成后关闭progress通道

5. 处理进度信息：
   ```go
   for p := range progress {
       log.Println(p)
   }
   log.Println("ExportWeChatHeadImage done")
   ```
   - 从progress通道读取并打印导出进度信息
   - 当通道关闭后循环结束，打印完成信息

这段代码的主要流程是：
1. 检查必要的文件是否存在
2. 检查是否已经导出了头像
3. 异步导出微信头像
4. 同步处理并显示导出进度

注意：这段代码似乎缺少一些上下文，比如`progress`通道的定义、`exportWeChatHeadImage`函数的实现等。


这段代码位于 [`pkg\wechat\wechat.go:964-966`](pkg\wechat\wechat.go:964-966)，其作用是在一个新的 Goroutine 中异步执行微信头像导出操作，并在操作完成后关闭进度通道。

以下是详细解释：

1.  **目的和功能**:
    *   **目的**: 这段代码的目的是在后台（即非阻塞地）执行 `exportWeChatHeadImage` 函数，该函数负责将微信头像导出到指定路径。
    *   **功能**:
        *   `go func() { ... }()`: 启动一个新的 Goroutine。这意味着 `exportWeChatHeadImage` 函数将在一个独立的并发执行单元中运行，而不会阻塞调用 `ExportWeChatHeadImage` 函数的主 Goroutine。
        *   `exportWeChatHeadImage(info, exportPath, progress)`: 调用实际的头像导出函数。它接收微信信息 (`info`)、导出路径 (`exportPath`) 和一个用于报告进度的通道 (`progress`) 作为参数。
        *   `close(progress)`: 在 `exportWeChatHeadImage` 函数执行完毕后，关闭 `progress` 通道。这向所有监听 `progress` 通道的接收者发出信号，表明不会再有数据发送到此通道。这对于主 Goroutine 中 `for p := range progress` 循环的正常终止至关重要，因为它依赖于通道的关闭来退出循环。

2.  **关键组件及其交互**:
    *   **`go func()`**: Go 语言中用于创建并发执行单元的关键字。它使得 `exportWeChatHeadImage` 的执行与程序的其余部分并行。
    *   **`exportWeChatHeadImage` 函数**: 这是一个耗时操作，负责从微信数据中提取并保存头像。它通过 `progress` 通道向外部报告其执行状态和进度。
    *   **`progress chan<- string`**: 这是一个只发送字符串的通道。它充当了 `exportWeChatHeadImage` Goroutine 和调用它的主 Goroutine 之间的通信桥梁。`exportWeChatHeadImage` 会通过这个通道发送 JSON 格式的进度信息（例如 `{"status":"processing", "result":"export WeChat Head Image", "progress": 81}`）。
    *   **`close(progress)`**: 这是通道管理的关键部分。当 `exportWeChatHeadImage` 完成所有工作后，它会关闭 `progress` 通道。这使得主 Goroutine 中监听 `progress` 通道的 `for range` 循环能够检测到通道关闭并优雅地退出，从而避免无限等待。

3.  **重要的模式或技术**:
    *   **并发编程 (Goroutines)**: 使用 `go` 关键字启动 Goroutine 是 Go 语言实现并发的核心方式。它允许程序同时执行多个任务，提高效率和响应性。
    *   **通道 (Channels) 进行通信**: `progress` 通道是 Goroutine 之间安全、同步通信的 Go 惯用方式。它避免了共享内存可能导致的竞态条件，使得并发代码更易于编写和理解。
    *   **异步操作**: 整个结构体现了异步操作模式。头像导出是一个可能需要较长时间的任务，通过将其放入 Goroutine 中异步执行，主程序可以继续执行其他任务，并在需要时通过通道获取进度更新。
    *   **通道关闭作为完成信号**: `close(progress)` 是一种常见的模式，用于向消费者 Goroutine 发送任务完成的信号。当通道关闭时，`for range` 循环会完成迭代，从而允许消费者 Goroutine 知道所有数据都已处理完毕。
**/
import (
	"bufio"        // 导入 bufio 包，用于带缓冲的 I/O 操作。
	"bytes"        // 导入 bytes 包，用于处理字节切片。
	"crypto/hmac"  // 导入 crypto/hmac 包，用于 HMAC 消息认证码。
	"crypto/sha1"  // 导入 crypto/sha1 包，用于 SHA-1 哈希算法。
	"database/sql" // 导入 database/sql 包，提供了通用的 SQL 数据库接口。
	"encoding/binary" // 导入 encoding/binary 包，用于二进制数据的编码和解码。
	"encoding/hex" // 导入 encoding/hex 包，用于十六进制编码和解码。
	"errors"       // 导入 errors 包，用于创建和处理错误。
	"fmt"          // 导入 fmt 包，用于格式化输入输出。
	"io"           // 导入 io 包，提供了基本的 I/O 接口。
	"log"          // 导入 log 包，用于记录日志。
	"os"           // 导入 os 包，提供了与操作系统交互的函数。
	"path/filepath" // 导入 path/filepath 包，用于处理文件路径。
	"strings"      // 导入 strings 包，用于字符串操作。
	"sync"         // 导入 sync 包，提供了基本的同步原语。
	"sync/atomic"  // 导入 sync/atomic 包，提供了原子操作。
	"time"         // 导入 time 包，用于时间操作。
	"unsafe"       // 导入 unsafe 包，用于不安全的操作，如指针转换。

	"github.com/git-jiadong/go-lame" // 导入 go-lame 包，用于 MP3 编码。
	"github.com/git-jiadong/go-silk" // 导入 go-silk 包，用于 SILK 音频解码。
	_ "github.com/mattn/go-sqlite3"  // 导入 go-sqlite3 驱动，用于 SQLite 数据库操作。
	"github.com/shirou/gopsutil/v3/process" // 导入 gopsutil/process 包，用于获取进程信息。
	"golang.org/x/sys/windows"       // 导入 golang.org/x/sys/windows 包，用于 Windows 系统调用。
)

// WeChatInfo 结构体定义了单个微信实例的详细信息。
type WeChatInfo struct {
	ProcessID   uint32    // 微信进程的 ID。
	FilePath    string    // 微信文件路径。
	AcountName  string    // 微信账户名。
	Version     string    // 微信版本号。
	Is64Bits    bool      // 微信进程是否是 64 位。
	DllBaseAddr uintptr   // 微信 DLL 模块的基地址。
	DllBaseSize uint32    // 微信 DLL 模块的大小。
	DBKey       string    // 微信数据库的密钥。
}

// WeChatInfoList 结构体定义了微信信息列表，包含多个 WeChatInfo 实例。
type WeChatInfoList struct {
	Info  []WeChatInfo `json:"Info"`  // 微信信息切片。
	Total int          `json:"Total"` // 微信信息总数。
}

// wechatMediaMSG 结构体用于存储微信媒体消息的相关信息。
type wechatMediaMSG struct {
	Key      string // 消息的 Key。
	MsgSvrID int    // 消息服务器 ID。
	Buf      []byte // 消息内容缓冲区。
}

// wechatHeadImgMSG 结构体用于存储微信头像消息的相关信息。
type wechatHeadImgMSG struct {
	userName string // 用户名。
	Buf      []byte // 头像图片缓冲区。
}

// GetWeChatAllInfo 函数用于获取所有微信实例的详细信息，包括数据库密钥。
// 返回一个 WeChatInfoList 结构体指针。
func GetWeChatAllInfo() *WeChatInfoList {
	list := GetWeChatInfo() // 获取微信基本信息列表。

	for i := range list.Info {
		list.Info[i].DBKey = GetWeChatKey(&list.Info[i]) // 为每个微信实例获取数据库密钥。
	}

	return list // 返回包含所有信息的列表。
}

// ExportWeChatAllData 函数用于导出指定微信账户的所有数据。
// info 参数是微信信息，expPath 参数是导出路径，progress 通道用于报告导出进度。
func ExportWeChatAllData(info WeChatInfo, expPath string, progress chan<- string) {
	defer close(progress) // 确保在函数返回时关闭进度通道。
	fileInfo, err := os.Stat(info.FilePath) // 获取微信文件路径的信息。
	if err != nil || !fileInfo.IsDir() {
		progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%s error\"}", info.FilePath) // 如果文件路径无效，发送错误信息。
		return
	}
	if !exportWeChatDateBase(info, expPath, progress) { // 导出微信数据库。
		return
	}

	exportWeChatBat(info, expPath, progress)         // 导出微信 Dat 文件。
	exportWeChatVideoAndFile(info, expPath, progress) // 导出微信视频和文件。
	exportWeChatVoice(info, expPath, progress)       // 导出微信语音。
	exportWeChatHeadImage(info, expPath, progress)   // 导出微信头像。
}

// exportWeChatHeadImage 函数用于导出微信头像。
// info 参数是微信信息，expPath 参数是导出路径，progress 通道用于报告导出进度。
func exportWeChatHeadImage(info WeChatInfo, expPath string, progress chan<- string) {
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat Head Image\", \"progress\": 81}" // 发送进度信息。

	headImgPath := fmt.Sprintf("%s\\FileStorage\\HeadImage", expPath) // 构建头像导出路径。
	if _, err := os.Stat(headImgPath); err != nil {
		if err := os.MkdirAll(headImgPath, 0644); err != nil { // 如果目录不存在，则创建。
			log.Printf("MkdirAll %s failed: %v\n", headImgPath, err) // 打印创建目录失败日志。
			progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%v error\"}", err) // 发送错误信息。
			return
		}
	}

	handleNumber := int64(0) // 已处理文件数量。
	fileNumber := int64(0)   // 总文件数量。

	var wg sync.WaitGroup     // 用于等待所有 Goroutine 完成。
	var reportWg sync.WaitGroup // 用于等待报告 Goroutine 完成。
	quitChan := make(chan struct{}) // 用于通知报告 Goroutine 退出。
	MSGChan := make(chan wechatHeadImgMSG, 100) // 消息通道，用于传递头像消息。
	go func() { // 在新的 Goroutine 中读取数据库并发送消息。
		for {
			miscDBPath := fmt.Sprintf("%s\\Msg\\Misc.db", expPath) // 构建 Misc.db 路径。
			_, err := os.Stat(miscDBPath)
			if err != nil {
				log.Println("no exist:", miscDBPath) // 如果 Misc.db 不存在，打印日志并退出。
				break
			}

			db, err := sql.Open("sqlite3", miscDBPath) // 打开 Misc.db 数据库。
			if err != nil {
				log.Printf("open %s failed: %v\n", miscDBPath, err) // 打印打开数据库失败日志。
				break
			}
			defer db.Close() // 确保在函数返回时关闭数据库连接。

			err = db.QueryRow("select count(*) from ContactHeadImg1;").Scan(&fileNumber) // 查询头像总数。
			if err != nil {
				log.Println("select count(*) failed", err) // 打印查询失败日志。
				break
			}
			log.Println("ContactHeadImg1 fileNumber", fileNumber) // 打印头像总数。
			rows, err := db.Query("select ifnull(usrName,'') as usrName, ifnull(smallHeadBuf,'') as smallHeadBuf from ContactHeadImg1;") // 查询头像数据。
			if err != nil {
				log.Printf("Query failed: %v\n", err) // 打印查询失败日志。
				break
			}

			msg := wechatHeadImgMSG{}
			for rows.Next() { // 遍历查询结果。
				err := rows.Scan(&msg.userName, &msg.Buf) // 扫描数据到结构体。
				if err != nil {
					log.Println("Scan failed: ", err) // 打印扫描失败日志。
					break
				}

				MSGChan <- msg // 将消息发送到通道。
			}
			break
		}
		close(MSGChan) // 关闭消息通道。
	}()

	for i := 0; i < 20; i++ { // 启动 20 个 Goroutine 并发处理头像导出。
		wg.Add(1)
		go func() {
			defer wg.Done() // 确保 Goroutine 完成时通知 WaitGroup。
			for msg := range MSGChan { // 从消息通道接收消息。
				imgPath := fmt.Sprintf("%s\\%s.headimg", headImgPath, msg.userName) // 构建头像图片路径。
				for {
					// log.Println("imgPath:", imgPath, len(msg.Buf))
					_, err := os.Stat(imgPath) // 检查文件是否已存在。
					if err == nil {
						break // 如果已存在，则跳过。
					}
					if len(msg.userName) == 0 || len(msg.Buf) == 0 {
						break // 如果用户名或缓冲区为空，则跳过。
					}
					err = os.WriteFile(imgPath, msg.Buf[:], 0666) // 写入头像文件。
					if err != nil {
						log.Println("WriteFile:", imgPath, err) // 打印写入文件失败日志。
					}
					break
				}
				atomic.AddInt64(&handleNumber, 1) // 原子增加已处理文件数量。
			}
		}()
	}

	reportWg.Add(1)
	go func() { // 在新的 Goroutine 中报告导出进度。
		defer reportWg.Done() // 确保 Goroutine 完成时通知 WaitGroup。
		for {
			select {
			case <-quitChan: // 接收到退出信号。
				log.Println("WeChat Head Image report progress end") // 打印退出日志。
				return
			default:
				if fileNumber != 0 {
					filePercent := float64(handleNumber) / float64(fileNumber) // 计算文件处理百分比。
					totalPercent := 81 + (filePercent * (100 - 81))            // 计算总进度百分比。
					totalPercentStr := fmt.Sprintf("{\"status\":\"processing\", \"result\":\"export WeChat Head Image doing\", \"progress\": %d}", int(totalPercent)) // 构建进度信息。
					progress <- totalPercentStr                                // 发送进度信息。
				}
				time.Sleep(time.Second) // 每秒更新一次进度。
			}
		}
	}()

	wg.Wait()         // 等待所有处理 Goroutine 完成。
	close(quitChan)   // 关闭退出通道，通知报告 Goroutine 退出。
	reportWg.Wait()   // 等待报告 Goroutine 完成。
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat Head Image end\", \"progress\": 100}" // 发送导出完成信息。
}


func exportWeChatVoice(info WeChatInfo, expPath string, progress chan<- string) {
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat voice start\", \"progress\": 61}"

	voicePath := fmt.Sprintf("%s\\FileStorage\\Voice", expPath)
	if _, err := os.Stat(voicePath); err != nil {
		if err := os.MkdirAll(voicePath, 0644); err != nil {
			log.Printf("MkdirAll %s failed: %v\n", voicePath, err)
			progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%v error\"}", err)
			return
		}
	}

	handleNumber := int64(0)
	fileNumber := int64(0)
	index := 0
	for {
		mediaMSGDB := fmt.Sprintf("%s\\Msg\\Multi\\MediaMSG%d.db", expPath, index)
		_, err := os.Stat(mediaMSGDB)
		if err != nil {
			break
		}
		index += 1
		fileNumber += 1
	}

	var wg sync.WaitGroup
	var reportWg sync.WaitGroup
	quitChan := make(chan struct{})
	index = -1
	MSGChan := make(chan wechatMediaMSG, 100)
	go func() {
		for {
			index += 1
			mediaMSGDB := fmt.Sprintf("%s\\Msg\\Multi\\MediaMSG%d.db", expPath, index)
			_, err := os.Stat(mediaMSGDB)
			if err != nil {
				break
			}

			db, err := sql.Open("sqlite3", mediaMSGDB)
			if err != nil {
				log.Printf("open %s failed: %v\n", mediaMSGDB, err)
				continue
			}
			defer db.Close()

			rows, err := db.Query("select Key, Reserved0, Buf from Media;")
			if err != nil {
				log.Printf("Query failed: %v\n", err)
				continue
			}

			msg := wechatMediaMSG{}
			for rows.Next() {
				err := rows.Scan(&msg.Key, &msg.MsgSvrID, &msg.Buf)
				if err != nil {
					log.Println("Scan failed: ", err)
					break
				}

				MSGChan <- msg
			}
			atomic.AddInt64(&handleNumber, 1)
		}
		close(MSGChan)
	}()

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range MSGChan {
				mp3Path := fmt.Sprintf("%s\\%d.mp3", voicePath, msg.MsgSvrID)
				_, err := os.Stat(mp3Path)
				if err == nil {
					continue
				}

				err = silkToMp3(msg.Buf[:], mp3Path)
				if err != nil {
					log.Printf("silkToMp3 %s failed: %v\n", mp3Path, err)
				}
			}
		}()
	}

	reportWg.Add(1)
	go func() {
		defer reportWg.Done()
		for {
			select {
			case <-quitChan:
				log.Println("WeChat voice report progress end")
				return
			default:
				filePercent := float64(handleNumber) / float64(fileNumber)
				totalPercent := 61 + (filePercent * (80 - 61))
				totalPercentStr := fmt.Sprintf("{\"status\":\"processing\", \"result\":\"export WeChat voice doing\", \"progress\": %d}", int(totalPercent))
				progress <- totalPercentStr
				time.Sleep(time.Second)
			}
		}
	}()

	wg.Wait()
	close(quitChan)
	reportWg.Wait()
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat voice end\", \"progress\": 80}"
}

func exportWeChatVideoAndFile(info WeChatInfo, expPath string, progress chan<- string) {
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat Video and File start\", , \"progress\": 41}"
	videoRootPath := info.FilePath + "\\FileStorage\\Video"
	fileRootPath := info.FilePath + "\\FileStorage\\File"
	cacheRootPath := info.FilePath + "\\FileStorage\\Cache"

	rootPaths := []string{videoRootPath, fileRootPath, cacheRootPath}

	handleNumber := int64(0)
	fileNumber := int64(0)
	for _, path := range rootPaths {
		fileNumber += getPathFileNumber(path, "")
	}
	log.Println("VideoAndFile ", fileNumber)

	var wg sync.WaitGroup
	var reportWg sync.WaitGroup
	quitChan := make(chan struct{})
	taskChan := make(chan [2]string, 100)
	go func() {
		for _, rootPath := range rootPaths {
			log.Println(rootPath)
			if _, err := os.Stat(rootPath); err != nil {
				continue
			}
			err := filepath.Walk(rootPath, func(path string, finfo os.FileInfo, err error) error {
				if err != nil {
					log.Printf("filepath.Walk：%v\n", err)
					return err
				}

				if !finfo.IsDir() {
					expFile := expPath + path[len(info.FilePath):]
					_, err := os.Stat(filepath.Dir(expFile))
					if err != nil {
						os.MkdirAll(filepath.Dir(expFile), 0644)
					}

					task := [2]string{path, expFile}
					taskChan <- task
					return nil
				}

				return nil
			})
			if err != nil {
				log.Println("filepath.Walk:", err)
				progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%v\"}", err)
			}
		}
		close(taskChan)
	}()

	for i := 1; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				_, err := os.Stat(task[1])
				if err == nil {
					atomic.AddInt64(&handleNumber, 1)
					continue
				}
				_, err = copyFile(task[0], task[1])
				if err != nil {
					log.Println("DecryptDat:", err)
					progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"copyFile %v\"}", err)
				}
				atomic.AddInt64(&handleNumber, 1)
			}
		}()
	}
	reportWg.Add(1)
	go func() {
		defer reportWg.Done()
		for {
			select {
			case <-quitChan:
				log.Println("WeChat Video and File report progress end")
				return
			default:
				filePercent := float64(handleNumber) / float64(fileNumber)
				totalPercent := 41 + (filePercent * (60 - 41))
				totalPercentStr := fmt.Sprintf("{\"status\":\"processing\", \"result\":\"export WeChat Video and File doing\", \"progress\": %d}", int(totalPercent))
				progress <- totalPercentStr
				time.Sleep(time.Second)
			}
		}
	}()
	wg.Wait()
	close(quitChan)
	reportWg.Wait()
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat Video and File end\", \"progress\": 60}"
}

func exportWeChatBat(info WeChatInfo, expPath string, progress chan<- string) {
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat Dat start\", \"progress\": 21}"
	datRootPath := info.FilePath + "\\FileStorage\\MsgAttach"
	imageRootPath := info.FilePath + "\\FileStorage\\Image"
	rootPaths := []string{datRootPath, imageRootPath}

	handleNumber := int64(0)
	fileNumber := int64(0)
	for i := range rootPaths {
		fileNumber += getPathFileNumber(rootPaths[i], ".dat")
	}
	log.Println("DatFileNumber ", fileNumber)

	var wg sync.WaitGroup
	var reportWg sync.WaitGroup
	quitChan := make(chan struct{})
	taskChan := make(chan [2]string, 100)
	go func() {
		for i := range rootPaths {
			if _, err := os.Stat(rootPaths[i]); err != nil {
				continue
			}

			err := filepath.Walk(rootPaths[i], func(path string, finfo os.FileInfo, err error) error {
				if err != nil {
					log.Printf("filepath.Walk：%v\n", err)
					return err
				}

				if !finfo.IsDir() && strings.HasSuffix(path, ".dat") {
					expFile := expPath + path[len(info.FilePath):]
					_, err := os.Stat(filepath.Dir(expFile))
					if err != nil {
						os.MkdirAll(filepath.Dir(expFile), 0644)
					}

					task := [2]string{path, expFile}
					taskChan <- task
					return nil
				}

				return nil
			})

			if err != nil {
				log.Println("filepath.Walk:", err)
				progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%v\"}", err)
			}
		}
		close(taskChan)
	}()

	for i := 1; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				_, err := os.Stat(task[1])
				if err == nil {
					atomic.AddInt64(&handleNumber, 1)
					continue
				}
				err = DecryptDat(task[0], task[1])
				if err != nil {
					log.Println("DecryptDat:", err)
					progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"DecryptDat %v\"}", err)
				}
				atomic.AddInt64(&handleNumber, 1)
			}
		}()
	}
	reportWg.Add(1)
	go func() {
		defer reportWg.Done()
		for {
			select {
			case <-quitChan:
				log.Println("WeChat Dat report progress end")
				return
			default:
				filePercent := float64(handleNumber) / float64(fileNumber)
				totalPercent := 21 + (filePercent * (40 - 21))
				totalPercentStr := fmt.Sprintf("{\"status\":\"processing\", \"result\":\"export WeChat Dat doing\", \"progress\": %d}", int(totalPercent))
				progress <- totalPercentStr
				time.Sleep(time.Second)
			}
		}
	}()
	wg.Wait()
	close(quitChan)
	reportWg.Wait()
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat Dat end\", \"progress\": 40}"
}

func exportWeChatDateBase(info WeChatInfo, expPath string, progress chan<- string) bool {

	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat DateBase start\", \"progress\": 1}"

	dbKey, err := hex.DecodeString(info.DBKey)
	if err != nil {
		log.Println("DecodeString:", err)
		progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%v\"}", err)
		return false
	}

	handleNumber := int64(0)
	fileNumber := getPathFileNumber(info.FilePath+"\\Msg", ".db")
	var wg sync.WaitGroup
	var reportWg sync.WaitGroup
	quitChan := make(chan struct{})
	taskChan := make(chan [2]string, 20)
	go func() {
		err = filepath.Walk(info.FilePath+"\\Msg", func(path string, finfo os.FileInfo, err error) error {
			if err != nil {
				log.Printf("filepath.Walk：%v\n", err)
				return err
			}
			if !finfo.IsDir() && strings.HasSuffix(path, ".db") {
				expFile := expPath + path[len(info.FilePath):]
				_, err := os.Stat(filepath.Dir(expFile))
				if err != nil {
					os.MkdirAll(filepath.Dir(expFile), 0644)
				}

				task := [2]string{path, expFile}
				taskChan <- task
			}

			return nil
		})
		if err != nil {
			log.Println("filepath.Walk:", err)
			progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%v\"}", err)
		}
		close(taskChan)
	}()

	for i := 1; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				if filepath.Base(task[0]) == "xInfo.db" {
					copyFile(task[0], task[1])
				} else {
					err = DecryptDataBase(task[0], dbKey, task[1])
					if err != nil {
						log.Println("DecryptDataBase:", err)
						progress <- fmt.Sprintf("{\"status\":\"error\", \"result\":\"%s %v\"}", task[0], err)
					}
				}
				atomic.AddInt64(&handleNumber, 1)
			}
		}()
	}

	reportWg.Add(1)
	go func() {
		defer reportWg.Done()
		for {
			select {
			case <-quitChan:
				log.Println("WeChat DateBase report progress end")
				return
			default:
				filePercent := float64(handleNumber) / float64(fileNumber)
				totalPercent := 1 + (filePercent * (20 - 1))
				totalPercentStr := fmt.Sprintf("{\"status\":\"processing\", \"result\":\"export WeChat DateBase doing\", \"progress\": %d}", int(totalPercent))
				progress <- totalPercentStr
				time.Sleep(time.Second)
			}
		}
	}()
	wg.Wait()
	close(quitChan)
	reportWg.Wait()
	progress <- "{\"status\":\"processing\", \"result\":\"export WeChat DateBase end\", \"progress\": 20}"
	return true
}

func GetWeChatInfo() (list *WeChatInfoList) {
	list = &WeChatInfoList{}
	list.Info = make([]WeChatInfo, 0)
	list.Total = 0

	processes, err := process.Processes()
	if err != nil {
		log.Println("Error getting processes:", err)
		return
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}
		info := WeChatInfo{}
		if name == "WeChat.exe" {
			info.ProcessID = uint32(p.Pid)
			info.Is64Bits, _ = Is64BitProcess(info.ProcessID)
			log.Println("ProcessID", info.ProcessID)
			files, err := p.OpenFiles()
			if err != nil {
				log.Println("OpenFiles failed")
				continue
			}

			for _, f := range files {
				if strings.HasSuffix(f.Path, "\\Media.db") {
					// fmt.Printf("opened %s\n", f.Path[4:])
					filePath := f.Path
					parts := strings.Split(filePath, string(filepath.Separator))
					if len(parts) < 4 {
						log.Println("Error filePath " + filePath)
						break
					}
					info.FilePath = strings.Join(parts[:len(parts)-2], string(filepath.Separator))
					info.AcountName = strings.Join(parts[len(parts)-3:len(parts)-2], string(filepath.Separator))
				}

			}

			if len(info.FilePath) == 0 {
				log.Println("wechat not log in")
				continue
			}

			hModuleSnap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, uint32(p.Pid))
			if err != nil {
				log.Println("CreateToolhelp32Snapshot failed", err)
				continue
			}
			defer windows.CloseHandle(hModuleSnap)

			var me32 windows.ModuleEntry32
			me32.Size = uint32(windows.SizeofModuleEntry32)

			err = windows.Module32First(hModuleSnap, &me32)
			if err != nil {
				log.Println("Module32First failed", err)
				continue
			}

			for ; err == nil; err = windows.Module32Next(hModuleSnap, &me32) {
				if windows.UTF16ToString(me32.Module[:]) == "WeChatWin.dll" {
					// fmt.Printf("MODULE NAME: %s\n", windows.UTF16ToString(me32.Module[:]))
					// fmt.Printf("executable NAME: %s\n", windows.UTF16ToString(me32.ExePath[:]))
					// fmt.Printf("base address: 0x%08X\n", me32.ModBaseAddr)
					// fmt.Printf("base ModBaseSize: %d\n", me32.ModBaseSize)
					info.DllBaseAddr = me32.ModBaseAddr
					info.DllBaseSize = me32.ModBaseSize

					var zero windows.Handle
					driverPath := windows.UTF16ToString(me32.ExePath[:])
					infoSize, err := windows.GetFileVersionInfoSize(driverPath, &zero)
					if err != nil {
						log.Println("GetFileVersionInfoSize failed", err)
						break
					}
					versionInfo := make([]byte, infoSize)
					if err = windows.GetFileVersionInfo(driverPath, 0, infoSize, unsafe.Pointer(&versionInfo[0])); err != nil {
						log.Println("GetFileVersionInfo failed", err)
						break
					}
					var fixedInfo *windows.VS_FIXEDFILEINFO
					fixedInfoLen := uint32(unsafe.Sizeof(*fixedInfo))
					err = windows.VerQueryValue(unsafe.Pointer(&versionInfo[0]), `\`, (unsafe.Pointer)(&fixedInfo), &fixedInfoLen)
					if err != nil {
						log.Println("VerQueryValue failed", err)
						break
					}
					// fmt.Printf("%s: v%d.%d.%d.%d\n", windows.UTF16ToString(me32.Module[:]),
					// 	(fixedInfo.FileVersionMS>>16)&0xff,
					// 	(fixedInfo.FileVersionMS>>0)&0xff,
					// 	(fixedInfo.FileVersionLS>>16)&0xff,
					// 	(fixedInfo.FileVersionLS>>0)&0xff)

					info.Version = fmt.Sprintf("%d.%d.%d.%d",
						(fixedInfo.FileVersionMS>>16)&0xff,
						(fixedInfo.FileVersionMS>>0)&0xff,
						(fixedInfo.FileVersionLS>>16)&0xff,
						(fixedInfo.FileVersionLS>>0)&0xff)
					list.Info = append(list.Info, info)
					list.Total += 1
					break
				}
			}
		}
	}
	return
}

func Is64BitProcess(pid uint32) (bool, error) {
	is64Bit := false
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		log.Println("Error opening process:", err)
		return is64Bit, errors.New("OpenProcess failed")
	}
	defer windows.CloseHandle(handle)

	err = windows.IsWow64Process(handle, &is64Bit)
	if err != nil {
		log.Println("Error IsWow64Process:", err)
	}
	return !is64Bit, err
}

func GetWeChatKey(info *WeChatInfo) string {
	mediaDB := info.FilePath + "\\Msg\\Media.db"
	if _, err := os.Stat(mediaDB); err != nil {
		log.Printf("open db %s error: %v", mediaDB, err)
		return ""
	}

	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, uint32(info.ProcessID))
	if err != nil {
		log.Println("Error opening process:", err)
		return ""
	}
	defer windows.CloseHandle(handle)

	buffer := make([]byte, info.DllBaseSize)
	err = windows.ReadProcessMemory(handle, uintptr(info.DllBaseAddr), &buffer[0], uintptr(len(buffer)), nil)
	if err != nil {
		log.Println("Error ReadProcessMemory:", err)
		return ""
	}

	offset := 0
	// searchStr := []byte(info.AcountName)
	for {
		index := hasDeviceSybmol(buffer[offset:])
		if index == -1 {
			log.Println("has not DeviceSybmol")
			break
		}
		fmt.Printf("hasDeviceSybmol: 0x%X\n", index)
		keys := findDBKeyPtr(buffer[offset:index], info.Is64Bits)
		// fmt.Println("keys:", keys)

		key, err := findDBkey(handle, info.FilePath+"\\Msg\\Media.db", keys)
		if err == nil {
			// fmt.Println("key:", key)
			return key
		}

		offset += (index + 20)
	}

	return ""
}

func hasDeviceSybmol(buffer []byte) int {
	sybmols := [...][]byte{
		{'a', 'n', 'd', 'r', 'o', 'i', 'd', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00},
		{'p', 'a', 'd', '-', 'a', 'n', 'd', 'r', 'o', 'i', 'd', 0x00, 0x00, 0x00, 0x00, 0x00, 0x0B, 0x00, 0x00, 0x00},
		{'i', 'p', 'h', 'o', 'n', 'e', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00},
		{'i', 'p', 'a', 'd', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00},
		{'O', 'H', 'O', 'S', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00},
	}
	for _, syb := range sybmols {
		if index := bytes.Index(buffer, syb); index != -1 {
			return index
		}
	}

	return -1
}

func findDBKeyPtr(buffer []byte, is64Bits bool) [][]byte {
	keys := make([][]byte, 0)
	step := 8
	keyLen := []byte{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if !is64Bits {
		keyLen = keyLen[:4]
		step = 4
	}

	offset := len(buffer) - step
	for {
		if bytes.Contains(buffer[offset:offset+step], keyLen) {
			keys = append(keys, buffer[offset-step:offset])
		}

		offset -= step
		if offset <= 0 {
			break
		}
	}

	return keys
}

func findDBkey(handle windows.Handle, path string, keys [][]byte) (string, error) {
	var keyAddrPtr uint64
	addrBuffer := make([]byte, 0x08)
	for _, key := range keys {
		copy(addrBuffer, key)
		err := binary.Read(bytes.NewReader(addrBuffer), binary.LittleEndian, &keyAddrPtr)
		if err != nil {
			log.Println("binary.Read:", err)
			continue
		}
		if keyAddrPtr == 0x00 {
			continue
		}
		log.Printf("keyAddrPtr: 0x%X\n", keyAddrPtr)
		keyBuffer := make([]byte, 0x20)
		err = windows.ReadProcessMemory(handle, uintptr(keyAddrPtr), &keyBuffer[0], uintptr(len(keyBuffer)), nil)
		if err != nil {
			// fmt.Println("Error ReadProcessMemory:", err)
			continue
		}
		if checkDataBaseKey(path, keyBuffer) {
			return hex.EncodeToString(keyBuffer), nil
		}
	}

	return "", errors.New("not found key")
}

func checkDataBaseKey(path string, password []byte) bool {
	fp, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fp.Close()

	fpReader := bufio.NewReaderSize(fp, defaultPageSize*100)

	buffer := make([]byte, defaultPageSize)

	n, err := fpReader.Read(buffer)
	if err != nil && n != defaultPageSize {
		log.Println("read failed:", err, n)
		return false
	}

	salt := buffer[:16]
	key := pbkdf2HMAC(password, salt, defaultIter, keySize)

	page1 := buffer[16:defaultPageSize]

	macSalt := xorBytes(salt, 0x3a)
	macKey := pbkdf2HMAC(key, macSalt, 2, keySize)

	hashMac := hmac.New(sha1.New, macKey)
	hashMac.Write(page1[:len(page1)-32])
	hashMac.Write([]byte{1, 0, 0, 0})

	return hmac.Equal(hashMac.Sum(nil), page1[len(page1)-32:len(page1)-12])
}

func (info WeChatInfo) String() string {
	return fmt.Sprintf("PID: %d\nVersion: v%s\nBaseAddr: 0x%08X\nDllSize: %d\nIs 64Bits: %v\nFilePath %s\nAcountName: %s",
		info.ProcessID, info.Version, info.DllBaseAddr, info.DllBaseSize, info.Is64Bits, info.FilePath, info.AcountName)
}

func copyFile(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		return bytesWritten, err
	}

	return bytesWritten, nil
}

func silkToMp3(amrBuf []byte, mp3Path string) error {
	amrReader := bytes.NewReader(amrBuf)

	var pcmBuffer bytes.Buffer
	sr := silk.NewWriter(&pcmBuffer)
	sr.Decoder.SetSampleRate(24000)
	amrReader.WriteTo(sr)
	sr.Close()

	if pcmBuffer.Len() == 0 {
		return errors.New("silk to mp3 failed " + mp3Path)
	}

	of, err := os.Create(mp3Path)
	if err != nil {
		return nil
	}
	defer of.Close()

	wr := lame.NewWriter(of)
	wr.Encoder.SetInSamplerate(24000)
	wr.Encoder.SetOutSamplerate(24000)
	wr.Encoder.SetNumChannels(1)
	wr.Encoder.SetQuality(5)
	// IMPORTANT!
	wr.Encoder.InitParams()

	pcmBuffer.WriteTo(wr)
	wr.Close()

	return nil
}

func getPathFileNumber(targetPath string, fileSuffix string) int64 {

	number := int64(0)
	err := filepath.Walk(targetPath, func(path string, finfo os.FileInfo, err error) error {
		if err != nil {
			log.Printf("filepath.Walk：%v\n", err)
			return err
		}
		if !finfo.IsDir() && strings.HasSuffix(path, fileSuffix) {
			number += 1
		}

		return nil
	})
	if err != nil {
		log.Println("filepath.Walk:", err)
	}

	return number
}

func ExportWeChatHeadImage(exportPath string) {
	progress := make(chan string)
	info := WeChatInfo{}

	miscDBPath := fmt.Sprintf("%s\\Msg\\Misc.db", exportPath)
	_, err := os.Stat(miscDBPath)
	if err != nil {
		log.Println("no exist:", miscDBPath)
		return
	}

	headImgPath := fmt.Sprintf("%s\\FileStorage\\HeadImage", exportPath)
	if _, err := os.Stat(headImgPath); err == nil {
		log.Println("has HeadImage")
		return
	}

	go func() {
		exportWeChatHeadImage(info, exportPath, progress)
		close(progress)
	}()

	for p := range progress {
		log.Println(p)
	}
	log.Println("ExportWeChatHeadImage done")
}
func CalculateSum(a, b int) int {
    return a + b
}
