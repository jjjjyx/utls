package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	tls "github.com/refraction-networking/utls"
	"github.com/refraction-networking/utls/examples/mu/socks"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type ClientSessionCache struct {
	sessionKeyMap map[string]*tls.ClientSessionState
}

func NewClientSessionCache() tls.ClientSessionCache {
	return &ClientSessionCache{
		sessionKeyMap: make(map[string]*tls.ClientSessionState),
	}
}

func (csc *ClientSessionCache) Get(sessionKey string) (session *tls.ClientSessionState, ok bool) {
	if session, ok = csc.sessionKeyMap[sessionKey]; ok {
		fmt.Printf("Getting session for %s\n", sessionKey)
		return session, true
	}
	fmt.Printf("Missing session for %s\n", sessionKey)
	return nil, false
}

func (csc *ClientSessionCache) Put(sessionKey string, cs *tls.ClientSessionState) {
	if cs == nil {
		fmt.Printf("Deleting session for %s\n", sessionKey)
		delete(csc.sessionKeyMap, sessionKey)
	} else {
		fmt.Printf("Putting session for %s\n", sessionKey)
		csc.sessionKeyMap[sessionKey] = cs
	}
}

var target *url.URL
var dialTimeout = time.Duration(5) * time.Second

var proxy = "192.168.10.10:33051"
var csc tls.ClientSessionCache

func main() {

	csc = tls.NewLRUClientSessionCache(5)

	//target, _ = url.Parse("https://tools.scrapfly.io/api/tls")
	target, _ = url.Parse("https://play.google.com/log")

	save(target.Host, tls.HelloCustom, getHelloNoPSKSpec())
	save(target.Host, tls.HelloCustom, getHelloSpec())
	//save(target.Host, tls.HelloCustom, getHelloSpec())
	save(target.Host, tls.HelloCustom, getHelloNoPSKSpec())
	save(target.Host, tls.HelloCustom, getHelloSpec())
	fmt.Println("end")
}

func save(hostname string, cid tls.ClientHelloID, spec *tls.ClientHelloSpec) {
	var response *http.Response
	var err error

	response, err = HttpGetTicket(hostname, cid, spec)

	if err != nil {
		fmt.Printf("#> HttpGetTicket failed: %+v\n", err)
	} else {
		buf := bytes.NewBuffer(nil)
		dumpResponse, _ := httputil.DumpResponse(response, false)
		buf.Write(dumpResponse)

		buf.WriteString("\n\n")
		buf.WriteString("resp body------------------\n")
		//b, _ := io.ReadAll(response.Body)
		var obj interface{}
		_ = json.NewDecoder(response.Body).Decode(&obj)

		b, _ := json.MarshalIndent(obj, "", "\t")

		buf.Write(b)

		buf.WriteString("\n\nresp body------------------end \n")

		write("001", buf.Bytes())
	}
}

func write(label string, b []byte) {
	filename := label
	filenamebak := filename
	for i := 1; ; i++ {
		targetPath := filepath.Join("./examples/mu/", filename)
		if !fileExists(targetPath) {
			break
		}

		filename = fmt.Sprintf("%s-%d", filenamebak, i)
	}

	_ = os.WriteFile(filepath.Join("./examples/mu/", filename), b, 0644)
}

func getHelloSpec() *tls.ClientHelloSpec {
	raw := "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"
	//raw := "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"
	clientHello, _ := base64.StdEncoding.DecodeString(raw)
	//fmt.Println(len(clientHello))
	var spec tls.ClientHelloSpec

	_ = spec.FromRaw(PrependRecordHeader(clientHello), true, true)

	return &spec
}
func getHelloNoPSKSpec() *tls.ClientHelloSpec {
	//notPSK := "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"
	notPSK := "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"

	clientHello, _ := base64.StdEncoding.DecodeString(notPSK)
	//fmt.Println(len(clientHello))
	var spec tls.ClientHelloSpec

	_ = spec.FromRaw(PrependRecordHeader(clientHello), true, true)

	return &spec
}

