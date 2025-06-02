package utils // 定义包名为 utils，包含一些常用的工具函数。

import (
	"crypto/md5"    // 导入 crypto/md5 包，用于计算 MD5 哈希值。
	"encoding/hex"  // 导入 encoding/hex 包，用于十六进制编码和解码。
	"errors"        // 导入 errors 包，用于创建和处理错误。
	"fmt"           // 导入 fmt 包，用于格式化输入输出。
	"io"            // 导入 io 包，提供了基本的 I/O 接口。
	"log"           // 导入 log 包，用于记录日志。
	"os"            // 导入 os 包，提供了与操作系统交互的函数。
	"os/exec"       // 导入 os/exec 包，用于执行外部命令。
	"path/filepath" // 导入 path/filepath 包，用于处理文件路径。
	"regexp"        // 导入 regexp 包，用于正则表达式操作。
	"strings"       // 导入 strings 包，用于字符串操作。

	"github.com/pkg/browser"     // 导入 browser 包，用于在默认浏览器中打开文件或 URL。
	"github.com/shirou/gopsutil/v3/disk" // 导入 gopsutil/disk 包，用于获取磁盘使用情况。
	"golang.org/x/net/html"      // 导入 golang.org/x/net/html 包，用于 HTML 解析。
	"golang.org/x/sys/windows/registry" // 导入 windows/registry 包，用于访问 Windows 注册表。
)

// PathStat 结构体定义了路径的统计信息。
type PathStat struct {
	Path        string  `json:"path"`        // 路径字符串。
	Total       uint64  `json:"total"`       // 总空间大小（字节）。
	Free        uint64  `json:"free"`        // 可用空间大小（字节）。
	Used        uint64  `json:"used"`        // 已用空间大小（字节）。
	UsedPercent float64 `json:"usedPercent"` // 已用空间百分比。
}

// getDefaultProgram 函数用于获取指定文件扩展名的默认打开程序。
// fileExtension 参数是文件扩展名（不带点）。
// 返回默认程序的名称和可能发生的错误。
func getDefaultProgram(fileExtension string) (string, error) {
	// 打开注册表中的 HKEY_CLASSES_ROOT 键，查找与文件扩展名关联的默认程序。
	key, err := registry.OpenKey(registry.CLASSES_ROOT, fmt.Sprintf(`.%s`, fileExtension), registry.QUERY_VALUE)
	if err != nil {
		return "", err // 如果打开注册表键失败，返回错误。
	}
	defer key.Close() // 确保在函数返回前关闭注册表键。

	// 读取默认程序关联值。
	defaultProgram, _, err := key.GetStringValue("")
	if err != nil {
		return "", err // 如果读取字符串值失败，返回错误。
	}

	return defaultProgram, nil // 返回默认程序名称。
}

// hasDefaultProgram 函数用于检查指定文件扩展名是否有默认打开程序。
// fileExtension 参数是文件扩展名（不带点）。
// 返回一个布尔值，true 表示有默认程序，false 表示没有。
func hasDefaultProgram(fileExtension string) bool {
	prog, err := getDefaultProgram(fileExtension) // 获取默认程序。
	if err != nil {
		log.Println("getDefaultProgram Error:", err) // 如果获取失败，打印错误日志。
		return false                                 // 返回 false。
	}

	if prog == "" {
		return false // 如果程序名为空，返回 false。
	}

	return true // 返回 true。
}

// OpenFileOrExplorer 函数用于打开文件或在资源管理器中显示文件。
// filePath 参数是文件路径，explorer 参数表示是否在资源管理器中打开（true）或直接打开文件（false）。
// 返回可能发生的错误。
func OpenFileOrExplorer(filePath string, explorer bool) error {
	if _, err := os.Stat(filePath); err != nil {
		log.Printf("%s %v\n", filePath, err) // 如果文件不存在，打印错误日志。
		return err                           // 返回错误。
	}

	canOpen := false
	fileExtension := ""
	index := strings.LastIndex(filePath, ".") // 查找文件扩展名的起始位置。
	if index > 0 {
		fileExtension = filePath[index+1:] // 提取文件扩展名。
		canOpen = hasDefaultProgram(fileExtension) // 检查是否有默认程序打开此类型文件。
	}

	if canOpen && !explorer {
		return browser.OpenFile(filePath) // 如果有默认程序且不要求在资源管理器中打开，则直接打开文件。
	}

	commandArgs := []string{"/select,", filePath} // 构建 explorer 命令的参数，用于选择文件。
	fmt.Println("cmd:", "explorer", commandArgs)  // 打印命令。

	// 创建一个 Cmd 结构体表示要执行的命令。
	cmd := exec.Command("explorer", commandArgs...)

	// 执行命令并等待它完成。
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %s\n", err) // 如果执行命令失败，打印错误日志。
		// return err
	}

	fmt.Println("Command executed successfully") // 打印命令执行成功信息。
	return nil                                   // 返回 nil 表示成功。
}

// GetPathStat 函数用于获取指定路径的磁盘使用情况统计信息。
// path 参数是需要统计的路径。
// 返回 PathStat 结构体和可能发生的错误。
func GetPathStat(path string) (PathStat, error) {
	pathStat := PathStat{}             // 创建 PathStat 实例。
	absPath, err := filepath.Abs(path) // 获取路径的绝对路径。
	if err != nil {
		return pathStat, err // 如果获取绝对路径失败，返回错误。
	}

	stat, err := disk.Usage(absPath) // 获取磁盘使用情况。
	if err != nil {
		return pathStat, err // 如果获取失败，返回错误。
	}

	// 填充 PathStat 结构体。
	pathStat.Path = stat.Path
	pathStat.Total = stat.Total
	pathStat.Used = stat.Used
	pathStat.Free = stat.Free
	pathStat.UsedPercent = stat.UsedPercent

	return pathStat, nil // 返回 PathStat 结构体。
}

