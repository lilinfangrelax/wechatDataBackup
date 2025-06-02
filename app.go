package main // 定义包名为 main，表示这是一个可执行程序的包。
/**
这段代码是Go语言编写的，主要包含两个部分：错误处理和函数定义。我来逐步解释：

1. **错误处理部分**：
   - 首先检查`err`是否不为nil（即是否有错误发生）
   - 如果有错误，会：
     * 使用`log.Println`打印可执行文件路径错误日志
     * 返回包含错误信息的字符串（"Executable:" + 错误详情）
   - 如果没有错误，继续执行：
     * 构建目标可执行文件路径`exeDstPath`（在`exPath`后追加"\\wechatDataBackup.exe"）
     * 打印源路径到目标路径的复制信息
     * 调用`utils.CopyFile`复制文件
     * 如果复制出错，打印并返回错误信息
     * 最后返回空字符串表示成功
   - 注意：有一个重复的`return ""`语句，这是多余的，应该删除

2. **GetAppIsShareData函数**：
   - 这是一个方法，属于`App`结构体
   - 功能是检查应用程序是否共享数据
   - 逻辑：
     * 先检查`a.provider`是否存在（非nil）
     * 如果存在，返回`a.provider.IsShareData`的值
     * 如果不存在，返回`false`

这段代码主要展示了：
- 基本的错误处理模式
- 文件操作（复制）
- 结构体方法的定义
- 简单的条件检查逻辑

代码风格上使用了日志记录和明确的错误返回，是良好的实践。唯一需要改进的是删除那个重复的return语句。</think>



以下是对 [`app.go:980-982`](app.go:980) 代码的解释：

1.  **目的和功能**:
    这行代码 `return ""` 是一个重复的返回语句。在 [`app.go:978`](app.go:978) 已经有一个 `return ""` 语句，表示 [`ExportWeChatDataByUserName()`](app.go:920) 函数在成功执行所有操作后返回一个空字符串。因此，第 980 行的 `return ""` 是冗余的，没有任何实际的功能作用。它位于函数的末尾，但在它之前已经有一个 `return` 语句，这意味着代码执行永远不会到达第 980 行。

2.  **关键组件及其交互**:
    *   `return ""`: 这是 Go 语言中的一个返回语句，用于终止当前函数的执行并返回一个空字符串。
    *   `ExportWeChatDataByUserName()`: 这是包含此重复返回语句的函数。该函数旨在将微信数据导出到指定路径，并在成功时返回一个空字符串，在失败时返回错误信息。

3.  **重要的模式或技术**:
    *   **冗余代码/死代码**: 这行代码是一个典型的冗余代码示例，因为它永远不会被执行。通常，这种代码是由于重构不彻底或编程错误导致的。
    *   **函数返回值**: 函数设计为返回一个字符串，用于表示操作的结果（成功时为空字符串，失败时为错误信息）。

简而言之，[`app.go:980`](app.go:980) 的 `return ""` 语句是多余的，可以安全删除，因为它前面的代码已经处理了函数的返回逻辑。
**/
import (
	"context"      // 导入 context 包，用于管理请求的生命周期和取消信号。
	"encoding/json" // 导入 encoding/json 包，用于 JSON 数据的编码和解码。
	"fmt"          // 导入 fmt 包，用于格式化输入输出。
	"log"          // 导入 log 包，用于记录程序运行时的日志信息。
	"mime"         // 导入 mime 包，用于处理 MIME 类型。
	"net/http"     // 导入 net/http 包，提供了 HTTP 客户端和服务器的实现。
	"os"           // 导入 os 包，提供了与操作系统交互的函数，如文件操作、环境变量等。
	"path/filepath" // 导入 path/filepath 包，用于处理文件路径。
	"strconv"      // 导入 strconv 包，用于字符串和基本数据类型之间的转换。
	"strings"      // 导入 strings 包，用于字符串操作。
	"wechatDataBackup/pkg/utils" // 导入自定义的 utils 包，包含一些工具函数。
	"wechatDataBackup/pkg/wechat" // 导入自定义的 wechat 包，包含微信数据处理相关逻辑。

	"github.com/spf13/viper" // 导入 viper 包，用于处理应用程序配置。
	"github.com/wailsapp/wails/v2/pkg/runtime" // 导入 Wails 运行时包，用于与前端进行交互。
)

// 定义应用程序中使用的常量。
const (
	defaultConfig        = "config"           // 默认配置文件的名称。
	configDefaultUserKey = "userConfig.defaultUser" // 配置文件中默认用户键名。
	configUsersKey       = "userConfig.users"       // 配置文件中用户列表键名。
	configExportPathKey  = "exportPath"       // 配置文件中导出路径键名。
	appVersion           = "v1.2.4"           // 应用程序的版本号。
)

// FileLoader 结构体用于处理静态文件的加载和 HTTP 服务。
type FileLoader struct {
	http.Handler // 嵌入 http.Handler 接口，使其可以作为 HTTP 处理程序。
	FilePrefix string     // 文件路径前缀，用于构建完整的文件路径。
}

// NewFileLoader 函数创建一个新的 FileLoader 实例。
// prefix 参数指定文件加载器的文件路径前缀。
func NewFileLoader(prefix string) *FileLoader {
	mime.AddExtensionType(".mp3", "audio/mpeg") // 为 .mp3 文件添加 MIME 类型，确保浏览器能正确识别和播放。
	return &FileLoader{FilePrefix: prefix}      // 返回一个新的 FileLoader 实例，并设置文件前缀。
}

// SetFilePrefix 方法用于设置 FileLoader 的文件路径前缀。
// prefix 参数是新的文件路径前缀。
func (h *FileLoader) SetFilePrefix(prefix string) {
	h.FilePrefix = prefix                 // 更新文件前缀。
	log.Println("SetFilePrefix", h.FilePrefix) // 记录设置文件前缀的日志。
}

