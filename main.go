package main // 定义包名为 main，表示这是一个可执行程序的包声明。

import (
	"embed" // 导入 embed 包，用于将静态资源（如 HTML、CSS、JavaScript 文件）嵌入到可执行文件中。
	"io"    // 导入 io 包，提供了基本的 I/O 接口，用于读取和写入数据。
	"log"   // 导入 log 包，用于记录程序运行时的日志信息。
	"os"    // 导入 os 包，提供了与操作系统交互的函数，如文件操作、环境变量等。

	"github.com/wailsapp/wails/v2" // 导入 Wails 框架的核心包，用于构建跨平台的桌面应用程序。
	"github.com/wailsapp/wails/v2/pkg/options" // 导入 Wails 应用程序选项包，用于配置应用程序的各种属性。
	"github.com/wailsapp/wails/v2/pkg/options/assetserver" // 导入 Wails 资产服务器选项包，用于配置静态资源服务器。
	"gopkg.in/natefinch/lumberjack.v2" // 导入 lumberjack 包，用于日志文件轮转，方便日志管理。
)

//go:embed all:frontend/dist // 嵌入前端构建后的所有静态文件。
// assets 变量用于存储嵌入的静态资源，这些资源在编译时会被嵌入到 Go 程序中。
var assets embed.FS

// init 函数在 main 函数之前执行，用于初始化操作。
func init() {
	// 设置日志输出格式，包括日期、微秒和文件名行号，方便调试和问题追踪。
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

// main 函数是程序的入口点，程序从这里开始执行。
func main() {
	// 创建一个 lumberjack.Logger 实例，用于配置日志文件轮转，实现日志文件的自动管理。
	logJack := &lumberjack.Logger{
		Filename:   "./app.log", // 指定日志文件的路径。
		MaxSize:    5,           // 设置每个日志文件的最大大小为 5MB。
		MaxBackups: 1,           // 设置保留的旧日志文件的最大数量为 1。
		MaxAge:     30,          // 设置保留旧日志文件的最大天数为 30 天。
		Compress:   false,       // 设置不压缩旧的日志文件。
	}
	defer logJack.Close() // 使用 defer 确保在 main 函数结束时关闭日志文件，释放资源。

	// 创建一个多重写入器，将日志同时输出到文件和标准输出（控制台），方便查看和管理日志。
	multiWriter := io.MultiWriter(logJack, os.Stdout)
	// 设置日志输出目标为多重写入器，将日志同时写入文件和标准输出。
	log.SetOutput(multiWriter)
	// 打印启动信息到日志，用于标识应用程序的启动。
	log.Println("====================== wechatDataBackup ======================")
	// 创建应用程序结构体的实例，用于管理应用程序的逻辑和状态。
	app := NewApp()

	// 使用 Wails 框架运行应用程序，并配置各种选项，启动应用程序的主循环。
	err := wails.Run(&options.App{
		Title:     "wechatDataBackup", // 设置应用程序窗口的标题。
		MinWidth:  800,                // 设置窗口的最小宽度。
		MinHeight: 600,                // 设置窗口的最小高度。
		Width:     1024,               // 设置窗口的默认宽度。
		Height:    768,                // 设置窗口的默认高度。
		AssetServer: &assetserver.Options{ // 配置资产服务器，用于提供前端静态资源。
			Assets:  assets,    // 指定嵌入的静态资源。
			Handler: app.FLoader, // 指定自定义的文件加载器。
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1}, // 设置应用程序的背景颜色。
		OnStartup:        app.startup,    // 指定应用程序启动时要调用的函数。
		OnBeforeClose:    app.beforeClose, // 指定应用程序关闭前要调用的函数。
		OnShutdown:       app.shutdown,   // 指定应用程序关闭时要调用的函数。
		Bind: []interface{}{ // 绑定 Go 结构体到前端，使得前端 JavaScript 可以调用 Go 的方法。
			app, // 绑定 app 结构体，使其方法可以被前端调用。
		},
		Frameless: true, // 设置窗口为无边框模式，创建更现代的界面。
	})

	// 检查 Wails 应用程序运行过程中是否发生错误。
	if err != nil {
		// 如果发生错误，打印错误信息到日志，方便调试和问题排查。
		log.Println("Error:", err.Error())
	}
}
