package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
)

type ProxyServer struct {
}

func ReadHttpRequestFromConnection(conn net.Conn) (*http.Request, error) {
	reqFromClient, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return nil, err
	}
	reqToServer, err := http.NewRequest(reqFromClient.Method, reqFromClient.RequestURI, reqFromClient.Body)
	if err != nil {
		return nil, err
	}
	reqToServer.Host = reqFromClient.Host
	reqToServer.URL.Scheme = "http"
	reqToServer.URL.Host = reqFromClient.URL.Host
	reqToServer.Header = reqFromClient.Header
	reqToServer.Header.Del("Proxy-Connection")
	return reqToServer, nil
}

func HandleRequestToServer(reqToServer *http.Request) ([]byte, error) {
	host := reqToServer.Host
	if reqToServer.URL.Port() == "" {
		host = host + ":80"
	}
	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	defer serverConn.Close()
	reqToServerBytes, err := httputil.DumpRequest(reqToServer, true)
	if err != nil {
		return nil, err
	}
	_, err = serverConn.Write(reqToServerBytes)
	if err != nil {
		return nil, err
	}
	responseFromServer := make([]byte, 1024)
	_, err = serverConn.Read(responseFromServer)
	if err != nil {
		return nil, err
	}
	return responseFromServer, nil
}

func HandleClientConnection(clientConn net.Conn) error {
	reqToServer, err := ReadHttpRequestFromConnection(clientConn)
	if err != nil {
		return err
	}
	answerFromServerBytes, err := HandleRequestToServer(reqToServer)
	if err != nil {
		return err
	}
	_, err = clientConn.Write(answerFromServerBytes)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProxyServer) Run() error {
	// Устанавливаем прослушивание порта
	l, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return err
	}
	defer l.Close()
	fmt.Println("listening on localhost:8080")
	// Запускаем цикл
	for {
		// Открываем порт
		clientConn, err := l.Accept()
		if err != nil {
			return err
		}
		defer clientConn.Close()
		go func() {
			err := HandleClientConnection(clientConn)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func main() {
	pServer := ProxyServer{}
	err := pServer.Run()
	if err != nil {
		fmt.Println("server run err = ", err)
		os.Exit(1)
	}
}