// ServeHTTP 方法实现了 http.Handler 接口，用于处理 HTTP 请求。
// 它根据请求的 Range 头处理文件分片下载。
func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// 构建请求的文件名，通过拼接文件前缀和去除 URL 路径前导斜杠后的部分。
	requestedFilename := h.FilePrefix + "\\" + strings.TrimPrefix(req.URL.Path, "/")

	// 尝试打开请求的文件。
	file, err := os.Open(requestedFilename)
	if err != nil {
		// 如果文件打开失败，返回 400 Bad Request 错误。
		http.Error(res, fmt.Sprintf("Could not load file %s", requestedFilename), http.StatusBadRequest)
		return
	}
	defer file.Close() // 确保在函数返回前关闭文件句柄。

	// 获取文件的统计信息。
	fileInfo, err := file.Stat()
	if err != nil {
		// 如果获取文件信息失败，返回 500 Internal Server Error 错误。
		http.Error(res, "Could not retrieve file info", http.StatusInternalServerError)
		return
	}

	fileSize := fileInfo.Size()           // 获取文件大小。
	rangeHeader := req.Header.Get("Range") // 获取请求头中的 Range 字段。
	if rangeHeader == "" {
		// 如果没有 Range 请求头，表示请求整个文件。
		res.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10)) // 设置 Content-Length 响应头。
		// 使用 http.ServeContent 直接返回整个文件内容。
		http.ServeContent(res, req, requestedFilename, fileInfo.ModTime(), file)
		return
	}

	var start, end int64
	// 检查 Range 请求头是否以 "bytes=" 开头。
	if strings.HasPrefix(rangeHeader, "bytes=") {
		// 解析 Range 请求头，获取请求的字节范围。
		ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		start, _ = strconv.ParseInt(ranges[0], 10, 64) // 解析起始字节。

		if len(ranges) > 1 && ranges[1] != "" {
			end, _ = strconv.ParseInt(ranges[1], 10, 64) // 解析结束字节。
		} else {
			end = fileSize - 1 // 如果结束字节为空，则表示到文件末尾。
		}
	} else {
		// 如果 Range 请求头格式无效，返回 416 Requested Range Not Satisfiable 错误。
		http.Error(res, "Invalid Range header", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// 检查请求的字节范围是否有效。
	if start < 0 || end >= fileSize || start > end {
		// 如果范围无效，返回 416 Requested Range Not Satisfiable 错误。
		http.Error(res, "Requested range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// 根据文件扩展名获取 MIME 类型。
	contentType := mime.TypeByExtension(filepath.Ext(requestedFilename))
	if contentType == "" {
		contentType = "application/octet-stream" // 如果无法确定 MIME 类型，则使用默认的二进制流类型。
	}
	res.Header().Set("Content-Type", contentType)                                     // 设置 Content-Type 响应头。
	res.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize)) // 设置 Content-Range 响应头。
	res.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))            // 设置 Content-Length 响应头。
	res.WriteHeader(http.StatusPartialContent)                                        // 设置 HTTP 状态码为 206 Partial Content。
	buffer := make([]byte, 102400)                                                    // 创建一个缓冲区用于读取文件内容。
	file.Seek(start, 0)                                                               // 将文件指针移动到请求的起始位置。
	for current := start; current <= end; {
		readSize := int64(len(buffer))
		if end-current+1 < readSize {
			readSize = end - current + 1 // 调整读取大小，确保不超过请求的范围。
		}

		n, err := file.Read(buffer[:readSize]) // 从文件中读取数据到缓冲区。
		if err != nil {
			break // 如果读取出错，则退出循环。
		}

		res.Write(buffer[:n]) // 将读取到的数据写入 HTTP 响应。
		current += int64(n)   // 更新当前读取位置。
	}
}

// App 结构体表示应用程序的主体，包含应用程序的上下文、数据提供者、用户信息等。
type App struct {
	ctx         context.Context             // 应用程序的上下文，用于管理生命周期和事件。
	infoList    *wechat.WeChatInfoList      // 微信信息列表，包含所有检测到的微信实例信息。
	provider    *wechat.WechatDataProvider  // 微信数据提供者，用于与微信数据库交互。
	defaultUser string                      // 默认用户账户名。
	users       []string                    // 已知用户账户名列表。
	firstStart  bool                        // 标记应用程序是否是第一次启动。
	firstInit   bool                        // 标记应用程序是否是第一次初始化。
	FLoader     *FileLoader                 // 文件加载器，用于处理静态文件服务。
}

// WeChatInfo 结构体定义了单个微信实例的详细信息。
type WeChatInfo struct {
	ProcessID  uint32 `json:"PID"`       // 微信进程的 ID。
	FilePath   string `json:"FilePath"`  // 微信可执行文件的路径。
	AcountName string `json:"AcountName"`// 微信账户名。
	Version    string `json:"Version"`   // 微信版本号。
	Is64Bits   bool   `json:"Is64Bits"`  // 微信进程是否是 64 位。
	DBKey      string `json:"DBkey"`     // 微信数据库的密钥。
}

// WeChatInfoList 结构体定义了微信信息列表，包含多个 WeChatInfo 实例。
type WeChatInfoList struct {
	Info  []WeChatInfo `json:"Info"`  // 微信信息切片。
	Total int          `json:"Total"` // 微信信息总数。
}

// WeChatAccountInfos 结构体定义了微信账户信息列表。
type WeChatAccountInfos struct {
	CurrentAccount string                     `json:"CurrentAccount"` // 当前活跃的微信账户名。
	Info           []wechat.WeChatAccountInfo `json:"Info"`           // 微信账户信息切片。
	Total          int                        `json:"Total"`          // 微信账户信息总数。
}

// ErrorMessage 结构体用于封装错误信息。
type ErrorMessage struct {
	ErrorStr string `json:"error"` // 错误字符串。
}

// NewApp 函数用于创建 App 应用程序结构体的实例。
func NewApp() *App {
	a := &App{}                                  // 创建 App 结构体的实例。
	log.Println("App version:", appVersion)          // 打印应用程序的版本号。
	a.firstInit = true                             // 设置 firstInit 标志为 true，表示首次初始化。
	a.FLoader = NewFileLoader(".\\")                // 创建 FileLoader 实例，用于加载文件。
	viper.SetConfigName(defaultConfig)               // 设置配置文件的名称为 "config"。
	viper.SetConfigType("json")                      // 设置配置文件的类型为 JSON。
	viper.AddConfigPath(".")                         // 添加当前目录作为配置文件的搜索路径。
	if err := viper.ReadInConfig(); err == nil {     // 尝试读取配置文件。
		a.defaultUser = viper.GetString(configDefaultUserKey) // 从配置文件中获取默认用户。
		a.users = viper.GetStringSlice(configUsersKey)       // 从配置文件中获取用户列表。
		prefix := viper.GetString(configExportPathKey)          // 从配置文件中获取导出路径前缀。
		if prefix != "" {                                      // 如果导出路径前缀不为空。
			log.Println("SetFilePrefix", prefix)                  // 打印设置的文件前缀。
			a.FLoader.SetFilePrefix(prefix)                      // 设置文件加载器的文件前缀。
		}
	} else {
		log.Println("not config exist") // 如果配置文件不存在，打印日志。
	}
	log.Printf("default: %s users: %v\n", a.defaultUser, a.users) // 打印默认用户和用户列表。
	if len(a.users) == 0 {                                       // 如果用户列表为空。
		a.firstStart = true // 设置 firstStart 标志为 true，表示应用程序是第一次启动。
	}

	return a // 返回 App 结构体的实例。
}

