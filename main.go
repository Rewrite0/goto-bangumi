package main

import (
	"fmt"
	"goto-bangumi/internal/parser"
)
func main() {
	t := parser.NewTitleMetaParse()
	// title := "【幻樱字幕组】【4月新番】【古见同学有交流障碍症 第二季 Komi-san wa, Komyushou Desu. S02】【22】【GB_MP4】【1920X1080】"
	// title :="[织梦字幕组][尼尔：机械纪元 NieR Automata Ver1.1a][02集][1080P][AVC][简日双语]"
	// title :="[梦蓝字幕组]New Doraemon 哆啦A梦新番[747][2023.02.25][AVC][1080P][GB_JP][MP4]"
	title := "[ANi] Grand Blue Dreaming /  GRAND BLUE 碧蓝之海 2 - 04 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]"
	ans :=t.Parse(title)
	fmt.Println(ans)

	

}
// func main() {
// 	tr := http.Transport{}
// 	header := http.Header{}
// 	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
// 	client := http.Client{Transport: &tr}
// 	resp, err := client.Get("https://www.baidu.com")
// 	body, err := io.ReadAll(resp.Body)
// 	fmt.Println(resp, err)
// 	// 获取其中的文本
// 	fmt.Println(string(body))
// }
