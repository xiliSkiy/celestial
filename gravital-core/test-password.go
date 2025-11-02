package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"
	hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	
	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("❌ 密码验证失败: %v\n", err)
		
		// 生成新的哈希
		fmt.Println("\n生成新的密码哈希...")
		newHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			fmt.Printf("生成哈希失败: %v\n", err)
			return
		}
		fmt.Printf("新的密码哈希: %s\n", string(newHash))
	} else {
		fmt.Println("✅ 密码验证成功！")
	}
}

