package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/bits"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

const (
	metaFilename           = "notes.meta"
	logFilenameFormat      = "%x_%d.log"
	logFilenameRegexFormat = `[0-9a-z]+_[0-9]+\.log`
)

type meta struct {
	Title   string    `json:"title"`
	Created time.Time `json:"created"`
}

func (m *meta) UnmarshalJSON(b []byte) error {
	var tmp = struct {
		Title   string `json:"title"`
		Created int64  `json:"created"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	m = &meta{
		Title:   tmp.Title,
		Created: time.Unix(0, tmp.Created),
	}
	return nil
}

var logFilenameRegex = regexp.MustCompile(logFilenameRegexFormat)

func ls() *cobra.Command {
	var ls = &cobra.Command{
		Use:   "ls",
		Short: "List all notes",
		Run:   lsFunc,
	}
	return ls
}

func lsFunc(cmd *cobra.Command, args []string) {
	info, err := ioutil.ReadDir(defaultNotesDirectory)
	if err != nil {
		panic(err)
	}

	for _, f := range info {
		if !logFilenameRegex.MatchString(f.Name()) {
			// TODO: log skipped files for debug
			continue
		}
		if f.Size() < 0 {
			panic(err) // FIXME
		}
		humanSize := humanizeBytesDecimal(uint64(f.Size()))

		fmt.Printf("%s | %s | %s\n", f.Name(), f.ModTime().Format(defaultModTimeFormat), humanSize)
	}
}

// humanizeBytesBinary displays the number of bytes in a human readable format with binary
// unit prefixes
func humanizeBytesBinary(numBytes uint64) string {
	if numBytes < 1024 {
		return fmt.Sprintf("% 7d", numBytes)
	}

	base := uint(bits.Len64(numBytes) / 10)
	val := float64(numBytes) / float64(uint64(1<<(base*10)))

	return fmt.Sprintf("% 7.2f %ciB", val, " KMGTPE"[base])
}

// humanizeBytesDecimal display the number of bytes in a human readable format with decimal
// unit prefixes
func humanizeBytesDecimal(numBytes uint64) string {
	if numBytes < 1000 {
		return fmt.Sprintf("% 7d", numBytes)
	}

	base := uint(bits.Len64(numBytes) / 10)
	val := float64(numBytes) / math.Pow(10, float64(base))

	return fmt.Sprintf("% 7.2f %cB", val, " KMGTPE"[base])
}