// startup 方法在应用程序启动时被调用。
// ctx 参数是应用程序的上下文，用于保存以便后续调用运行时方法。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx // 保存上下文。
}

// beforeClose 方法在应用程序关闭前被调用。
// 返回 true 可以阻止应用程序关闭，返回 false 则允许关闭。
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false // 默认不阻止应用程序关闭。
}

// shutdown 方法在应用程序关闭时被调用。
// 用于清理资源，例如关闭数据提供者。
func (a *App) shutdown(ctx context.Context) {
	if a.provider != nil {
		a.provider.WechatWechatDataProviderClose() // 关闭微信数据提供者。
		a.provider = nil                           // 将数据提供者设置为 nil。
	}
	log.Printf("App Version %s exit!", appVersion) // 打印应用程序退出信息。
}

// GetWeChatAllInfo 方法用于获取所有微信实例的信息。
// 返回一个 JSON 字符串，包含所有微信实例的详细信息。
func (a *App) GetWeChatAllInfo() string {
	infoList := WeChatInfoList{}           // 创建 WeChatInfoList 实例。
	infoList.Info = make([]WeChatInfo, 0) // 初始化 Info 切片。
	infoList.Total = 0                     // 初始化 Total 为 0。

	if a.provider != nil {
		a.provider.WechatWechatDataProviderClose() // 如果数据提供者已存在，则先关闭。
		a.provider = nil                           // 将数据提供者设置为 nil。
	}

	a.infoList = wechat.GetWeChatAllInfo() // 调用 wechat 包的 GetWeChatAllInfo 函数获取微信信息。
	for i := range a.infoList.Info {
		var info WeChatInfo // 创建 WeChatInfo 实例。
		// 复制信息到新的 WeChatInfo 实例。
		info.ProcessID = a.infoList.Info[i].ProcessID
		info.FilePath = a.infoList.Info[i].FilePath
		info.AcountName = a.infoList.Info[i].AcountName
		info.Version = a.infoList.Info[i].Version
		info.Is64Bits = a.infoList.Info[i].Is64Bits
		info.DBKey = a.infoList.Info[i].DBKey
		infoList.Info = append(infoList.Info, info) // 将信息添加到列表中。
		infoList.Total += 1                         // 总数加 1。
		log.Printf("ProcessID %d, FilePath %s, AcountName %s, Version %s, Is64Bits %t", info.ProcessID, info.FilePath, info.AcountName, info.Version, info.Is64Bits) // 打印详细信息。
	}
	infoStr, _ := json.Marshal(infoList) // 将信息列表转换为 JSON 字符串。
	// log.Println(string(infoStr))

	return string(infoStr) // 返回 JSON 字符串。
}

// ExportWeChatAllData 方法用于导出指定微信账户的所有数据。
// full 参数表示是否完全导出（包括消息），acountName 参数是需要导出的账户名。
func (a *App) ExportWeChatAllData(full bool, acountName string) {

	if a.provider != nil {
		a.provider.WechatWechatDataProviderClose() // 如果数据提供者已存在，则先关闭。
		a.provider = nil                           // 将数据提供者设置为 nil。
	}

	progress := make(chan string) // 创建一个字符串类型的通道，用于接收导出进度信息。
	go func() {                   // 在新的 Goroutine 中执行导出操作，避免阻塞主线程。
		var pInfo *wechat.WeChatInfo
		for i := range a.infoList.Info {
			if a.infoList.Info[i].AcountName == acountName {
				pInfo = &a.infoList.Info[i] // 找到匹配的账户信息。
				break
			}
		}

		if pInfo == nil {
			close(progress) // 关闭进度通道。
			// 发送错误事件到前端。
			runtime.EventsEmit(a.ctx, "exportData", fmt.Sprintf("{\"status\":\"error\", \"result\":\"%s error\"}", acountName))
			return
		}

		prefixExportPath := a.FLoader.FilePrefix + "\\User\\" // 构建导出路径前缀。
		_, err := os.Stat(prefixExportPath)                   // 检查导出路径是否存在。
		if err != nil {
			os.Mkdir(prefixExportPath, os.ModeDir) // 如果不存在，则创建目录。
		}

		expPath := prefixExportPath + pInfo.AcountName // 构建完整的导出路径。
		_, err = os.Stat(expPath)                      // 检查目标导出路径是否存在。
		if err == nil {
			if !full {
				os.RemoveAll(expPath + "\\Msg") // 如果不是完全导出，则只删除 Msg 目录。
			} else {
				os.RemoveAll(expPath) // 如果是完全导出，则删除整个目录。
			}
		}

		_, err = os.Stat(expPath)
		if err != nil {
			os.Mkdir(expPath, os.ModeDir) // 再次检查并创建目录，以防被删除后不存在。
		}

		go wechat.ExportWeChatAllData(*pInfo, expPath, progress) // 在新的 Goroutine 中开始导出微信数据。

		for p := range progress { // 循环接收进度通道中的信息。
			log.Println(p)                 // 打印进度信息到日志。
			runtime.EventsEmit(a.ctx, "exportData", p) // 发送进度事件到前端。
		}

		a.defaultUser = pInfo.AcountName // 设置当前导出的账户为默认用户。
		hasUser := false
		for _, user := range a.users {
			if user == pInfo.AcountName {
				hasUser = true // 检查用户是否已存在于用户列表中。
				break
			}
		}
		if !hasUser {
			a.users = append(a.users, pInfo.AcountName) // 如果用户不存在，则添加到用户列表。
		}
		a.setCurrentConfig() // 保存当前配置。
	}()
}

