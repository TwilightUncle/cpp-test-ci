package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	log_output_dir = "./log"
)

func WithOutputLog(filename string, callback func(*slog.Logger) error) error {
	t := time.Now()

	// ディレクトリ作成(必要な場合のみ)
	if err := makeOutputDir(t); err != nil {
		return err
	}

	// ファイル生成(既存の場合上書き)
	fp, err := os.OpenFile(
		filepath.Join(getOutputDirPath(t), strconv.FormatInt(t.Unix(), 10)+"-"+filename),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0777,
	)
	if err != nil {
		return err
	}
	defer fp.Close()

	return callback(slog.New(slog.NewTextHandler(fp, nil)))
}

// 出力先ディレクトリが存在しない時のみ作成する
func makeOutputDir(t time.Time) error {
	out_dirpath := getOutputDirPath(time.Now())
	f, err := os.Stat(out_dirpath)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(out_dirpath, 0777); err != nil {
			return err
		}
	} else if !f.IsDir() {
		return fmt.Errorf(
			"A file with the same name as the directory you are trying to create exists. filename: \"%s\"",
			out_dirpath,
		)
	}

	return nil
}

func getOutputDirPath(t time.Time) string {
	return filepath.Join(log_output_dir, t.Format("20060102"))
}
