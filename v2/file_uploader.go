package toolkit

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	files, err := t.UploadFiles(r, uploadDir, rename...)
	if err != nil {
		return nil, err
	}
	return files[0], nil

}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := (len(rename) > 0 && rename[0]) || false
	var uploadedFileParts []*UploadedFile
	// if t.MaxFileSize == 0 {
	// 	t.MaxFileSize = 1024 * 1024 * 1024
	// }
	if err := r.ParseMultipartForm(t.MaxFileSize); err != nil {
		return nil, errors.New("file is too big")
	}

	if err := t.CreateDirIfNotExist(uploadDir); err != nil {
		return nil, err
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			// log.Println(hdr)
			// log.Println(fHeaders)
			parts, err := func(uploadedFileParts []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()
				buf := make([]byte, 512)
				_, err = infile.Read(buf)
				if err != nil {
					return nil, err
				}
				allowed := false
				fileType := http.DetectContentType(buf)
				// allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
				if len(t.AllowedTypes) > 0 {
					for _, at := range t.AllowedTypes {
						if strings.EqualFold(at, fileType) {
							allowed = true
						}
					}
				}
				if !allowed {
					return nil, errors.New("uploaded file type not permitted")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				uploadedFile.OriginalFileName = hdr.Filename

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				var outfile *os.File
				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					filesize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = filesize
				}
				defer outfile.Close()
				uploadedFileParts = append(uploadedFileParts, &uploadedFile)
				// log.Println("inner", len(uploadedFileParts))
				return uploadedFileParts, nil

			}(uploadedFileParts)

			if err != nil {
				// log.Println("err", len(uploadedFileParts))
				return nil, err
			} else {
				uploadedFileParts = parts
			}
		}
	}
	// log.Println("outer", len(uploadedFileParts))
	return uploadedFileParts, nil
}
