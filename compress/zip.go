package compress

import (
	"archive/zip"
	"errors"
	"github.com/byzk-org/common-utils/commonutils"
	"io"
	"os"
	"path/filepath"
)

// Unzip 解压zip
func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return errors.New("创建zip解压流失败 => " + err.Error())
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := commonutils.PathJoin(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return errors.New("创建临时目录失败 => " + err.Error())
			}

			inFile, err := f.Open()
			if err != nil {
				return errors.New("打开临时目录失败 => " + err.Error())
			}
			defer func(file io.ReadCloser) {
				file.Close()
			}(inFile)

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return errors.New("打开临时文件失败 => " + err.Error())
			}
			defer func(file *os.File) {
				file.Close()
			}(outFile)

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return errors.New("从压缩包内解压文件失败 => " + err.Error())
			}
		}
	}
	return nil
}
