package main // import "github.com/dgodd/websitechange"

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("USAGE: %s [URL]\n", os.Args[0])
		os.Exit(1)
	}

	site := os.Args[1]
	if err := download(site); err != nil {
		panic(err)
	}
}

func download(site string) error {
	siteMD5 := fmt.Sprintf("%x", md5.Sum([]byte(site)))

	res, err := http.Get(site)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	html, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download: %d", res.StatusCode)
	}
	sha := fmt.Sprintf("%x", sha256.Sum256(html))

	headers := res.Header
	headers["SHA256"] = []string{sha}

	changed, err := writeIfNew(filepath.Join(siteMD5, "pages", sha+".html"), html)
	if err != nil {
		return err
	}

	if changed {
		fmt.Println("Website has CHANGED:", filepath.Join(siteMD5, "pages", sha+".html"))
	} else {
		fmt.Println("Website is unchanged:", filepath.Join(siteMD5, "pages", sha+".html"))
	}

	os.MkdirAll(filepath.Join(siteMD5, "dates"), 0777)
	headerPath := filepath.Join(siteMD5, "dates", time.Now().Format(time.RFC3339)+".toml")
	if changed {
		headerPath = filepath.Join(siteMD5, "dates", time.Now().Format(time.RFC3339)+"-changed.toml")
	}
	fh, err := os.Create(headerPath)
	if err != nil {
		return err
	}
	defer fh.Close()
	return toml.NewEncoder(fh).Encode(headers)
}

func writeIfNew(path string, txt []byte) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	os.MkdirAll(filepath.Dir(path), 0777)
	fh, err := os.Create(path)
	if err != nil {
		return false, err
	}
	defer fh.Close()
	if _, err := fh.Write(txt); err != nil {
		return false, err
	}
	return true, nil
}
