package pkg

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	urlutil "github.com/projectdiscovery/utils/url"
	"github.com/secoba/rawhttp/client"
)

// StatusError is a HTTP status error object
type StatusError struct {
	client.Status
}

func (s *StatusError) Error() string {
	return s.Status.String()
}

type readCloser struct {
	io.Reader
	io.Closer
}

func toRequest(method string, host, path string, query []string,
	headers map[string][]string, body io.Reader, raw []byte, options *Options) *client.Request {
	if headers == nil {
		headers = make(map[string][]string, 0)
	}
	var (
		rawBuffer []byte
		version   = client.HTTP_1_1
	)

	if len(options.CustomHeaders) > 0 {
		for _, header := range options.CustomHeaders {
			headers[header.Key] = []string{header.Value}
		}
	}
	reqHeaders := toHeaders(headers)

	if raw != nil && len(raw) > 0 {
		seperator := "\n"
		if bytes.Contains(raw, []byte("\r\n")) {
			seperator = "\r\n"
		}
		multiSeperator := append([]byte(seperator), []byte(seperator)...)

		var (
			hasHost   bool
			hasBody   bool
			hasLength bool
		)

		bufferArr := bytes.SplitN(raw, multiSeperator, 2)
		if len(bufferArr) == 2 {
			hasBody = true
		}

		buffer := new(bytes.Buffer)

		prePkg := bytes.Split(bufferArr[0], []byte(seperator))
		for i := 0; i < len(prePkg); i++ {
			buffer.Write(prePkg[i])
			buffer.WriteString("\r\n")
		}
		for _, header := range reqHeaders {
			buffer.WriteString(fmt.Sprintf("%s: %s", header.Key, header.Value))
			buffer.WriteString("\r\n")
		}

		if options.AutomaticHostHeader {
			reg := regexp.MustCompile("(?i)(host:\\s*(.*))")
			ret := reg.FindString(buffer.String())
			if len(ret) > 0 {
				hasHost = true
				buff := bytes.ReplaceAll(buffer.Bytes(), []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Host: %s", host)))
				buffer.Reset()
				buffer.Write(buff)
				// bufferArr[0] = bytes.ReplaceAll(bufferArr[0], []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Host: %s", host)))
			}
		}

		if options.AutomaticContentLength {
			reg := regexp.MustCompile("(?i)(content-length:\\s+(\\d+))")
			ret := reg.FindString(buffer.String())
			if len(ret) > 0 {
				hasLength = true
				if hasBody {
					buff := bytes.ReplaceAll(buffer.Bytes(), []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Content-Length: %d", len(bufferArr[1]))))
					buffer.Reset()
					buffer.Write(buff)
					// bufferArr[0] = bytes.ReplaceAll(bufferArr[0], []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Content-Length: %d", len(bufferArr[1]))))
				}
			}
		}

		if !hasHost {
			buffer.WriteString(fmt.Sprintf("Host: %s", host))
			buffer.WriteString("\r\n")
		}
		if !hasLength && hasBody {
			buffer.WriteString(fmt.Sprintf("Content-Length: %d", len(bufferArr[1])))
			buffer.WriteString("\r\n")
		}

		buffer.WriteString("\r\n")

		// body
		if hasBody {
			buffer.Write(bufferArr[1])
		}

		rawBuffer = buffer.Bytes()

		firstLine := bytes.Split(raw, []byte(seperator))[0]
		headerLine := strings.SplitN(string(firstLine), " ", 3)
		if len(headerLine) == 3 {
			protoMajor, protoMinor, _ := parseHttpVersion(headerLine[2])
			version = client.Version{
				Major: protoMajor,
				Minor: protoMinor,
			}
		}
	} else if len(options.CustomRawBytes) > 0 {
		seperator := "\n"
		if bytes.Contains(options.CustomRawBytes, []byte("\r\n")) {
			seperator = "\r\n"
		}
		multiSeperator := append([]byte(seperator), []byte(seperator)...)

		var (
			hasHost   bool
			hasBody   bool
			hasLength bool
		)

		bufferArr := bytes.SplitN(options.CustomRawBytes, multiSeperator, 2)
		if len(bufferArr) == 2 {
			hasBody = true
		}

		buffer := new(bytes.Buffer)

		// pkg prefix
		prePkg := bytes.Split(bufferArr[0], []byte(seperator))
		for i := 0; i < len(prePkg); i++ {
			buffer.Write(prePkg[i])
			buffer.WriteString("\r\n")
		}
		for _, header := range reqHeaders {
			buffer.WriteString(fmt.Sprintf("%s: %s", header.Key, header.Value))
			buffer.WriteString("\r\n")
		}

		if options.AutomaticHostHeader {
			reg := regexp.MustCompile("(?i)(host:\\s*(.*))")
			ret := reg.FindString(string(bufferArr[0]))
			if len(ret) > 0 {
				hasHost = true
				buff := bytes.ReplaceAll(buffer.Bytes(), []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Host: %s", host)))
				buffer.Reset()
				buffer.Write(buff)
				//bufferArr[0] = bytes.ReplaceAll(bufferArr[0], []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Host: %s", host)))
			}
		}

		if options.AutomaticContentLength {
			reg := regexp.MustCompile("(?i)(content-length:\\s+(\\d+))")
			ret := reg.FindString(string(bufferArr[0]))
			if len(ret) > 0 {
				hasLength = true
				if hasBody {
					buff := bytes.ReplaceAll(buffer.Bytes(), []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Content-Length: %d", len(bufferArr[1]))))
					buffer.Reset()
					buffer.Write(buff)
					//bufferArr[0] = bytes.ReplaceAll(bufferArr[0], []byte(strings.TrimSpace(ret)), []byte(fmt.Sprintf("Content-Length: %d", len(bufferArr[1]))))
				}
			}
		}

		if !hasHost {
			buffer.WriteString(fmt.Sprintf("Host: %sf", host))
			buffer.WriteString("\r\n")
		}
		if !hasLength && hasBody {
			buffer.WriteString(fmt.Sprintf("Content-Length: %d", len(bufferArr[1])))
			buffer.WriteString("\r\n")
		}

		buffer.WriteString("\r\n")

		// body
		if hasBody {
			buffer.Write(bufferArr[1])
			//sufPkg := bytes.Split(bufferArr[1], []byte(seperator))
			//for i := 0; i < len(sufPkg); i++ {
			//	buffer.Write(sufPkg[i])
			//	buffer.WriteString("\r\n")
			//}
		}

		rawBuffer = buffer.Bytes()
		//options.CustomRawBytes = buffer.Bytes()

		firstLine := bytes.Split(options.CustomRawBytes, []byte(seperator))[0]
		headerLine := strings.SplitN(string(firstLine), " ", 3)
		if len(headerLine) == 3 {
			protoMajor, protoMinor, _ := parseHttpVersion(headerLine[2])
			version = client.Version{
				Major: protoMajor,
				Minor: protoMinor,
			}
		}
	}

	return &client.Request{
		Method:   method,
		Path:     path,
		Query:    query,
		Version:  version,
		Headers:  reqHeaders,
		Body:     body,
		RawBytes: rawBuffer,
	}
}

func toHTTPResponse(conn Conn, resp *client.Response) (*http.Response, error) {
	rheaders := fromHeaders(resp.Headers)
	r := http.Response{
		ProtoMinor:    resp.Version.Minor,
		ProtoMajor:    resp.Version.Major,
		Status:        resp.Status.String(),
		StatusCode:    resp.Status.Code,
		Header:        rheaders,
		ContentLength: resp.ContentLength(),
	}

	var err error
	rbody := resp.Body
	if headerValue(rheaders, "Content-Encoding") == "gzip" {
		rbody, err = gzip.NewReader(rbody)
		if err != nil {
			return nil, err
		}
	}
	rc := &readCloser{rbody, conn}

	r.Body = rc

	return &r, nil
}

func toHeaders(h map[string][]string) []client.Header {
	var r []client.Header
	for k, v := range h {
		for _, vv := range v {
			r = append(r, client.Header{Key: k, Value: vv})
		}
	}
	return r
}

func fromHeaders(h []client.Header) map[string][]string {
	if h == nil {
		return nil
	}
	var r = make(map[string][]string)
	for _, hh := range h {
		r[hh.Key] = append(r[hh.Key], hh.Value)
	}
	return r
}

func headerValue(headers map[string][]string, key string) string {
	return strings.Join(headers[key], " ")
}

func firstErr(err1, err2 error) error {
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func parseHttpVersion(vers string) (major, minor int, ok bool) {
	if !strings.HasPrefix(vers, "HTTP/") {
		return 0, 0, false
	}
	if len(vers) != len("HTTP/X.Y") {
		return 0, 0, false
	}
	if vers[6] != '.' {
		return 0, 0, false
	}
	maj, err := strconv.ParseInt(vers[5:6], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	min, err := strconv.ParseInt(vers[7:8], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return int(maj), int(min), true
}

// DumpRequestRaw to string
func DumpRequestRaw(method, url, uripath string, headers map[string][]string, body io.Reader, rawBuffer []byte, options *Options) ([]byte, error) {
	if len(options.CustomRawBytes) > 0 {
		return options.CustomRawBytes, nil
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	u, err := urlutil.ParseURL(url, true)
	if err != nil {
		return nil, err
	}

	// Handle only if host header is missing
	_, hasHostHeader := headers["Host"]
	if !hasHostHeader {
		host := u.Host
		headers["Host"] = []string{host}
	}

	// standard path
	path := u.Path
	if path == "" {
		path = "/"
	}
	if !u.Params.IsEmpty() {
		path += "?" + u.Params.Encode()
	}
	// override if custom one is specified
	if uripath != "" {
		path = uripath
	}

	req := toRequest(method, u.Host, path, nil, headers, body, rawBuffer, options)
	//b := strings.Builder{}
	b := new(bytes.Buffer)

	q := strings.Join(req.Query, "&")
	if len(q) > 0 {
		q = "?" + q
	}

	b.WriteString(fmt.Sprintf("%s %s%s %s"+client.NewLine, req.Method, req.Path, q, req.Version.String()))

	for _, header := range req.Headers {
		if header.Value != "" {
			b.WriteString(fmt.Sprintf("%s: %s"+client.NewLine, header.Key, header.Value))
		} else {
			b.WriteString(fmt.Sprintf("%s"+client.NewLine, header.Key))
		}
	}

	l := req.ContentLength()
	if req.AutomaticContentLength && l >= 0 {
		b.WriteString(fmt.Sprintf("Content-Length: %d"+client.NewLine, l))
	}

	b.WriteString(client.NewLine)

	if req.Body != nil {
		var buf bytes.Buffer
		tee := io.TeeReader(req.Body, &buf)
		bd, e := io.ReadAll(tee)
		if e != nil {
			return nil, e
		}
		b.Write(bd)
	}

	//return []byte(strings.ReplaceAll(b.String(), "\n", client.NewLine)), nil
	return b.Bytes(), nil //[]byte(strings.ReplaceAll(b.String(), "\n", client.NewLine)), nil
}
