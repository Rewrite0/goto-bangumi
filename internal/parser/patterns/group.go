package patterns

import "github.com/dlclark/regexp2"

// GroupRe 字幕组名称匹配
var GroupRe = regexp2.MustCompile(
	BoundaryStart+`
    (ANi
    |LoliHouse
    |SweetSub
    |Pre-S
    |H-Enc
    |TOC
    |Billion Meta Lab
    |Lilith-Raws
    |DBD-Raws
    |NEO·QSW
    |SBSUB
    |MagicStar
    |7³ACG
    |KitaujiSub
    |Doomdos
    |Prejudice-Studio
    |GM-Team
    |VCB-Studio
    |神椿观测站
    |极影字幕社
    |百冬练习组
    |猎户手抄部
    |喵萌奶茶屋
		|喵萌Production
    |萌樱字幕组
    |三明治摆烂组
    |绿茶字幕组
    |梦蓝字幕组
    |幻樱字幕组
    |织梦字幕组
    |北宇治字组
    |北宇治字幕组
    |霜庭云花Sub
    |氢气烤肉架
    |豌豆字幕组
    |豌豆
    |DBD
    |风之圣殿字幕组
    |黒ネズミたち
    |桜都字幕组
    |漫猫字幕组
    |猫恋汉化组
    |黑白字幕组
    |猎户压制部
    |猎户手抄部
    |沸班亚马制作组
    |星空字幕组
    |光雨字幕组
    |樱桃花字幕组
    |动漫国字幕组
    |动漫国
    |千夏字幕组
    |SW字幕组
    |澄空学园
    |华盟字幕社
    |诸神字幕组
    |雪飘工作室
    |❀拨雪寻春❀
    |夜莺家族
    |YYQ字幕组
    |APTX4869
    |Prejudice-Studio
    |丸子家族
    |六四位元字幕组
    )
    `+BoundaryEnd,
	regexp2.IgnorePatternWhitespace,
)
