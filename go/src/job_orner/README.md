# make  

## clean  

```
make clean
```

## build

```
make
```

## run  
```
./job_orner
```

# 概要  
MaxWorkerの長さのキューを使ってジャッジのジョブを管理する。  
csvでデータをやり取りし、先頭にcodeSessionが入る。codeSessionにerrorが来るとエラー出力として判定する。  
デフォルトで、  
frontからの通信をport 4649  
judgeからの通信をport 5963  
で行う。