// createWechatDataProvider 方法用于创建微信数据提供者实例。
// resPath 参数是资源路径，prefix 参数是前缀。
func (a *App) createWechatDataProvider(resPath string, prefix string) error {
	// 如果数据提供者已存在且当前账户与请求的账户相同，则无需重新创建。
	if a.provider != nil && a.provider.SelfInfo != nil && filepath.Base(resPath) == a.provider.SelfInfo.UserName {
		log.Println("WechatDataProvider not need create:", a.provider.SelfInfo.UserName)
		return nil
	}

	if a.provider != nil {
		a.provider.WechatWechatDataProviderClose() // 如果数据提供者已存在，则先关闭。
		a.provider = nil                           // 将数据提供者设置为 nil。
		log.Println("createWechatDataProvider WechatWechatDataProviderClose") // 打印关闭日志。
	}

	provider, err := wechat.CreateWechatDataProvider(resPath, prefix) // 创建新的微信数据提供者。
	if err != nil {
		log.Println("CreateWechatDataProvider failed:", resPath) // 如果创建失败，打印错误日志。
		return err                                               // 返回错误。
	}

	a.provider = provider // 设置应用程序的数据提供者。
	// infoJson, _ := json.Marshal(a.provider.SelfInfo)
	// runtime.EventsEmit(a.ctx, "selfInfo", string(infoJson))
	return nil // 返回 nil 表示成功。
}

// WeChatInit 方法用于初始化微信相关功能。
func (a *App) WeChatInit() {

	if a.firstInit {
		a.firstInit = false                        // 将 firstInit 标志设置为 false。
		a.scanAccountByPath(a.FLoader.FilePrefix) // 扫描指定路径下的微信账户。
		log.Println("scanAccountByPath:", a.FLoader.FilePrefix) // 打印扫描路径。
	}

	if len(a.defaultUser) == 0 {
		log.Println("not defaultUser") // 如果没有默认用户，打印日志。
		return
	}

	expPath := a.FLoader.FilePrefix + "\\User\\" + a.defaultUser // 构建导出路径。
	prefixPath := "\\User\\" + a.defaultUser                     // 构建前缀路径。
	wechat.ExportWeChatHeadImage(expPath)                        // 导出微信头像。
	if a.createWechatDataProvider(expPath, prefixPath) == nil {  // 创建微信数据提供者。
		infoJson, _ := json.Marshal(a.provider.SelfInfo) // 将自身信息转换为 JSON 字符串。
		runtime.EventsEmit(a.ctx, "selfInfo", string(infoJson)) // 发送自身信息事件到前端。
	}
}

// GetWechatSessionList 方法用于获取微信会话列表。
// pageIndex 参数是页码，pageSize 参数是每页大小。
// 返回一个 JSON 字符串，包含会话列表信息。
func (a *App) GetWechatSessionList(pageIndex int, pageSize int) string {
	if a.provider == nil {
		log.Println("provider not init") // 如果数据提供者未初始化，打印日志。
		return "{\"Total\":0}"           // 返回空的总数。
	}
	log.Printf("pageIndex: %d\n", pageIndex) // 打印页码。
	list, err := a.provider.WeChatGetSessionList(pageIndex, pageSize) // 获取会话列表。
	if err != nil {
		return "{\"Total\":0}" // 如果获取失败，返回空的总数。
	}

	listStr, _ := json.Marshal(list) // 将列表转换为 JSON 字符串。
	log.Println("GetWechatSessionList:", list.Total) // 打印会话总数。
	return string(listStr)           // 返回 JSON 字符串。
}

// GetWechatContactList 方法用于获取微信联系人列表。
// pageIndex 参数是页码，pageSize 参数是每页大小。
// 返回一个 JSON 字符串，包含联系人列表信息。
func (a *App) GetWechatContactList(pageIndex int, pageSize int) string {
	if a.provider == nil {
		log.Println("provider not init") // 如果数据提供者未初始化，打印日志。
		return "{\"Total\":0}"           // 返回空的总数。
	}
	log.Printf("pageIndex: %d\n", pageIndex) // 打印页码。
	list, err := a.provider.WeChatGetContactList(pageIndex, pageSize) // 获取联系人列表。
	if err != nil {
		return "{\"Total\":0}" // 如果获取失败，返回空的总数。
	}

	listStr, _ := json.Marshal(list) // 将列表转换为 JSON 字符串。
	log.Println("WeChatGetContactList:", list.Total) // 打印联系人总数。
	return string(listStr)           // 返回 JSON 字符串。
}

// GetWechatMessageListByTime 方法用于根据时间获取微信消息列表。
// userName 参数是用户名，time 参数是时间戳，pageSize 参数是每页大小，direction 参数是搜索方向。
// 返回一个 JSON 字符串，包含消息列表信息。
func (a *App) GetWechatMessageListByTime(userName string, time int64, pageSize int, direction string) string {
	log.Println("GetWechatMessageListByTime:", userName, pageSize, time, direction) // 打印参数。
	if len(userName) == 0 {
		return "{\"Total\":0, \"Rows\":[]}" // 如果用户名为空，返回空列表。
	}
	dire := wechat.Message_Search_Forward // 默认搜索方向为向前。
	if direction == "backward" {
		dire = wechat.Message_Search_Backward // 如果方向为 "backward"，则设置为向后。
	} else if direction == "both" {
		dire = wechat.Message_Search_Both // 如果方向为 "both"，则设置为双向。
	}
	list, err := a.provider.WeChatGetMessageListByTime(userName, time, pageSize, dire) // 获取消息列表。
	if err != nil {
		log.Println("GetWechatMessageListByTime failed:", err) // 如果获取失败，打印错误日志。
		return ""                                              // 返回空字符串。
	}
	listStr, _ := json.Marshal(list) // 将列表转换为 JSON 字符串。
	log.Println("GetWechatMessageListByTime:", list.Total) // 打印消息总数。

	return string(listStr) // 返回 JSON 字符串。
}

// GetWechatMessageListByType 方法用于根据消息类型获取微信消息列表。
// userName 参数是用户名，time 参数是时间戳，pageSize 参数是每页大小，msgType 参数是消息类型，direction 参数是搜索方向。
// 返回一个 JSON 字符串，包含消息列表信息。
func (a *App) GetWechatMessageListByType(userName string, time int64, pageSize int, msgType string, direction string) string {
	log.Println("GetWechatMessageListByType:", userName, pageSize, time, msgType, direction) // 打印参数。
	if len(userName) == 0 {
		return "{\"Total\":0, \"Rows\":[]}" // 如果用户名为空，返回空列表。
	}
	dire := wechat.Message_Search_Forward // 默认搜索方向为向前。
	if direction == "backward" {
		dire = wechat.Message_Search_Backward // 如果方向为 "backward"，则设置为向后。
	} else if direction == "both" {
		dire = wechat.Message_Search_Both // 如果方向为 "both"，则设置为双向。
	}
	list, err := a.provider.WeChatGetMessageListByType(userName, time, pageSize, msgType, dire) // 获取消息列表。
	if err != nil {
		log.Println("WeChatGetMessageListByType failed:", err) // 如果获取失败，打印错误日志。
		return ""                                              // 返回空字符串。
	}
	listStr, _ := json.Marshal(list) // 将列表转换为 JSON 字符串。
	log.Println("WeChatGetMessageListByType:", list.Total) // 打印消息总数。

	return string(listStr) // 返回 JSON 字符串。
}

