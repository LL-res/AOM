package socket

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/uds_socket")

	defer func() {
		time.Sleep(5 * time.Second)
		conn.Close()
	}()
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write([]byte("hello server\n")); err != nil {
		log.Print(err)
		return
	}
	//var buf = make([]byte, 1)
	bs, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println(err)
	}
	//if _, err := conn.Read(buf); err != nil {
	//	panic(err)
	//}
	fmt.Println("client recv: ", string(bs))
}

func New(address string) {

}
