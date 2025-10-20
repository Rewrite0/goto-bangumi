package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

func getData(id int64) string {
	fmt.Println("query...")
	time.Sleep(3 * time.Second) // 模拟一个比较耗时的操作
	return "liwenzhou.com"
}

func test_reg() {
	MikanSeasonPaattern := regexp.MustCompile(`\s(?:第(.)季|(贰))$`)

	// title := "拥有超常技能的异世界流浪美食家 第二季"
	title := "妖怪旅馆营业中 贰"
	res := MikanSeasonPaattern.FindStringSubmatch(title)
	for v, i := range res {
		fmt.Println("v=", v, "i=", i)
	}
}

// type User struct {
//     gorm.Model
//     Name      string
//     Languages []Language `gorm:"many2many:user_languages;"` // 连接表
// }

type Language struct {
	gorm.Model
	Name  string
	Users []User `gorm:"many2many:user_languages;"` // 连接表
}

type User struct {
	gorm.Model
	Name      string
	Addresses []Address // 一个用户可以有多个地址
}

type Address struct {
	gorm.Model
	Address1 string
	UserID   uint // 外键字段，指向 User
	User     User // 关联的 User 对象
}

func main() {
	client := resty.New()
	client.SetBaseURL("https://api.github.com")
	fmt.Println(client.BaseURL)
	client.SetBaseURL("https://api.example.com")
	fmt.Println(client.BaseURL)
}

// test_reg()
// g := new(singleflight.Group)
//
// // 第1次调用
// go func() {
// 	v1, _, shared := g.Do("getData", func() (interface{}, error) {
// 		ret := getData(1)
// 		return ret, nil
// 	})
// 	fmt.Printf("1st call: v1:%v, shared:%v\n", v1, shared)
// }()
//
// time.Sleep(2 * time.Second)
//
// // 第2次调用（第1次调用已开始但未结束）
// v2, _, shared := g.Do("getData", func() (interface{}, error) {
// 	ret := getData(1)
// 	return ret, nil
// })
// fmt.Printf("2nd call: v2:%v, shared:%v\n", v2, shared)
