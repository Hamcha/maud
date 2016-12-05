package bbcode

import (
	"os"
	"testing"

	"github.com/hamcha/maud/maud/modules"
)

var bb modules.Formatter

func doWrap(tag string) func(string) string {
	return func(s string) string {
		return "[" + tag + "]" + s + "[/" + tag + "]"
	}
}

func doWrapParam(tag, par string) func(string) string {
	return func(s string) string {
		return "[" + tag + "=" + par + "]" + s + "[/" + tag + "]"
	}
}

func doWrapFreeParam(tag string) func(string, string) string {
	return func(s, par string) string {
		return "[" + tag + "=" + par + "]" + s + "[/" + tag + "]"
	}
}

func runTests(t *testing.T, tests map[string]string) {
	for in, out := range tests {
		if res := bb.Format(in); res != out {
			t.Error("\nOn input: ", in,
				"\nexpected: ", out,
				"\ngot:      ", res)
		}
	}
}

func TestMain(m *testing.M) {
	bb = Provide()
	os.Exit(m.Run())
}

func TestYoutube(t *testing.T) {
	const code = "12345678901"
	out := func(code, starttime string) string {
		s := `<iframe width="560" height="315" src="//www.youtube.com/embed/` + code
		if len(starttime) > 0 {
			s += "?start=" + starttime
		}
		s += `" frameborder="0" allowfullscreen></iframe>`
		return s
	}
	wrap := doWrap("youtube")
	tests := map[string]string{
		// Full link
		wrap(`https://www.youtube.com/watch?v=` + code):              out(code, ""),
		wrap(`youtube.com/watch?v=` + code):                          out(code, ""),
		wrap(`http://youtube.com/watch?v=` + code):                   out(code, ""),
		wrap(`https://www.youtube.com/watch?v=` + code + `&t=1m42s`): out(code, "102"),
		wrap(`https://www.youtube.com/watch?v=` + code + `&t=142s`):  out(code, "142"),
		wrap(`youtube.com/watch?v=` + code + `&t=2m`):                out(code, "120"),
		wrap(`http://youtube.com/watch?v=` + code + `&t=0m22s`):      out(code, "22"),
		// Short link
		wrap(`https://youtu.be/` + code):              out(code, ""),
		wrap(`youtu.be/` + code):                      out(code, ""),
		wrap(`https://youtu.be/` + code + `?t=1m42s`): out(code, "102"),
		wrap(`https://youtu.be/` + code + `?t=142s`):  out(code, "142"),
		wrap(`youtu.be/` + code + `?t=1m42s`):         out(code, "102"),
		// Video code
		wrap(code): out(code, ""),
	}

	runTests(t, tests)
}

func TestUrl(t *testing.T) {
	wrap := doWrap("url")
	wrapPar := doWrapFreeParam("url")
	out := func(par, con string) string {
		return `<a href="` + par + `" rel="nofollow">` + con + `</a>`
	}
	tests := map[string]string{
		wrap(`http://www.example.com`): out(`http://www.example.com`, `http://www.example.com`),
		wrap(`example.com`):            out(`http://example.com`, `example.com`),
		wrap(`//example.com`):          out(`//example.com`, `//example.com`),
		wrap(`http://www.example.com?foo=bar&baz=quz`): out(`http://www.example.com?foo=bar&baz=quz`,
			`http://www.example.com?foo=bar&baz=quz`),
		wrapPar("Example", `http://www.example.com`): out(`http://www.example.com`, "Example"),
		wrapPar("Example", `//www.example.com`):      out(`//www.example.com`, "Example"),
		wrapPar("Example", `www.example.com`):        out(`http://www.example.com`, "Example"),
	}

	runTests(t, tests)
}

func TestVideo(t *testing.T) {
	wrap := doWrap("video")
	wrapGif := doWrapParam("video", "gif")
	outWebm := func(con, opts string) string {
		return `<video height="250" src="` + con + `" ` + opts +
			`>[Your browser is unable to play this video]</video>`
	}
	outOth := func(con, opts, ext string) string {
		return `<video height="250" ` + opts + `><source src="` + con + `" type="video/` +
			ext + `"/>[Your browser is unable to play this video]</video>`
	}
	const (
		defaultOpts = "controls"
		gifOpts     = "autoplay muted loop"
	)
	tests := map[string]string{
		wrap(`https://www.example.com/asd.webm`):    outWebm(`https://www.example.com/asd.webm`, defaultOpts),
		wrap(`//www.example.com/asd.webm`):          outWebm(`//www.example.com/asd.webm`, defaultOpts),
		wrap(`www.example.com/asd.webm`):            outWebm(`www.example.com/asd.webm`, defaultOpts),
		wrap(`https://www.example.com/asd.gifv`):    outWebm(`https://www.example.com/asd.webm`, defaultOpts),
		wrap(`//www.example.com/asd.gifv`):          outWebm(`//www.example.com/asd.webm`, defaultOpts),
		wrap(`www.example.com/asd.gifv`):            outWebm(`www.example.com/asd.webm`, defaultOpts),
		wrap(`https://www.example.com/asd.ogg`):     outOth(`https://www.example.com/asd.ogg`, defaultOpts, "ogg"),
		wrap(`//www.example.com/asd.ogg`):           outOth(`//www.example.com/asd.ogg`, defaultOpts, "ogg"),
		wrap(`www.example.com/asd.ogg`):             outOth(`www.example.com/asd.ogg`, defaultOpts, "ogg"),
		wrap(`https://www.example.com/asd.ogv`):     outOth(`https://www.example.com/asd.ogv`, defaultOpts, "ogv"),
		wrap(`https://www.example.com/asd.mp4`):     outOth(`https://www.example.com/asd.mp4`, defaultOpts, "mp4"),
		wrapGif(`https://www.example.com/asd.webm`): outWebm(`https://www.example.com/asd.webm`, gifOpts),
		wrapGif(`//www.example.com/asd.webm`):       outWebm(`//www.example.com/asd.webm`, gifOpts),
		wrapGif(`www.example.com/asd.webm`):         outWebm(`www.example.com/asd.webm`, gifOpts),
		wrapGif(`https://www.example.com/asd.gifv`): outWebm(`https://www.example.com/asd.webm`, gifOpts),
		wrapGif(`https://www.example.com/asd.ogg`):  outOth(`https://www.example.com/asd.ogg`, gifOpts, "ogg"),
		wrapGif(`//www.example.com/asd.ogg`):        outOth(`//www.example.com/asd.ogg`, gifOpts, "ogg"),
		wrapGif(`www.example.com/asd.ogg`):          outOth(`www.example.com/asd.ogg`, gifOpts, "ogg"),
		wrapGif(`https://www.example.com/asd.ogv`):  outOth(`https://www.example.com/asd.ogv`, gifOpts, "ogv"),
		wrapGif(`https://www.example.com/asd.mp4`):  outOth(`https://www.example.com/asd.mp4`, gifOpts, "mp4"),
	}

	runTests(t, tests)
}
