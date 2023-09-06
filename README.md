# local-test-for-cpp

C++ ローカルテスト実行環境

## 環境

Windows11

## 依存ソフトのインストール/準備

-   GO言語(1.21以降)
-   dagger
    ※powershellで`Invoke-WebRequest -UseBasicParsing -Uri https://dl.dagger.io/dagger/install.ps1 | Invoke-Expression`を実行
-   docker
    ※実行時dockerを起動している必要あり
-   main.go がある階層で`go mod tidy`実行
-   `setting.json.sample`をコピーし、`setting.json`を作成の上、内容を適宜変更

## 指定可能な C++ プロジェクトの条件

- `setting.json`の`target_dir_path`には、`CMakeLists.txt`が配置されていること
- `target_dir_path`の階層で`ctest`コマンドを打ち込んだ場合、テストが走るような構成であること

## テスト実行コマンド

```ps1
dagger run go run main.go
```