// GetWechatMessageListByKeyWord 方法用于根据关键词获取微信消息列表。
// userName 参数是用户名，time 参数是时间戳，keyword 参数是关键词，msgType 参数是消息类型，pageSize 参数是每页大小。
// 返回一个 JSON 字符串，包含消息列表信息。
func (a *App) GetWechatMessageListByKeyWord(userName string, time int64, keyword string, msgType string, pageSize int) string {
	log.Println("GetWechatMessageListByKeyWord:", userName, pageSize, time, msgType) // 打印参数。
	if len(userName) == 0 {
		return "{\"Total\":0, \"Rows\":[]}" // 如果用户名为空，返回空列表。
	}
	list, err := a.provider.WeChatGetMessageListByKeyWord(userName, time, keyword, msgType, pageSize) // 获取消息列表。
	if err != nil {
		log.Println("WeChatGetMessageListByKeyWord failed:", err) // 如果获取失败，打印错误日志。
		return ""                                                  // 返回空字符串。
	}
	listStr, _ := json.Marshal(list) // 将列表转换为 JSON 字符串。
	log.Println("WeChatGetMessageListByKeyWord:", list.Total, list.KeyWord) // 打印消息总数和关键词。

	return string(listStr) // 返回 JSON 字符串。
}

// GetWechatMessageDate 方法用于获取微信消息的日期列表。
// userName 参数是用户名。
// 返回一个 JSON 字符串，包含消息日期信息。
func (a *App) GetWechatMessageDate(userName string) string {
	log.Println("GetWechatMessageDate:", userName) // 打印用户名。
	if len(userName) == 0 {
		return "{\"Total\":0, \"Date\":[]}" // 如果用户名为空，返回空列表。
	}

	messageData, err := a.provider.WeChatGetMessageDate(userName) // 获取消息日期数据。
	if err != nil {
		log.Println("GetWechatMessageDate:", err) // 如果获取失败，打印错误日志。
		return ""                                 // 返回空字符串。
	}

	messageDataStr, _ := json.Marshal(messageData) // 将消息日期数据转换为 JSON 字符串。
	log.Println("GetWechatMessageDate:", messageData.Total) // 打印消息日期总数。

	return string(messageDataStr) // 返回 JSON 字符串。
}

// setCurrentConfig 方法用于保存当前配置到配置文件。
func (a *App) setCurrentConfig() {
	viper.Set(configDefaultUserKey, a.defaultUser)     // 设置默认用户。
	viper.Set(configUsersKey, a.users)                 // 设置用户列表。
	viper.Set(configExportPathKey, a.FLoader.FilePrefix) // 设置导出路径。
	err := viper.SafeWriteConfig()                     // 尝试安全写入配置文件。
	if err != nil {
		log.Println(err) // 如果安全写入失败，打印错误日志。
		err = viper.WriteConfig() // 尝试直接写入配置文件。
		if err != nil {
			log.Println(err) // 如果直接写入也失败，打印错误日志。
		}
	}
}

// userList 结构体用于封装用户列表。
type userList struct {
	Users []string `json:"Users"` // 用户名切片。
}

// GetWeChatUserList 函数用于获取微信用户列表。
// 返回一个 JSON 字符串，包含用户列表信息。
func (a *App) GetWeChatUserList() string {
	l := userList{}     // 创建 userList 实例。
	l.Users = a.users // 设置用户列表。

	usersStr, _ := json.Marshal(l) // 将用户列表转换为 JSON 字符串。
	str := string(usersStr)        // 转换为字符串。
	log.Println("users:", str)     // 打印用户列表。
	return str                     // 返回 JSON 字符串。
}

// OpenFileOrExplorer 函数用于打开文件或资源管理器。
// filePath 参数是文件路径，explorer 参数表示是否打开资源管理器（true）或文件（false）。
// 返回一个 JSON 字符串，表示操作结果。
func (a *App) OpenFileOrExplorer(filePath string, explorer bool) string {
	// if root, err := os.Getwd(); err == nil {
	// 	filePath = root + filePath[1:]
	// }
	// log.Println("OpenFileOrExplorer:", filePath)

	path := a.FLoader.FilePrefix + filePath // 构建完整的文件路径。
	err := utils.OpenFileOrExplorer(path, explorer) // 调用工具函数打开文件或资源管理器。
	if err != nil {
		return "{\"result\": \"OpenFileOrExplorer failed\", \"status\":\"failed\"}" // 如果操作失败，返回失败信息。
	}

	return fmt.Sprintf("{\"result\": \"%s\", \"status\":\"OK\"}", "") // 返回成功信息。
}

// GetWeChatRoomUserList 函数用于获取微信聊天室的用户列表。
// roomId 参数是聊天室 ID。
// 返回一个 JSON 字符串，包含聊天室用户列表信息。
func (a *App) GetWeChatRoomUserList(roomId string) string {
	userlist, err := a.provider.WeChatGetChatRoomUserList(roomId) // 获取聊天室的用户列表。
	if err != nil {                                                  // 如果获取失败。
		log.Println("WeChatGetChatRoomUserList:", err)                  // 打印错误日志。
		return ""                                                     // 返回空字符串。
	}

	userListStr, _ := json.Marshal(userlist) // 将用户列表转换为 JSON 字符串。

	return string(userListStr) // 返回 JSON 字符串。
}

// GetAppVersion 函数用于获取应用程序的版本号。
// 返回应用程序的版本号字符串。
func (a *App) GetAppVersion() string {
	return appVersion // 返回应用程序的版本号。
}

// GetAppIsFirstStart 函数用于获取应用程序是否是第一次启动。
// 返回一个布尔值，true 表示第一次启动，false 表示不是。
func (a *App) GetAppIsFirstStart() bool {
	defer func() { a.firstStart = false }() // 使用 defer 确保在函数结束时将 firstStart 标志设置为 false。
	return a.firstStart                     // 返回 firstStart 标志。
}

