package local_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
)

func setup() (string, func() error, error) {
	done := func() error { return nil } // noop
	dir, err := ioutil.TempDir("testdata", "stow")
	if err != nil {
		return dir, done, err
	}
	done = func() error {
		return os.RemoveAll(dir)
	}
	// add some "containers"
	err = os.Mkdir(filepath.Join(dir, "one"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = os.Mkdir(filepath.Join(dir, "two"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = os.Mkdir(filepath.Join(dir, "three"), 0777)
	if err != nil {
		return dir, done, err
	}

	// add three items
	err = ioutil.WriteFile(filepath.Join(dir, "three", "item1"), []byte("3.1"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "three", "item2"), []byte("3.2"), 0777)
	if err != nil {
		return dir, done, err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "three", "item3"), []byte("3.3"), 0777)
	if err != nil {
		return dir, done, err
	}

	// make testpath absolute
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return dir, done, err
	}
	return absdir, done, nil
}

func TestCreateItem(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}
	l, err := stow.Dial(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	containers, cursor, err := l.Containers("t", stow.CursorStart)
	is.NoErr(err)
	is.OK(containers)
	c1 := containers[0]
	items, cursor, err := c1.Items("", stow.CursorStart)
	is.NoErr(err)
	is.Equal(cursor, "")
	beforecount := len(items)

	content := "new item contents"
	newitem, err := c1.Put("new_item", strings.NewReader(content), int64(len(content)))
	is.NoErr(err)
	is.OK(newitem)
	is.Equal(newitem.Name(), "new_item")

	// get the container again
	containers, cursor, err = l.Containers("t", stow.CursorStart)
	is.NoErr(err)
	is.OK(containers)
	is.Equal(cursor, "")
	c1 = containers[0]
	items, cursor, err = c1.Items("", stow.CursorStart)
	is.NoErr(err)
	is.Equal(cursor, "")
	aftercount := len(items)

	is.Equal(aftercount, beforecount+1)

	// get new item
	item := items[len(items)-1]
	md5, err := item.MD5()
	is.NoErr(err)
	is.Equal(md5, "1d4b28e33c8bfcfdb75e116ed2319632")
	etag, err := item.ETag()
	is.NoErr(err)
	is.OK(etag)
	r, err := item.Open()
	is.NoErr(err)
	defer r.Close()
	itemContents, err := ioutil.ReadAll(r)
	is.NoErr(err)
	is.Equal("new item contents", string(itemContents))

}

func TestItems(t *testing.T) {
	is := is.New(t)
	testDir, teardown, err := setup()
	is.NoErr(err)
	defer teardown()

	cfg := stow.ConfigMap{"path": testDir}

	l, err := stow.Dial(local.Kind, cfg)
	is.NoErr(err)
	is.OK(l)

	containers, cursor, err := l.Containers("t", stow.CursorStart)
	is.NoErr(err)
	is.OK(containers)
	is.Equal(cursor, "")
	three, err := l.Container(containers[0].ID())
	is.NoErr(err)
	items, cursor, err := three.Items("", stow.CursorStart)
	is.NoErr(err)
	is.OK(items)
	is.Equal(cursor, "")

	is.Equal(len(items), 3)
	is.Equal(items[0].ID(), filepath.Join(containers[0].ID(), "item1"))
	is.Equal(items[0].Name(), "item1")
}

func isDir(is is.I, path string) {
	info, err := os.Stat(path)
	is.NoErr(err)
	is.True(info.IsDir())
}