// PathIsCanWriteFile 函数用于检查指定路径是否可写。
// path 参数是需要检查的路径。
// 返回一个布尔值，true 表示可写，false 表示不可写。
func PathIsCanWriteFile(path string) bool {

	filepath := fmt.Sprintf("%s\\CanWrite.txt", path) // 构建一个临时文件路径。
	// 尝试创建并打开文件，如果成功则表示可写。
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false // 如果打开文件失败，返回 false。
	}

	file.Close()       // 关闭文件。
	os.Remove(filepath) // 删除临时文件。

	return true // 返回 true 表示可写。
}

// CopyFile 函数用于复制文件。
// src 参数是源文件路径，dst 参数是目标文件路径。
// 返回复制的字节数和可能发生的错误。
func CopyFile(src, dst string) (int64, error) {
	stat, err := os.Stat(src) // 获取源文件信息。
	if err != nil {
		return 0, err // 如果获取失败，返回错误。
	}
	if stat.IsDir() {
		return 0, errors.New(src + " is dir") // 如果源路径是目录，返回错误。
	}
	sourceFile, err := os.Open(src) // 打开源文件。
	if err != nil {
		return 0, err // 如果打开失败，返回错误。
	}
	defer sourceFile.Close() // 确保在函数返回前关闭源文件。

	destFile, err := os.Create(dst) // 创建目标文件。
	if err != nil {
		return 0, err // 如果创建失败，返回错误。
	}
	defer destFile.Close() // 确保在函数返回前关闭目标文件。

	bytesWritten, err := io.Copy(destFile, sourceFile) // 复制文件内容。
	if err != nil {
		return bytesWritten, err // 如果复制失败，返回错误。
	}

	return bytesWritten, nil // 返回复制的字节数。
}

// extractTextFromHTML 函数用于从 HTML 字符串中提取纯文本。
// htmlStr 参数是 HTML 字符串。
// 返回提取的纯文本。
func extractTextFromHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr)) // 解析 HTML 字符串。
	if err != nil {
		fmt.Println("Error parsing HTML:", err) // 如果解析失败，打印错误日志。
		return ""                               // 返回空字符串。
	}

	var extractText func(*html.Node) string // 定义一个递归函数用于提取文本。
	extractText = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data // 如果是文本节点，返回其数据。
		}

		var text string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text += extractText(c) // 递归提取子节点的文本。
		}
		return text // 返回所有子节点的文本。
	}

	return extractText(doc) // 从解析后的文档中提取文本。
}

// removeCustomTags 函数用于从字符串中移除自定义标签。
// input 参数是包含自定义标签的字符串。
// 返回移除标签后的字符串。
func removeCustomTags(input string) string {

	re := regexp.MustCompile(`<(_wc_custom_link_)[^>]*?>`) // 定义匹配自定义标签的正则表达式。
	return re.ReplaceAllString(input, `$2`)               // 替换匹配到的标签。
}

// Html2Text 函数用于将 HTML 字符串转换为纯文本。
// htmlStr 参数是 HTML 字符串。
// 返回转换后的纯文本。
func Html2Text(htmlStr string) string {
	// if htmlStr == "" {
	// 	return ""
	// }

	if len(htmlStr) == 0 || htmlStr[0] != '<' {
		return htmlStr // 如果不是 HTML 字符串，直接返回。
	}

	text := extractTextFromHTML(htmlStr) // 从 HTML 中提取纯文本。
	if strings.Contains(text, `<_wc_custom_link_`) {
		text = "\U0001F9E7" + removeCustomTags(text) // 如果包含自定义链接，添加表情并移除标签。
	}

	return text // 返回转换后的文本。
}

// HtmlMsgGetAttr 函数用于从 HTML 消息中获取指定标签的属性。
// htmlStr 参数是 HTML 字符串，tag 参数是标签名。
// 返回一个包含属性键值对的 map。
func HtmlMsgGetAttr(htmlStr, tag string) map[string]string {

	doc, err := html.Parse(strings.NewReader(htmlStr)) // 解析 HTML 字符串。
	if err != nil {
		fmt.Println("Error parsing HTML:", err) // 如果解析失败，打印错误日志。
		return nil                              // 返回 nil。
	}

	var attributes map[string]string
	var findAttributes func(*html.Node) // 定义一个递归函数用于查找属性。
	findAttributes = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			attributes = make(map[string]string) // 创建属性 map。
			for _, attr := range n.Attr {
				attributes[attr.Key] = attr.Val // 将属性添加到 map 中。
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findAttributes(c) // 递归查找子节点的属性。
		}
	}

	findAttributes(doc) // 从解析后的文档中查找属性。
	return attributes   // 返回属性 map。
}

// Hash256Sum 函数用于计算数据的 MD5 哈希值并以十六进制字符串形式返回。
// data 参数是需要哈希的字节切片。
// 返回 MD5 哈希值的十六进制字符串表示。
func Hash256Sum(data []byte) string {
	hash := md5.New()      // 创建一个新的 MD5 哈希实例。
	hash.Write([]byte(data)) // 将数据写入哈希实例。
	hashSum := hash.Sum(nil) // 计算哈希值。

	return hex.EncodeToString(hashSum) // 将哈希值编码为十六进制字符串并返回。
}
