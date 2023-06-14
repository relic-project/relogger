package writers

import (
	"fmt"
	"github.com/relic-project/relogger/utils"
	"os"
	"sync"
	"time"
)

type AsyncFileLoggingWriter struct {
	currFile *os.File
	date     time.Time
	mx       sync.Mutex
	folder   string
}

func (f *AsyncFileLoggingWriter) Write(p []byte) (n int, err error) {
	go f._asyncWrite(p)
	return len(p), nil
}

func (f *AsyncFileLoggingWriter) _asyncWrite(p []byte) (n int, err error) {
	f.mx.Lock()
	defer f.mx.Unlock()
	if time.Now().Day() != f.date.Day() {
		oldDate := f.date
		if f.currFile != nil {
			err := f.currFile.Close()
			if err != nil {
				return 0, err
			}
		}
		// tar.gz old file
		go func() {
			archiveFileName := fmt.Sprintf("%s/%s.tar.gz", f.folder, oldDate.Format("2006-01-02"))
			logFileName := fmt.Sprintf("%s/%s.relog", f.folder, oldDate.Format("2006-01-02"))
			if err := archiveFileAndDelete(archiveFileName, logFileName); err != nil {
				println("relogger #", err)
				println("relogger # failed to archive old log file")
			}
		}()

		f.currFile = nil
		f.date = time.Now()
	}

	if f.currFile == nil {
		logFileName := fmt.Sprintf("%s/%s.relog", f.folder, f.date.Format("2006-01-02"))
		f.currFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			println("failed to write to file", err)
			return 0, err
		}
	}
	// write to file
	bytes, err := f.currFile.Write(p)
	if err != nil {
		println("failed to write to file", err)
	}
	return bytes, err
}
func NewFileLoggingWriter(folder string) *AsyncFileLoggingWriter {
	// create logs directory if not exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.Mkdir(folder, 0755); err != nil {
			panic(err)
		}
	} else {
		// zip old .relog files
		files, err := os.ReadDir(folder)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if file.Name()[len(file.Name())-6:] != ".relog" {
				continue
			}

			if file.Name() == time.Now().Format("2006-01-02")+".relog" {
				continue
			}
			// tar.gz old file sync
			archiveFileName := fmt.Sprintf("%s/%s.tar.gz", folder, file.Name()[:len(file.Name())-6])
			logFileName := fmt.Sprintf("%s/%s", folder, file.Name())

			if err := archiveFileAndDelete(archiveFileName, logFileName); err != nil {
				panic(err)
				println("relogger #", err)
				println("relogger # failed to archive old log file")
			}

		}
	}
	return &AsyncFileLoggingWriter{date: time.Now(), mx: sync.Mutex{}, folder: folder}
}

func archiveFileAndDelete(archiveFileName string, fileName string) error {
	// tar.gz old file sync
	// create archive file
	out, err1 := os.Create(archiveFileName)
	if err1 != nil {
		return err1
	}

	if err := utils.CreateArchive(out, fileName); err != nil {
		return err
	}

	if err := os.Remove(fileName); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return nil
}

var DefaultFileLoggingWriter = NewFileLoggingWriter("./logs")
