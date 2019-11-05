package remotelinux

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/bserdar/watermelon/server"
	scp "github.com/hnakamur/go-scp"
)

// becomeSudo runs the command in sudo
func becomeSudo(in string) string {
	delim := uuid.New().String()
	return fmt.Sprintf(`sudo -s --<<%s
%s
%s
`, delim, in, delim)
}

func scpBecomeSudo(in *scp.SCP) *scp.SCP {
	in.SCPCommand = "sudo scp "
	return in
}

// Become rewrites the command to become another user
func Become(host *server.Host, in string) string {
	if host.Become == "sudo" {
		return becomeSudo(in)
	}
	return in
}

func BecomeSCP(host *server.Host, in *scp.SCP) *scp.SCP {
	if host.Become == "sudo" {
		return scpBecomeSudo(in)
	}
	return in
}
