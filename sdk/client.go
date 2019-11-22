package sdk

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/king526/goupdater/sdk/pb"
	"google.golang.org/grpc"
)

var (
	noCtx = context.Background()
	null  = &pb.Null{}
)

type ServerEntity struct {
	Addr string
	Tag  []string
}

func GetVersion(s ServerEntity) (*pb.VersionRsp, error) {
	c, e := newClient(s.Addr)
	if e != nil {
		return nil, e
	}
	return c.Version(noCtx, null)
}

func Update(s ServerEntity, path, tag string) error {
	c, err := newClient(s.Addr)
	if err != nil {
		return err
	}
	stream, err := c.Update(noCtx)
	if err != nil {
		return err
	}
	fs, err := os.Open(path)
	if err != nil {
		return err
	}
	fi, _ := fs.Stat()
	if tag == "" {
		tag = fi.ModTime().Format("20060102-150405.999")
	}
	req := &pb.UpdateReq{
		Tag:  tag,
		Data: nil,
	}
	bytes := make([]byte, 1024*512)
	var read int
	for err == nil {
		if read, err = fs.Read(bytes); err == nil {
			req.Data = bytes[:read]
			err = stream.Send(req)
		}
	}
	_, err = stream.CloseAndRecv()
	if err == io.EOF {
		err = nil
	}
	return err
}

func UploadFile(s ServerEntity, path string) error {
	c, err := newClient(s.Addr)
	if err != nil {
		return err
	}
	stream, err := c.Upload(noCtx)
	if err != nil {
		return err
	}
	fs, err := os.Open(path)
	if err != nil {
		return err
	}
	name := filepath.Base(path)
	req := &pb.UploadReq{
		Name: name,
		Data: nil,
	}
	bytes := make([]byte, 1024*512)
	var read int
	for err == nil {
		if read, err = fs.Read(bytes); err == nil {
			req.Data = bytes[:read]
			err = stream.Send(req)
		}
	}
	_, err = stream.CloseAndRecv()
	if err == io.EOF {
		err = nil
	}
	return err
}

func Exec(s ServerEntity, cmd string) (string, error) {
	c, err := newClient(s.Addr)
	if err != nil {
		return "", err
	}
	rsp, err := c.Exec(noCtx, &pb.ExecReq{Cmd: cmd})
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

func Rollback(s ServerEntity, version string) (*pb.RollbackRsp, error) {
	c, err := newClient(s.Addr)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Rollback(noCtx, &pb.RollbackReq{Version: version})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func Signal(s ServerEntity, signal int32) error {
	c, err := newClient(s.Addr)
	if err != nil {
		return err
	}
	_, err = c.Signal(noCtx, &pb.SignalReq{Signal: signal})
	return err
}

func Command(s ServerEntity, cmd string, args []string) (string, error) {
	c, err := newClient(s.Addr)
	if err != nil {
		return "", err
	}
	rsp, err := c.Command(noCtx, &pb.CommandReq{Cmd: cmd, Args: args})
	if err != nil {
		return "", err
	}
	return rsp.Msg, nil
}

func newClient(target string) (pb.UpgradeServiceClient, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewUpgradeServiceClient(conn), nil
}
