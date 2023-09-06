# local-test-for-cpp

ローカルテスト実行環境

# 環境

Windows11

# 要インストール

-   GO言語(1.20以降)
-   dagger
    ※powershellで`Invoke-WebRequest -UseBasicParsing -Uri https://dl.dagger.io/dagger/install.ps1 | Invoke-Expression`を実行
-   docker
    ※実行時dockerを起動している必要あり

# テスト実行コマンド

```ps1
dagger run go run <該当のmain.go>
```
