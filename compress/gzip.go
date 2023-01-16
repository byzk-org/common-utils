package compress

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"strings"
)

// CompressMemoryGzip 压缩tar.gz
func Gzip(srcFilePath, distFilePath string) error {
	_ = os.RemoveAll(distFilePath)
	distFile, err := os.Create(distFilePath)
	if err != nil {
		return errors.New("创建压缩文件失败")
	}
	gzipFile, err := gzip.NewWriterLevel(distFile, 9)
	if err != nil {
		return errors.New("创建压缩文件失败 => " + err.Error())
	}
	defer gzipFile.Close()
	tarWriter := tar.NewWriter(gzipFile)
	defer tarWriter.Close()
	open, err := os.Open(srcFilePath)
	if err != nil {
		return errors.New("打开文件失败 => " + err.Error())
	}
	defer open.Close()
	err = compressGzip(tarWriter, open, "")
	if err != nil {
		return err
	}
	return nil
}

func compressGzip(writer *tar.Writer, file *os.File, prefix string) error {
	var (
		err     error
		stat    os.FileInfo
		readdir []os.FileInfo
		fi      *os.File
		header  *tar.Header
	)

	defer file.Close()
	stat, err = file.Stat()
	if err != nil {
		return errors.New("读取目录结构失败 => " + err.Error())
	}
	if stat.IsDir() {
		readdir, err = file.Readdir(-1)
		if err != nil {
			if err != io.EOF {
				return errors.New("读取目录失败! => " + err.Error())
			}
		}
		for _, val := range readdir {
			fi, err = os.Open(file.Name() + "/" + val.Name())
			if err != nil {
				return errors.New("读取目录中的文件失败 => " + err.Error())
			}
			err = compressGzip(writer, fi, prefix+"/"+val.Name())
			if err != nil {
				return err
			}
		}
	} else {
		header, err = tar.FileInfoHeader(stat, "")
		if err != nil {
			return errors.New("创建压缩文件头信息失败! ")
		}
		header.Name = prefix
		err = writer.WriteHeader(header)
		if err != nil {
			return errors.New("写入压缩文信息失败 => " + err.Error())
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return errors.New("拷贝文件到压缩包内失败 => " + err.Error())
		}
	}
	return nil
}

func DeCompressGzip(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return errors.New("打开tar.gz文件失败 => " + err.Error())
	}
	defer srcFile.Close()
	return DeCompressGzipByReader(srcFile, dest)
}

func DeCompressGzipByReader(tarFile io.Reader, dest string) error {
	gr, err := gzip.NewReader(tarFile)
	if err != nil {
		return errors.New("创建gzip解压流失败 => " + err.Error())
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return errors.New("读取压缩包内文件失败 => " + err.Error())
			}
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return errors.New("创建本地临时存储文件失败 => " + err.Error())
		}
		_, _ = io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, errors.New("创建临时目录失败 => " + err.Error())
	}
	return os.Create(name)
}
