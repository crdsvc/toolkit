package toolkit

import (
	"fmt"
	"net/http"
)

func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, filePath, displayName string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))
	http.ServeFile(w, r, filePath)

}
