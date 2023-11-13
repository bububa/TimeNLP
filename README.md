# TimeNLP 中文语句中的时间语义识别的golang版本   

[![Go Reference](https://pkg.go.dev/badge/github.com/bububa/TimeNLP.svg)](https://pkg.go.dev/github.com/bububa/TimeNLP)
[![Go](https://github.com/bububa/TimeNLP/actions/workflows/go.yml/badge.svg)](https://github.com/bububa/TimeNLP/actions/workflows/go.yml)
[![goreleaser](https://github.com/bububa/TimeNLP/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/bububa/TimeNLP/actions/workflows/goreleaser.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/bububa/TimeNLP.svg)](https://github.com/bububa/TimeNLP)
[![GoReportCard](https://goreportcard.com/badge/github.com/bububa/TimeNLP)](https://goreportcard.com/report/github.com/bububa/TimeNLP)
[![GitHub license](https://img.shields.io/github/license/bububa/TimeNLP.svg)](https://github.com/bububa/TimeNLP/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/bububa/TimeNLP.svg)](https://gitHub.com/bububa/TimeNLP/releases/) 

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

## Reference 
python3 版本 https://github.com/zhanzecheng/Time_NLP

Java 版本https://github.com/shinyke/Time-NLP

PHP 版本https://github.com/crazywhalecc/Time-NLP-PHP

Javascript 版本https://github.com/JohnnieFucker/ChiTimeNLP

