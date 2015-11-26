package proxy

import (
	"errors"
	"github.com/bamiaux/rez"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type Proxy struct {
	Root      string
	MaxWidth  int
	MaxHeight int
	client    http.Client
}

type ImageData struct {
	Width  int
	Height int
	Size   int64
}

type AsyncContentResult struct {
	Path  string
	Error error
}

type AsyncImageResult struct {
	Path  string
	Data  ImageData
	Error error
}

var ErrFormatNotSupported = errors.New("Format not supported")

// GetContent retrieves a content from a remote server and returns
// the relative path to it
// If the content already exists then it just returns the path to it
func (p Proxy) GetContent(contentURL string) (string, error) {
	contentPath := getPathURL(contentURL)
	ospath := p.Root + filepath.FromSlash(contentPath)

	// Check if file exists
	_, err := os.Stat(ospath)

	// If the file already exists then just return it
	if err == nil {
		return contentPath, nil
	}

	// If file does not exist then fetch it
	if os.IsNotExist(err) {
		err = p.Fetch(contentURL)
	}

	return contentPath, err
}

// GetImage retrieves an image from a remote server, makes it a thumbnail
// and returs the relative path to it in addition to some metadata
// If the thumbnail already exists then it just returns its path and metadata
func (p Proxy) GetImage(contentURL string) (string, ImageData, error) {
	data := ImageData{}

	// Get static path
	contentPath := getPathURL(contentURL)
	ospath := p.Root + filepath.FromSlash(contentPath)

	// Check if file exists
	stat, err := os.Stat(ospath)

	// If the file already exists or has been just fetched then
	// get its metadata and return it
	if err == nil {

		file, err := os.Open(ospath)
		if err != nil {
			return contentPath, data, err
		}
		defer file.Close()

		img, _, err := image.DecodeConfig(file)
		if err != nil {
			return contentPath, data, err
		}

		// Assign image data
		data.Width = img.Width
		data.Height = img.Height
		data.Size = stat.Size()

		// Check that width/height are within limits and resize if necessary
		var ratio float32
		if data.Width > p.MaxWidth {
			ratio = float32(data.Width) / float32(p.MaxWidth)
			data.Width = p.MaxWidth
			data.Height = int(float32(data.Height) / ratio)
		}
		if data.Height > p.MaxHeight {
			ratio = float32(data.Height) / float32(p.MaxHeight)
			data.Height = p.MaxHeight
			data.Width = int(float32(data.Width) / ratio)
		}

		return contentPath, data, nil
	}

	// If image does not exist then fetch it and make a thumbnail
	if os.IsNotExist(err) {
		data, err = p.FetchThumb(contentURL)

		// If the format is not supported, fetch as standard file
		if err == ErrFormatNotSupported {
			err = p.Fetch(contentURL)
			return contentPath, ImageData{}, err
		}
	}

	return contentPath, data, err
}

// Fetch retreives the resources at `contentURL` and saves it
// under p.Root/domain/path/to/file.
func (p Proxy) Fetch(contentURL string) error {
	resp, err := p.client.Get(contentURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Get static path
	path := getPathURL(contentURL)
	ospath := p.Root + filepath.FromSlash(path)

	// Create the directory tree
	err = os.MkdirAll(filepath.Dir(ospath), 0701)
	if err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(ospath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// FetchThumb retreives the resources at `contentURL`, makes it a thumbnail
// and saves it under p.Root/domain/path/to/file.
// It returns either the file metadata or an error
func (p Proxy) FetchThumb(contentURL string) (ImageData, error) {
	data := ImageData{}

	// Request image
	resp, err := p.client.Get(contentURL)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	// Check content type to see if it is a supported format
	switch resp.Header.Get("Content-Type") {
	case "image/png", "image/jpeg":
		break
	default:
		return data, ErrFormatNotSupported
	}

	// Get image data
	srcimg, _, err := image.Decode(resp.Body)
	if err != nil {
		return data, err
	}

	// Scale image down
	size := srcimg.Bounds().Size()
	var ratio float32
	if size.X > p.MaxWidth {
		ratio = float32(size.X) / float32(p.MaxWidth)
		size.X = p.MaxWidth
		size.Y = int(float32(size.Y) / ratio)
	}
	if size.Y > p.MaxHeight {
		ratio = float32(size.Y) / float32(p.MaxHeight)
		size.Y = p.MaxHeight
		size.X = int(float32(size.X) / ratio)
	}

	data.Width = size.X
	data.Height = size.Y

	// Create target image and resize
	resizerect := image.Rectangle{image.ZP, size}
	resizeimg := provideImg(srcimg, resizerect)

	err = rez.Convert(resizeimg, srcimg, rez.NewBilinearFilter())
	if err != nil {
		return data, err
	}

	// Get static path
	path := getPathURL(contentURL)
	ospath := p.Root + filepath.FromSlash(path)

	// Create the directory tree
	err = os.MkdirAll(filepath.Dir(ospath), 0701)
	if err != nil {
		return data, err
	}

	// Create the file
	file, err := os.Create(ospath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	// Save as jpeg or png depending on color model
	switch resizeimg.ColorModel() {
	case color.YCbCrModel:
		err = jpeg.Encode(file, resizeimg, &jpeg.Options{
			Quality: 100,
		})
	case color.RGBAModel:
		fallthrough
	case color.RGBA64Model:
		fallthrough
	case color.AlphaModel:
		fallthrough
	case color.Alpha16Model:
		fallthrough
	default:
		err = png.Encode(file, resizeimg)
	}

	return data, err
}

func getPathURL(rawURL string) string {
	urldata, _ := url.Parse(rawURL)
	return "/" + urldata.Host + urldata.Path
}

func (p Proxy) GetContentAsync(contentURL string) <-chan AsyncContentResult {
	ch := make(chan AsyncContentResult)

	go func() {
		url, err := p.GetContent(contentURL)
		ch <- AsyncContentResult{
			Path:  url,
			Error: err,
		}
	}()

	return ch
}

func (p Proxy) GetImageAsync(contentURL string) <-chan AsyncImageResult {
	ch := make(chan AsyncImageResult)

	go func() {
		url, data, err := p.GetImage(contentURL)
		ch <- AsyncImageResult{
			Path:  url,
			Data:  data,
			Error: err,
		}
	}()

	return ch
}

func provideImg(src image.Image, rect image.Rectangle) image.Image {
	switch src.ColorModel() {
	case color.NRGBAModel:
		return image.NewNRGBA(rect)
	case color.NRGBA64Model:
		return image.NewNRGBA64(rect)
	case color.RGBAModel:
		return image.NewRGBA(rect)
	case color.RGBA64Model:
		return image.NewRGBA64(rect)
	case color.YCbCrModel:
		return image.NewYCbCr(rect, image.YCbCrSubsampleRatio420)
	case color.AlphaModel:
		return image.NewAlpha(rect)
	case color.Alpha16Model:
		return image.NewAlpha16(rect)
	case color.GrayModel:
		return image.NewGray(rect)
	case color.Gray16Model:
		return image.NewGray16(rect)
	default:
		// Convert source to RGB, might be slow
		bounds := src.Bounds()
		newsrc := image.NewRGBA(bounds)
		draw.Draw(newsrc, bounds, src, image.ZP, draw.Src)
		src = newsrc
		return image.NewRGBA(rect)
	}
}
