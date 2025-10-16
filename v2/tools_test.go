package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

func Test_Tools_RandomString(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("length not matching")
	}
}

var uploadFilesTestCases = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	// {name: "allowed rename", allowedTypes: []string{"image/jpg", "image/png", "image/jpeg"}, renameFile: true, errorExpected: false},
	// {name: "allowed no rename", allowedTypes: []string{"image/jpg", "image/png", "image/jpeg"}, renameFile: false, errorExpected: false},
	{name: "not allowed file", allowedTypes: []string{"image/jpg", "image/jpeg"}, renameFile: false, errorExpected: true},
}

func Test_Tools_UploadFiles(t *testing.T) {
	for _, e := range uploadFilesTestCases {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()
			part, err := writer.CreateFormFile("file", "./testdata/img_0gi4neu8.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img_0gi4neu8.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Error(err)
			}
			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())
		var testTools Tools
		testTools.AllowedTypes = e.allowedTypes
		files, err := testTools.UploadFiles(request, "./testdata/uploads", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			_, err = os.Stat(fmt.Sprintf("./testdata/uploads/%s", files[0].NewFileName))
			if err != nil {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			// _ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", files[0].NewFileName))
		}

		if err == nil && e.errorExpected {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()

	}
}

func Test_Tools_UploadOneFile(t *testing.T) {
	for _, e := range uploadFilesTestCases {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		go func() {
			defer writer.Close()
			part, err := writer.CreateFormFile("file", "./testdata/img_0gi4neu8.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img_0gi4neu8.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Error(err)
			}
			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())
		var testTools Tools
		testTools.AllowedTypes = e.allowedTypes
		file, err := testTools.UploadOneFile(request, "./testdata/uploads", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			_, err = os.Stat(fmt.Sprintf("./testdata/uploads/%s", file.NewFileName))
			if err != nil {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			// _ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", files[0].NewFileName))
		}

		if err == nil && e.errorExpected {
			t.Errorf("%s: error expected but none received", e.name)
		}
	}
}

func Test_Tools_CreateDirIfNotExist(t *testing.T) {
	var testTools Tools
	path := "./testdata/myDir"
	err := testTools.CreateDirIfNotExist(path)
	if err != nil {
		t.Error(err)
	}
	err = testTools.CreateDirIfNotExist(path)
	if err != nil {
		t.Error(err)
	}
	err = os.Remove(path)
	if err != nil {
		t.Error(err)
	}
}

var slugifyTestCases = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "now is the time", expected: "now-is-the-time", errorExpected: false},
	{name: "empty string", s: "", expected: "whatever ...", errorExpected: true},
	{name: "complex string", s: "now is the TiMe !@#@!# for all MeN __** 123+& @", expected: "now-is-the-time-for-all-men-123", errorExpected: false},
	{name: "japanese string", s: "こんにちは世界", expected: "", errorExpected: true},
	{name: "japanese and roman string", s: "hello こんに worldちは世界", expected: "hello-world", errorExpected: false},
}

func Test_Tools_Slugify(t *testing.T) {
	var testTools Tools
	for _, c := range slugifyTestCases {
		o, err := testTools.Slugify(c.s)
		if err != nil && !c.errorExpected {
			t.Errorf("%s: error received but not expected", c.name)
		}
		if !strings.EqualFold(o, c.expected) && !c.errorExpected {
			t.Errorf("%s expected %s but received %s", c.name, c.expected, o)
		}
	}
}

func Test_Tools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	var testTools Tools
	testTools.DownloadStaticFile(rr, req, "./testdata/some_download.jpg", "noise.jpg")
	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "216221" {
		t.Errorf("File size expected is %s but received %s", "216221", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachment; filename=\"noise.jpg\"" {
		t.Error("incorrect content disposition")
	}

	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
}{
	{name: "good json", json: `{"name": "deb", "age": 90}`, errorExpected: false, maxSize: 1024, allowUnknown: false},
	{name: "badly formatted json", json: `{"name": }`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type json", json: `{"name": "deb", "age": "90"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "two jsons", json: `{"name": "deb", "age": 90} {"name": "deb", "age": 90}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "empty body json", json: ``, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "syntax json", json: `{"name": "deb" "age": 90}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "dont allow unknown field json", json: `{"namyyy": "deb", "age": 90}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "allow unknown field json", json: `{"nameyyy": "deb", "age": 90}`, errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "missing field json", json: `{name: "deb"}`, errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "file too large json", json: `{"name": "deb", "age": 90}`, errorExpected: true, maxSize: 4, allowUnknown: true},
	{name: "not json", json: `hello world`, errorExpected: true, maxSize: 1024, allowUnknown: true},
}

func Test_Tools_ReadJSON(t *testing.T) {
	for _, e := range jsonTests {
		var testTools Tools
		testTools.MaxJSONBytes = e.maxSize
		testTools.AllowUnknownFields = e.allowUnknown
		var decodedJSON struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Errorf("%s: error creating request", err.Error())
		}
		defer req.Body.Close()
		rec := httptest.NewRecorder()
		err = testTools.ReadJSON(rec, req, &decodedJSON)
		if err == nil && e.errorExpected {
			t.Errorf("%s: error expected but none received", e.name)
		}
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error not expected but received - %s", e.name, err.Error())
		}
	}

}

func Test_Tools_WriteJSON(t *testing.T) {
	var testTools Tools
	rec := httptest.NewRecorder()
	jres := JSONResponse{
		Error:   false,
		Message: "hello world",
		Data:    `{"name": "dev","age":90}`,
	}
	hdr := make(http.Header)
	hdr.Add("foo", "bar")
	err := testTools.WriteJSON(rec, http.StatusOK, jres, hdr)
	if err != nil {
		t.Errorf("failed to write JSON %v", err)
	}

}

func Test_Tools_ErrorJSON(t *testing.T) {
	var testTools Tools
	rec := httptest.NewRecorder()
	err := testTools.ErrorJSON(rec, errors.New("this is a new error for testing"), http.StatusServiceUnavailable)
	if err != nil {
		t.Errorf("failed to write ERROR JSON %v", err)
	}

	var payload JSONResponse
	dec := json.NewDecoder(rec.Body)
	err = dec.Decode(&payload)
	if err != nil {
		t.Errorf("failed to decode JSON %v", err)
	}
	if !payload.Error {
		t.Errorf("expected json with error true but got false")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected json with status %d", http.StatusServiceUnavailable)
	}

}

type RoundTripFunc func(*http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func Test_Tools_PushJSONTORemote(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
			Header:     make(http.Header),
		}
	})

	var testTools Tools
	res, _, err := testTools.PushJSONTORemote("https://example.com/somepath/etc", map[string]string{"hello": "world"}, client)
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected %d but received %d", http.StatusOK, res.StatusCode)
	}
	if err != nil {
		t.Errorf("error happeneed %s", err.Error())
	}

}
