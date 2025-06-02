package wechat
/**
这段代码是一个用于解密微信数据库文件的Go语言实现。我来逐步解释其主要功能和实现逻辑：

### 1. 包和导入
- 包名：`wechat`，说明这是处理微信相关功能的模块
- 导入的库主要用于加密解密、哈希计算和文件操作

### 2. 常量定义
- `keySize = 32`：密钥长度32字节
- `defaultIter = 64000`：PBKDF2算法的默认迭代次数
- `defaultPageSize = 4096`：默认页大小4KB（SQLite数据库的常见页大小）

### 3. 主要函数 `DecryptDataBase`
这是核心解密函数，参数：
- `path`：加密的数据库文件路径
- `password`：解密密码
- `expPath`：解密后输出的文件路径

#### 解密流程：
1. **准备SQLite文件头**：`"SQLite format 3"`加一个空字节
2. **打开加密文件**：使用缓冲读取器提高性能
3. **读取第一页(4KB)**：获取salt值和加密数据
4. **密钥派生**：
   - 使用PBKDF2-HMAC-SHA1算法从密码和salt派生主密钥
   - 计算macSalt（salt与0x3a异或）
   - 派生macKey用于验证
5. **密码验证**：
   - 计算HMAC并验证密码是否正确
6. **创建输出文件**并写入SQLite文件头
7. **解密第一页**：
   - 使用AES-CBC模式解密
   - 从数据中提取IV(初始化向量)
   - 解密后写入输出文件
8. **解密剩余页**：
   - 循环读取每一页
   - 每页最后48字节包含IV等信息
   - 使用相同方式解密并写入

### 4. 辅助函数
- `pbkdf2HMAC`：实现PBKDF2密钥派生算法
  - 使用HMAC-SHA1
  - 支持多次迭代增强安全性
- `xorBytes`：字节数组与单个字节异或

### 5. 注释掉的main函数
这是测试代码示例，展示了如何使用这个解密功能

### 技术要点
1. 加密方案：AES-CBC模式
2. 密钥派生：PBKDF2-HMAC-SHA1
3. 完整性验证：HMAC-SHA1
4. 文件格式：处理SQLite数据库的特殊结构
5. 分页处理：按4KB页处理数据库文件

### 典型用途
这段代码很可能是用于解密微信本地存储的加密数据库文件（如聊天记录数据库），使其恢复为标准SQLite数据库格式以便查看或分析。

注意：代码中省略了`decryptMsg`函数的实现，但从上下文看它可能是`DecryptDataBase`的早期版本或别名。

</think>
**/

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

const (
	keySize         = 32
	defaultIter     = 64000
	defaultPageSize = 4096
)

func DecryptDataBase(path string, password []byte, expPath string) error {
	sqliteFileHeader := []byte("SQLite format 3")
	sqliteFileHeader = append(sqliteFileHeader, byte(0))

	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	fpReader := bufio.NewReaderSize(fp, defaultPageSize*100)
	// fpReader := bufio.NewReader(fp)

	buffer := make([]byte, defaultPageSize)

	n, err := fpReader.Read(buffer)
	if err != nil && n != defaultPageSize {
		return fmt.Errorf("read failed")
	}

	salt := buffer[:16]
	key := pbkdf2HMAC(password, salt, defaultIter, keySize)

	page1 := buffer[16:defaultPageSize]

	macSalt := xorBytes(salt, 0x3a)
	macKey := pbkdf2HMAC(key, macSalt, 2, keySize)

	hashMac := hmac.New(sha1.New, macKey)
	hashMac.Write(page1[:len(page1)-32])
	hashMac.Write([]byte{1, 0, 0, 0})

	if !hmac.Equal(hashMac.Sum(nil), page1[len(page1)-32:len(page1)-12]) {
		return fmt.Errorf("incorrect password")
	}

	outFilePath := expPath
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write SQLite file header
	_, err = outFile.Write(sqliteFileHeader)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	page1 = buffer[16:defaultPageSize]
	iv := page1[len(page1)-48 : len(page1)-32]
	stream := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(page1)-48)
	stream.CryptBlocks(decrypted, page1[:len(page1)-48])
	_, err = outFile.Write(decrypted)
	if err != nil {
		return err
	}
	_, err = outFile.Write(page1[len(page1)-48:])
	if err != nil {
		return err
	}

	for {
		n, err = fpReader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		} else if n < defaultPageSize {
			return fmt.Errorf("read data to short %d", n)
		}

		iv := buffer[len(buffer)-48 : len(buffer)-32]
		stream := cipher.NewCBCDecrypter(block, iv)
		decrypted := make([]byte, len(buffer)-48)
		stream.CryptBlocks(decrypted, buffer[:len(buffer)-48])
		_, err = outFile.Write(decrypted)
		if err != nil {
			return err
		}
		_, err = outFile.Write(buffer[len(buffer)-48:])
		if err != nil {
			return err
		}
	}

	return nil
}

func pbkdf2HMAC(password, salt []byte, iter, keyLen int) []byte {
	dk := make([]byte, keyLen)
	loop := (keyLen + sha1.Size - 1) / sha1.Size
	key := make([]byte, 0, len(salt)+4)
	u := make([]byte, sha1.Size)
	for i := 1; i <= loop; i++ {
		key = key[:0]
		key = append(key, salt...)
		key = append(key, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
		hmac := hmac.New(sha1.New, password)
		hmac.Write(key)
		digest := hmac.Sum(nil)
		copy(u, digest)
		for j := 2; j <= iter; j++ {
			hmac.Reset()
			hmac.Write(digest)
			digest = hmac.Sum(digest[:0])
			for k, di := range digest {
				u[k] ^= di
			}
		}
		copy(dk[(i-1)*sha1.Size:], u)
	}
	return dk
}

func xorBytes(a []byte, b byte) []byte {
	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b
	}
	return result
}

/*
func main() {

	str := "82b1a210335140a1bc8a57397391186494abe666595b4f408095538b5518f7d5"
	// 将十六进制字符串解码为字节
	password, err := hex.DecodeString(str)
	if err != nil {
		fmt.Println("解码出错:", err)
		return
	}

	fmt.Println(hex.EncodeToString(password))

	err = decryptMsg("Media.db", password)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Decryption successful!")
	}
}
*/
