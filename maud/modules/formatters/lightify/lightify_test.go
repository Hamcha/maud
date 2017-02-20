package lightify

import (
	"testing"

	"github.com/hamcha/maud/maud/data"
	"github.com/hamcha/maud/maud/modules"
)

const (
	inNonImg     = `This text <strong>contains <em>no</em> images</strong>.`
	fNonImg      = inNonImg
	rtNonImg     = inNonImg
	inImg        = `This text contains an image: <img src="http://example.com/foo.png" title="foo" />`
	fImg         = inImg
	rtImg        = `This text contains an image: <a class='toggleImage' target='_blank' href="http://example.com/foo.png">[Click to view image]</a>`
	inImgur      = `This image is from Imgur: <img src='http://i.imgur.com/abcdefg.jpg'/>`
	fImgur       = `This image is from Imgur: <a target="_blank" href='http://i.imgur.com/abcdefg.jpg'><img src="http://i.imgur.com/abcdefgm.jpg" alt="http://i.imgur.com/abcdefg.jpg"/></a>`
	rtImgur      = inImgur
	inImgurThumb = `This image is from Imgur: <img src='http://i.imgur.com/abcdefgs.jpg'/>`
	fImgurThumb  = `This image is from Imgur: <a target="_blank" href='http://i.imgur.com/abcdefgs.jpg'><img src="http://i.imgur.com/abcdefgs.jpg" alt="http://i.imgur.com/abcdefgs.jpg"/></a>`
	rtImgurThumb = inImgurThumb
	inImgurGifv  = `This image is from Imgur: <img src='http://i.imgur.com/abcdefg.gifv'/>`
	fImgurGifv   = `This image is from Imgur: <a target="_blank" href='http://i.imgur.com/abcdefg.gifv'><video height="250" src="http://i.imgur.com/abcdefg.webm" autoplay loop muted>[Your browser is unable to play this video]</video></a>`
	rtImgurGifv  = inImgurGifv
	inDerpi      = `This image is from Derpibooru: <img src='https://derpicdn.net/img/2017/2/20/1368311/medium.jpeg'/>`
	fDerpi       = `This image is from Derpibooru: <a target="_blank" href='https://derpicdn.net/img/2017/2/20/1368311/medium.jpeg'><img src='https://derpicdn.net/img/2017/2/20/1368311/thumb.jpeg' alt='https://derpicdn.net/img/2017/2/20/1368311/medium.jpeg'/></a>`
	rtDerpi      = inDerpi
	inIframe     = `This text contains an iframe <iframe  src='http://example.com/'>`
	fIframe      = inIframe
	rtIframe     = `This text contains an iframe <a target='_blank' href='http://example.com/'>[Embedded: 'http://example.com/']</a><br />`
	inYoutube    = `This text contains a youtube video <iframe src='http://youtube.com/embed/abcdefghjkl'>`
	fYoutube     = inYoutube
	rtYoutube    = `This text contains a youtube video <a target='_blank' href='http://youtube.com/watch?v=abcdefghjkl'>[YouTube: 'http://youtube.com/watch?v=abcdefghjkl']</a><br />`
)

var ltf *LightifyFormatter

func init() {
	ltf = Provide()
}

func shouldFormat(t *testing.T, in, out string) {
	if result := ltf.Format(in); result != out {
		t.Error("Format(): for input\r\n\r\n    ", in, "\r\n\r\nexpected\r\n\r\n    ", out,
			"\r\n\r\nbut got\r\n\r\n    ", result)
	}
}

func shouldReplaceTags(t *testing.T, in, out string) {
	var pmd modules.PostMutatorData
	pmd.Post = &data.Post{
		Content: in,
	}
	if ltf.ReplaceTags(pmd); pmd.Post.Content != out {
		t.Error("ReplaceTags(): for input\r\n\r\n    ", in, "\r\n\r\nexpected\r\n\r\n    ", out,
			"\r\n\r\nbut got\r\n\r\n    ", pmd.Post.Content)
	}
}

func TestImg(t *testing.T) {
	shouldFormat(t, inNonImg, fNonImg)
	shouldReplaceTags(t, inNonImg, rtNonImg)
	shouldFormat(t, inImg, fImg)
	shouldReplaceTags(t, inImg, rtImg)
	shouldFormat(t, inImgur, fImgur)
	shouldReplaceTags(t, inImgur, rtImgur)
	shouldFormat(t, inImgurThumb, fImgurThumb)
	shouldReplaceTags(t, inImgurThumb, rtImgurThumb)
	shouldFormat(t, inImgurGifv, fImgurGifv)
	shouldReplaceTags(t, inImgurGifv, rtImgurGifv)
	shouldFormat(t, inDerpi, fDerpi)
	shouldReplaceTags(t, inDerpi, rtDerpi)
	shouldFormat(t, inIframe, fIframe)
	shouldReplaceTags(t, inIframe, rtIframe)
	shouldFormat(t, inYoutube, fYoutube)
	shouldReplaceTags(t, inYoutube, rtYoutube)
}
