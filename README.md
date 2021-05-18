## 简介
Time-NLP 中文语句中的时间语义识别的golang版本   

python 版本https://github.com/sunfiyes/Time-NLPY  

python3 版本 https://github.com/zhanzecheng/Time_NLP

Java 版本https://github.com/shinyke/Time-NLP

PHP 版本https://github.com/crazywhalecc/Time-NLP-PHP

## 使用
go get -u github.com/bububa/TimeNLP

## 功能说明
用于句子中时间词的抽取和转换 
```golang
import (
    "log"

    "github.com/bububa/TimeNLP"
)

func main() {
    target := "Hi，all.下周一下午三点开会"
    preferFuture := true
    tn := timenlp.TimeNormalizer(preferFuture)
    ret, err := tn.Parse(target)
    if err != nil {
        log.Fatalln(err)
    }
    log.Printf("%+v\n", ret)
}
```