//
//func HttpGetTicket(hostname string, addr string) (*http.Response, error) {
//	cid := tls.HelloCustom
//	//cid := tls.HelloChrome_100_PSK
//	dialConn, _ := net.DialTimeout("tcp", proxy, dialTimeout)
//
//	d := socks.NewDialer("tcp", dialConn.RemoteAddr().String())
//	d.DialWithConn(context.Background(), dialConn, "tcp", hostname+":443")
//
//	client := tls.UClient(dialConn, &tls.Config{
//		ServerName:   hostname,
//		OmitEmptyPsk: true,
//		//PreferSkipResumptionOnNilExtension: true,
//		ClientSessionCache: csc,
//	}, cid)
//	spec := getHelloNoPSKSpec()
//	_ = client.ApplyPreset(spec)
//
//	//client.SetSessionTicketExtension(&tls.SessionTicketExtension{
//	//
//	//})
//
//	err := client.Handshake()
//
//	if err != nil {
//		return nil, fmt.Errorf("uTlsConn.Handshake() error: %+v", err)
//	}
//	stat := client.ConnectionState()
//	client.SetReadDeadline(time.Now().Add(1 * time.Second))
//	client.Read(make([]byte, 1024)) // trigger a read so NewSessionTicket gets handled
//	_ = client.Close()
//	fmt.Println("connect stat", stat)
//
//}

func HttpGetTicket(hostname string, cid tls.ClientHelloID, spec *tls.ClientHelloSpec) (*http.Response, error) {
	var err error
	dialConnPSK, _ := net.DialTimeout("tcp", proxy, dialTimeout)

	dpsk := socks.NewDialer("tcp", dialConnPSK.RemoteAddr().String())
	dpsk.DialWithConn(context.Background(), dialConnPSK, "tcp", hostname+":443")

	clientPSK := tls.UClient(dialConnPSK, &tls.Config{
		ServerName:   hostname,
		OmitEmptyPsk: true,
		//PreferSkipResumptionOnNilExtension: true,
		ClientSessionCache: csc,
	}, cid)
	//clientPSK.SetSessionTicketExtension()
	//spec = getHelloSpec()
	_ = clientPSK.ApplyPreset(spec)

	err = clientPSK.Handshake()

	if err != nil {
		return nil, fmt.Errorf("uTlsConn.Handshake() error: %+v", err)
	}
	stat := clientPSK.ConnectionState()
	//os.WriteFile("./examples/mu/xxx", clientPSK.HandshakeState.Hello.Raw, 0644)

	if clientPSK.ConnectionState().HandshakeComplete {
		if true {
			fmt.Println("Handshake complete")
		}
		newVer := clientPSK.ConnectionState().Version
		if true {
			fmt.Printf("TLS Version: %04x\n", newVer)
		}

		if newVer == tls.VersionTLS13 && clientPSK.HandshakeState.State13.UsingPSK {
			fmt.Println("[PSK used]")
		} else {
			fmt.Println("xxxx")
		}
	}
	//clientPSK.HandshakeState.State13

	return httpGetOverConn(clientPSK, stat.NegotiatedProtocol)
}
func httpGetOverConn(conn net.Conn, alpn string) (*http.Response, error) {

	req := &http.Request{
		Method: "GET",
		URL:    target,
		Header: make(http.Header),
		Host:   target.Host,
	}

	switch alpn {
	case "h2":
		req.Proto = "HTTP/2.0"
		req.ProtoMajor = 2
		req.ProtoMinor = 0

		tr := http2.Transport{}
		cConn, err := tr.NewClientConn(conn)
		if err != nil {
			return nil, err
		}
		return cConn.RoundTrip(req)
	case "http/1.1", "":
		req.Proto = "HTTP/1.1"
		req.ProtoMajor = 1
		req.ProtoMinor = 1

		err := req.Write(conn)
		if err != nil {
			return nil, err
		}
		return http.ReadResponse(bufio.NewReader(conn), req)
	default:
		return nil, fmt.Errorf("unsupported ALPN: %v", alpn)
	}
}

func dumpResponseNoBody(response *http.Response) string {
	resp, err := httputil.DumpResponse(response, false)
	if err != nil {
		return fmt.Sprintf("failed to dump response: %v", err)
	}
	return string(resp)
}

// PrependRecordHeader prepends a record header to a handshake messsage
// if attempting to mimic an existing connection the minTLSVersion can be found
// in the Conn.vers field
func PrependRecordHeader(hello []byte) []byte {
	l := len(hello)

	// utls.VersionTLS10 这个无所谓的， 最后会被 实际 raw 中的 扩展值修正
	minTLSVersion := tls.VersionTLS10
	recordType := 22
	//22 == recordTypeHandshake
	header := []byte{
		uint8(recordType),                                             // type
		uint8(minTLSVersion >> 8 & 0xff), uint8(minTLSVersion & 0xff), // record version is the minimum supported
		uint8(l >> 8 & 0xff), uint8(l & 0xff), // length
	}
	return append(header, hello...)
}

func fileExists(path string) bool {
	fileStat, err := os.Stat(path)
	return (err == nil || os.IsExist(err)) && !fileStat.IsDir()
}
