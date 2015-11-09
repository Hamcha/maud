package modules

func WrapImg(url, content string, attribs map[string]string) string {
	s := `<a target='_blank' href='` + url + "'"
	for k, v := range attribs {
		s += " " + k + "='" + v + "'"
	}
	s += ">" + content + "</a>"
	return s
}