// GetWechatLocalAccountInfo 函数用于获取本地微信账户信息。
// 返回一个 JSON 字符串，包含本地微信账户信息列表。
func (a *App) GetWechatLocalAccountInfo() string {
	infos := WeChatAccountInfos{}                                  // 创建 WeChatAccountInfos 实例。
	infos.Info = make([]wechat.WeChatAccountInfo, 0)                  // 初始化微信账户信息切片。
	infos.Total = 0                                                    // 初始化微信账户信息总数为 0。
	infos.CurrentAccount = a.defaultUser                               // 设置当前账户。
	for i := range a.users {                                           // 遍历用户列表。
		resPath := a.FLoader.FilePrefix + "\\User\\" + a.users[i]       // 构建资源路径。
		if _, err := os.Stat(resPath); err != nil {                     // 检查资源路径是否存在。
			log.Println("GetWechatLocalAccountInfo:", resPath, err)          // 打印错误日志。
			continue                                                     // 继续下一个循环。
		}

		prefixResPath := "\\User\\" + a.users[i]
		info, err := wechat.WechatGetAccountInfo(resPath, prefixResPath, a.users[i]) // 获取微信账户信息。
		if err != nil {
			log.Println("GetWechatLocalAccountInfo", err) // 如果获取失败，打印错误日志。
			continue                                     // 继续下一个循环。
		}

		infos.Info = append(infos.Info, *info) // 将账户信息添加到列表中。
		infos.Total += 1                       // 总数加 1。
	}

	infoString, _ := json.Marshal(infos) // 将信息转换为 JSON 字符串。
	log.Println(string(infoString))      // 打印信息。

	return string(infoString) // 返回 JSON 字符串。
}

// WechatSwitchAccount 函数用于切换微信账户。
// account 参数是目标账户名。
// 返回一个布尔值，true 表示切换成功，false 表示切换失败。
func (a *App) WechatSwitchAccount(account string) bool {
	for i := range a.users {                               // 遍历用户列表。
		if a.users[i] == account {                         // 如果找到匹配的账户。
			if a.provider != nil {                         // 如果数据提供者存在。
				a.provider.WechatWechatDataProviderClose() // 关闭数据提供者。
				a.provider = nil                             // 将数据提供者设置为 nil。
			}
			a.defaultUser = account // 设置默认用户。
			a.setCurrentConfig()    // 保存配置。
			return true             // 返回 true，表示切换成功。
		}
	}

	return false // 返回 false，表示切换失败。
}

// GetExportPathStat 函数用于获取导出路径的状态信息。
// 返回一个 JSON 字符串，包含导出路径的状态信息。
func (a *App) GetExportPathStat() string {
	path := a.FLoader.FilePrefix                                  // 获取导出路径。
	log.Println("utils.GetPathStat ++")                             // 打印日志，表示开始获取路径状态。
	stat, err := utils.GetPathStat(path)                            // 获取路径状态。
	log.Println("utils.GetPathStat --")                             // 打印日志，表示结束获取路径状态。
	if err != nil {                                                  // 如果获取失败。
		log.Println("GetPathStat error:", path, err)                  // 打印错误日志。
		var msg ErrorMessage                                          // 创建 ErrorMessage 实例。
		msg.ErrorStr = fmt.Sprintf("%s:%v", path, err)                // 设置错误信息。
		msgStr, _ := json.Marshal(msg)                                // 将错误信息转换为 JSON 字符串。
		return string(msgStr)                                          // 返回 JSON 字符串。
	}

	statString, _ := json.Marshal(stat) // 将路径状态转换为 JSON 字符串。

	return string(statString) // 返回 JSON 字符串。
}

// ExportPathIsCanWrite 函数用于判断导出路径是否可写。
// 返回一个布尔值，true 表示可写，false 表示不可写。
func (a *App) ExportPathIsCanWrite() bool {
	path := a.FLoader.FilePrefix                                  // 获取导出路径。
	return utils.PathIsCanWriteFile(path)                            // 返回导出路径是否可写。
}

// OpenExportPath 函数用于在浏览器中打开导出路径。
func (a *App) OpenExportPath() {
	path := a.FLoader.FilePrefix                                  // 获取导出路径。
	runtime.BrowserOpenURL(a.ctx, path)                            // 在浏览器中打开导出路径。
}

// OpenDirectoryDialog 函数用于打开选择目录的对话框。
// 返回用户选择的目录路径字符串。
func (a *App) OpenDirectoryDialog() string {
	dialogOptions := runtime.OpenDialogOptions{                                  // 创建 OpenDialogOptions 实例。
		Title: "选择导出路径",                                                    // 设置对话框的标题。
	}
	selectedDir, err := runtime.OpenDirectoryDialog(a.ctx, dialogOptions) // 打开选择目录的对话框。
	if err != nil {                                                          // 如果选择失败。
		log.Println("OpenDirectoryDialog:", err)                              // 打印错误日志。
		return ""                                                             // 返回空字符串。
	}

	if selectedDir == "" {
		log.Println("Cancel selectedDir") // 如果用户取消选择，打印日志。
		return ""
	}

	if selectedDir == a.FLoader.FilePrefix {
		log.Println("same path No need SetFilePrefix") // 如果选择的路径与当前路径相同，打印日志。
		return ""
	}

	if !utils.PathIsCanWriteFile(selectedDir) {
		log.Println("PathIsCanWriteFile:", selectedDir, "error") // 如果路径不可写，打印错误日志。
		return ""
	}

	a.FLoader.SetFilePrefix(selectedDir) // 设置文件加载器的文件前缀。
	log.Println("OpenDirectoryDialog:", selectedDir) // 打印选择的目录。
	a.scanAccountByPath(selectedDir)     // 扫描新路径下的账户。
	return selectedDir                   // 返回选择的目录。
}

