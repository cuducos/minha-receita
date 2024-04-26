package mirror

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const unit = 1024

type File struct {
	URL            string `json:"url"`
	Size           int64  `json:"size"`
	name           string
	lastModifiedAt time.Time
}

func (f *File) HumanReadableSize() string {
	if f.Size < unit {
		return fmt.Sprintf("%d B", f.Size)
	}
	div, exp := int64(unit), 0
	for n := f.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(f.Size)/float64(div), "KMGTPE"[exp])
}

func (f *File) ShortName() string {
	p := strings.Split(f.name, "/")
	return p[len(p)-1]
}

func (f *File) group() string {
	p := strings.Split(f.name, "/")
	if len(p) == 1 {
		return "Binários"
	}
	return p[0]
}

type Group struct {
	Name  string `json:"name"`
	Files []File `json:"urls"`
}

func newGroups(fs []File) []Group {
	var m = make(map[string][]File)
	for _, f := range fs {
		n := f.group()
		m[n] = append(m[n], f)
	}
	ks := []string{}
	for k := range m {
		ks = append(ks, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ks)))
	var gs []Group
	for _, k := range ks {
		gs = append(gs, Group{k, m[k]})
	}
	return gs
}
