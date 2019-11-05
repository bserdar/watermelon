package server

import (
	"fmt"

	"github.com/bserdar/watermelon/server/pb"
)

// CmdErr is used for command errors. Regular go errors are used for
// fatal infrastructure errors
type CmdErr interface {
	Error() string
	Host() string
	Msg() string
	ToPb() *pb.CommandError
}

type cmdErr struct {
	host string
	msg  string
}

// Error returns an error string
func (c cmdErr) Error() string {
	return fmt.Sprintf("%s: %s", c.host, c.msg)
}

// Host returns the host
func (c cmdErr) Host() string { return c.host }

// Msg returns the error msg
func (c cmdErr) Msg() string { return c.msg }

// ToPb returns grpc object
func (c cmdErr) ToPb() *pb.CommandError {
	return &pb.CommandError{Host: c.host, Msg: c.msg}
}

// NewCmdErr constructs a new CmdErr
func NewCmdErr(host *Host, msg string, args ...interface{}) CmdErr {
	return cmdErr{host: host.ID, msg: fmt.Sprintf(msg, args...)}
}

// CmdErrFromErr returns a CmdErr from error. if err is nil, returns nil
func CmdErrFromErr(host *Host, err error) CmdErr {
	if err == nil {
		return nil
	}
	return cmdErr{host: host.ID, msg: err.Error()}
}

// // Error for command error
// func (e *pb.CommandError) Error() string {
// 	return fmt.Sprintf("%s: %s", e.GetHost(), e.GetMsg())
// }