// scanAccountByPath 函数用于扫描指定路径下的微信账户。
// path 参数是需要扫描的路径。
// 返回一个 error，如果扫描过程中发生错误。
func (a *App) scanAccountByPath(path string) error {
	infos := WeChatAccountInfos{}                                  // 创建 WeChatAccountInfos 实例。
	infos.Info = make([]wechat.WeChatAccountInfo, 0)                  // 初始化微信账户信息切片。
	infos.Total = 0                                                    // 初始化微信账户信息总数为 0。
	infos.CurrentAccount = ""                                          // 初始化当前账户为空字符串。

	userPath := path + "\\User\\" // 构建用户数据路径。
	if _, err := os.Stat(userPath); err != nil {
		return err // 如果用户路径不存在，返回错误。
	}

	dirs, err := os.ReadDir(userPath) // 读取用户路径下的目录。
	if err != nil {
		log.Println("ReadDir", err) // 如果读取目录失败，打印错误日志。
		return err                  // 返回错误。
	}

	for i := range dirs {
		if !dirs[i].Type().IsDir() {
			continue // 如果不是目录，则跳过。
		}
		log.Println("dirs[i].Name():", dirs[i].Name()) // 打印目录名。
		resPath := path + "\\User\\" + dirs[i].Name()  // 构建资源路径。
		prefixResPath := "\\User\\" + dirs[i].Name()
		info, err := wechat.WechatGetAccountInfo(resPath, prefixResPath, dirs[i].Name()) // 获取微信账户信息。
		if err != nil {
			log.Println("GetWechatLocalAccountInfo", err) // 如果获取失败，打印错误日志。
			continue                                     // 继续下一个循环。
		}

		infos.Info = append(infos.Info, *info) // 将账户信息添加到列表中。
		infos.Total += 1                       // 总数加 1。
	}

	users := make([]string, 0)
	for i := 0; i < infos.Total; i++ {
		users = append(users, infos.Info[i].AccountName) // 提取账户名到用户列表。
	}

	a.users = users // 更新应用程序的用户列表。
	found := false
	for i := range a.users {
		if a.defaultUser == a.users[i] {
			found = true // 检查默认用户是否在新的用户列表中。
		}
	}

	if !found {
		a.defaultUser = "" // 如果默认用户不在列表中，则清空默认用户。
	}
	if a.defaultUser == "" && len(a.users) > 0 {
		a.defaultUser = a.users[0] // 如果没有默认用户且有其他用户，则设置第一个用户为默认用户。
	}

	if len(a.users) > 0 {
		a.setCurrentConfig() // 如果有用户，保存当前配置。
	}

	return nil // 返回 nil 表示成功。
}

// OepnLogFileExplorer 函数用于打开日志文件所在目录。
func (a *App) OepnLogFileExplorer() {
	utils.OpenFileOrExplorer(".\\app.log", true) // 打开日志文件所在目录。
}

// SaveFileDialog 函数用于打开保存文件对话框。
// file 参数是源文件路径，alisa 参数是默认文件名。
// 返回保存的文件路径字符串，如果保存失败则返回错误信息。
func (a *App) SaveFileDialog(file string, alisa string) string {
	filePath := a.FLoader.FilePrefix + file                                  // 构建完整的文件路径。
	if _, err := os.Stat(filePath); err != nil {                     // 检查文件是否存在。
		log.Println("SaveFileDialog:", err)                              // 打印错误日志。
		return err.Error()                                                 // 返回错误信息。
	}

	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{ // 打开保存文件对话框。
		DefaultFilename: alisa, // 设置默认文件名。
		Title:           "选择保存路径", // 设置对话框标题。
	})
	if err != nil {
		log.Println("SaveFileDialog:", err) // 如果保存失败，打印错误日志。
		return err.Error()                  // 返回错误信息。
	}

	if savePath == "" {
		return "" // 如果用户取消保存，返回空字符串。
	}

	dirPath := filepath.Dir(savePath) // 获取保存路径的目录。
	if !utils.PathIsCanWriteFile(dirPath) {
		errStr := "Path Is Can't Write File: " + filepath.Dir(savePath) // 构建错误信息。
		log.Println(errStr)                                            // 打印错误日志。
		return errStr                                                  // 返回错误信息。
	}

	_, err = utils.CopyFile(filePath, savePath) // 复制文件。
	if err != nil {
		log.Println("Error CopyFile", filePath, savePath, err) // 如果复制失败，打印错误日志。
		return err.Error()                                     // 返回错误信息。
	}

	return "" // 返回空字符串表示成功。
}

// GetSessionLastTime 函数用于获取会话的最后时间。
// userName 参数是用户名。
// 返回一个 JSON 字符串，包含会话的最后时间信息。
func (a *App) GetSessionLastTime(userName string) string {
	if a.provider == nil || userName == "" {                   // 如果数据提供者未初始化或用户名为空。
		lastTime := &wechat.WeChatLastTime{}                      // 创建 WeChatLastTime 实例。
		lastTimeString, _ := json.Marshal(lastTime)                // 将 lastTime 转换为 JSON 字符串。
		return string(lastTimeString)                               // 返回 JSON 字符串。
	}

	lastTime := a.provider.WeChatGetSessionLastTime(userName) // 获取会话的最后时间。

	lastTimeString, _ := json.Marshal(lastTime) // 将 lastTime 转换为 JSON 字符串。

	return string(lastTimeString) // 返回 JSON 字符串。
}

// SetSessionLastTime 函数用于设置会话的最后时间。
// userName 参数是用户名，stamp 参数是时间戳，messageId 参数是消息 ID。
// 返回一个字符串，如果设置失败则返回错误信息。
func (a *App) SetSessionLastTime(userName string, stamp int64, messageId string) string {
	if a.provider == nil { // 如果数据提供者未初始化。
		return "" // 返回空字符串。
	}

	lastTime := &wechat.WeChatLastTime{ // 创建 WeChatLastTime 实例。
		UserName:  userName,  // 设置用户名。
		Timestamp: stamp,     // 设置时间戳。
		MessageId: messageId, // 设置消息 ID。
	}
	err := a.provider.WeChatSetSessionLastTime(lastTime) // 设置会话的最后时间。
	if err != nil {
		log.Println("WeChatSetSessionLastTime failed:", err.Error()) // 如果设置失败，打印错误日志。
		return err.Error()                                           // 返回错误信息。
	}

	return "" // 返回空字符串表示成功。
}

// SetSessionBookMask 函数用于设置会话的书签掩码。
// userName 参数是用户名，tag 参数是标签，info 参数是信息。
// 返回一个字符串，如果设置失败则返回错误信息。
func (a *App) SetSessionBookMask(userName, tag, info string) string {
	if a.provider == nil || userName == "" { // 如果数据提供者未初始化或用户名为空。
		return "invaild params" // 返回 "invaild params"。
	}
	err := a.provider.WeChatSetSessionBookMask(userName, tag, info) // 设置会话的书签掩码。
	if err != nil {
		log.Println("WeChatSetSessionBookMask failed:", err.Error()) // 如果设置失败，打印错误日志。
		return err.Error()                                           // 返回错误信息。
	}

	return "" // 返回空字符串表示成功。
}

