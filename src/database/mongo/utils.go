package mongo

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"strings"
)

func generateURL(db Database, name string) string {
	buf := make([]byte, 8)
	num, _ := db.NextId(name)
	binary.PutVarint(buf, num+1)
	btr := bytes.TrimRight(buf, "\000")
	str := base64.URLEncoding.EncodeToString(btr)
	return strings.TrimRight(str, "=")
}
