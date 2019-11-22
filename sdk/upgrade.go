package sdk

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/king526/goupdater/sdk/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	us = &upgradeService{
		dir:    "files",
		cmdMap: map[string]func([]string) (string, error){},
	}
)

type upgradeService struct {
	dir    string
	cmdMap map[string]func(args []string) (string, error)
	l      sync.RWMutex
}

func (u *upgradeService) RegisterCommand(cmd string, f func([]string) (string, error)) {
	u.l.Lock()
	u.cmdMap[cmd] = f
	u.l.Unlock()
}

func (u *upgradeService) Command(ctx context.Context, req *pb.CommandReq) (*pb.CommandRsp, error) {
	u.l.RLock()
	f, ok := u.cmdMap[req.Cmd]
	if !ok {
		return nil, status.Error(404, "command not found")
	}
	u.l.RUnlock()
	ret, err := f(req.Args)
	if err != nil {
		return nil, err
	}
	return &pb.CommandRsp{
		Msg: ret,
	}, nil
}

func (u *upgradeService) Signal(ctx context.Context, req *pb.SignalReq) (*pb.Null, error) {
	go func() {
		time.Sleep(time.Millisecond * 100)
		err := Kill(os.Getpid(), syscall.Signal(req.Signal))
		if err != nil {
			fmt.Println(err)
		}
	}()
	return &pb.Null{}, nil
}

func (u *upgradeService) Rollback(ctx context.Context, req *pb.RollbackReq) (*pb.RollbackRsp, error) {
	rsp := &pb.RollbackRsp{}
	if req.Version == "" {
		filepath.Walk(u.dir, func(path string, info os.FileInfo, err error) error {
			if !strings.HasPrefix(info.Name(), "__update__.") {
				return nil
			}
			tag := strings.TrimLeft(info.Name(), "__update__.")
			rsp.Version = append(rsp.Version, tag)
			return nil
		})
		return rsp, nil
	}
	fileName := filepath.Join(u.dir, "__update__."+req.Version)
	if _, err := os.Stat(fileName); err != nil {
		return nil, err
	}
	if err := os.Remove(os.Args[0]); err != nil {
		return nil, err
	}
	return rsp, os.Symlink(fileName, os.Args[0])
}

func (u *upgradeService) Exec(ctx context.Context, req *pb.ExecReq) (*pb.ExecRsp, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh")
	cmd.Stdin = strings.NewReader(req.Cmd)
	ret, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return &pb.ExecRsp{
		Data: string(ret),
	}, nil
}

func (u *upgradeService) Upload(us pb.UpgradeService_UploadServer) (err error) {
	var (
		req *pb.UploadReq
		fs  *os.File
	)
	for {
		req, err = us.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if fs == nil {
			fileName := filepath.Join(u.dir, req.Name)
			fs, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				break
			}
		}
		if _, err = fs.Write(req.Data); err != nil {
			break
		}
	}
	return
}

func (u *upgradeService) Update(us pb.UpgradeService_UpdateServer) (err error) {
	var (
		fileName string
		req      *pb.UpdateReq
		fs       *os.File
	)
	for {
		req, err = us.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if fs == nil {
			fileName = "__update__." + req.Tag
			fileName = filepath.Join(u.dir, fileName)
			fs, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
			if err != nil {
				break
			}
		}
		if _, err = fs.Write(req.Data); err != nil {
			break
		}
	}
	if err != nil {
		return
	}
	if fs != nil {
		fs.Close()
	} else {
		return status.Error(400, "no data received")
	}
	if err := os.Remove(os.Args[0]); err != nil {
		return err
	}
	return os.Symlink(fileName, os.Args[0])
}

func (u *upgradeService) Version(context.Context, *pb.Null) (*pb.VersionRsp, error) {
	return &pb.VersionRsp{
		Version: Version,
		Commit:  Commit,
		ModTime: modTime,
		Branch:  Branch,
	}, nil
}

func RegisterUpgradeService(s *grpc.Server) {
	pb.RegisterUpgradeServiceServer(s, us)
}

func RegisterCommand(cmd string, f func(args []string) (string, error)) {
	us.RegisterCommand(cmd, f)
}