// DelSessionBookMask 函数用于删除会话的书签掩码。
// markId 参数是书签 ID。
// 返回一个字符串，如果删除失败则返回错误信息。
func (a *App) DelSessionBookMask(markId string) string {
	if a.provider == nil || markId == "" { // 如果数据提供者未初始化或 markId 为空。
		return "invaild params" // 返回 "invaild params"。
	}

	err := a.provider.WeChatDelSessionBookMask(markId) // 删除会话的书签掩码。
	if err != nil {
		log.Println("WeChatDelSessionBookMask failed:", err.Error()) // 如果删除失败，打印错误日志。
		return err.Error()                                           // 返回错误信息。
	}

	return "" // 返回空字符串表示成功。
}

// GetSessionBookMaskList 函数用于获取会话的书签掩码列表。
// userName 参数是用户名。
// 返回一个 JSON 字符串，包含会话的书签掩码列表信息。
func (a *App) GetSessionBookMaskList(userName string) string {
	if a.provider == nil || userName == "" { // 如果数据提供者未初始化或用户名为空。
		return "invaild params" // 返回 "invaild params"。
	}
	markLIst, err := a.provider.WeChatGetSessionBookMaskList(userName) // 获取会话的书签掩码列表。
	if err != nil {
		log.Println("WeChatGetSessionBookMaskList failed:", err.Error()) // 如果获取失败，打印错误日志。
		_list := &wechat.WeChatBookMarkList{}                            // 创建空的 WeChatBookMarkList 实例。
		_listString, _ := json.Marshal(_list)                            // 将空列表转换为 JSON 字符串。
		return string(_listString)                                       // 返回 JSON 字符串。
	}

	markLIstString, _ := json.Marshal(markLIst) // 将列表转换为 JSON 字符串。
	return string(markLIstString)                // 返回 JSON 字符串。
}

// SelectedDirDialog 函数用于打开选择目录的对话框。
// title 参数是对话框的标题。
// 返回用户选择的目录路径字符串。
func (a *App) SelectedDirDialog(title string) string {
	dialogOptions := runtime.OpenDialogOptions{ // 创建 OpenDialogOptions 实例。
		Title: title, // 设置对话框的标题。
	}
	selectedDir, err := runtime.OpenDirectoryDialog(a.ctx, dialogOptions) // 打开选择目录的对话框。
	if err != nil {                                                          // 如果选择失败。
		log.Println("OpenDirectoryDialog:", err)                              // 打印错误日志。
		return ""                                                             // 返回空字符串。
	}

	if selectedDir == "" {
		return "" // 如果用户取消选择，返回空字符串。
	}

	return selectedDir // 返回选择的目录。
}

// ExportWeChatDataByUserName 函数用于按用户名导出微信数据。
// userName 参数是用户名，path 参数是导出路径。
// 返回一个字符串，如果导出失败则返回错误信息。
func (a *App) ExportWeChatDataByUserName(userName, path string) string {
	if a.provider == nil || userName == "" || path == "" { // 如果数据提供者未初始化或用户名或路径为空。
		return "invaild params" + userName // 返回 "invaild params" 加上用户名。
	}

	if !utils.PathIsCanWriteFile(path) {
		log.Println("PathIsCanWriteFile: " + path) // 如果路径不可写，打印错误日志。
		return "PathIsCanWriteFile: " + path       // 返回错误信息。
	}

	exPath := path + "\\" + "wechatDataBackup_" + userName // 构建导出目录路径。
	if _, err := os.Stat(exPath); err != nil {
		os.MkdirAll(exPath, os.ModePerm) // 如果目录不存在，则创建所有必要的目录。
	} else {
		return "path exist:" + exPath // 如果目录已存在，返回错误信息。
	}

	log.Println("ExportWeChatDataByUserName:", userName, exPath) // 打印导出信息。
	err := a.provider.WeChatExportDataByUserName(userName, exPath) // 导出微信数据。
	if err != nil {
		log.Println("WeChatExportDataByUserName failed:", err) // 如果导出失败，打印错误日志。
		return "WeChatExportDataByUserName failed:" + err.Error() // 返回错误信息。
	}

	config := map[string]interface{}{ // 构建配置映射。
		"exportpath": ".\\",
		"userconfig": map[string]interface{}{
			"defaultuser": a.defaultUser,
			"users":       []string{a.defaultUser},
		},
	}

	configJson, err := json.MarshalIndent(config, "", "	") // 将配置映射转换为带缩进的 JSON 字符串。
	if err != nil {
		log.Println("MarshalIndent:", err) // 如果转换失败，打印错误日志。
		return "MarshalIndent:" + err.Error() // 返回错误信息。
	}

	configPath := exPath + "\\" + "config.json" // 构建配置文件路径。
	err = os.WriteFile(configPath, configJson, os.ModePerm) // 写入配置文件。
	if err != nil {
		log.Println("WriteFile:", err) // 如果写入失败，打印错误日志。
		return "WriteFile:" + err.Error() // 返回错误信息。
	}

	exeSrcPath, err := os.Executable() // 获取当前可执行文件的路径。
	if err != nil {
		log.Println("Executable:", exeSrcPath) // 如果获取失败，打印错误日志。
		return "Executable:" + err.Error()     // 返回错误信息。
	}

	exeDstPath := exPath + "\\" + "wechatDataBackup.exe" // 构建目标可执行文件路径。
	log.Printf("Copy [%s] -> [%s]\n", exeSrcPath, exeDstPath) // 打印复制信息。
	_, err = utils.CopyFile(exeSrcPath, exeDstPath) // 复制可执行文件。
	if err != nil {
		log.Println("CopyFile:", err) // 如果复制失败，打印错误日志。
		return "CopyFile:" + err.Error() // 返回错误信息。
	}
	return "" // 返回空字符串表示成功。

	return "" // 重复的返回语句，可以删除。
}

// GetAppIsShareData 函数用于获取应用程序是否共享数据。
// 返回一个布尔值，true 表示共享数据，false 表示不共享。
func (a *App) GetAppIsShareData() bool {
	if a.provider != nil { // 如果数据提供者存在。
		return a.provider.IsShareData // 返回数据提供者是否共享数据。
	}
	return false // 否则，返回 false。
}
