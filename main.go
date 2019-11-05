package main

import (
	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/session"

	"github.com/bserdar/watermelon/cmd"

	_ "github.com/bserdar/watermelon/server/backends/localhost"
	_ "github.com/bserdar/watermelon/server/backends/remotelinux"
)

func main() {

	server.SessionFactory = session.Factory

	server.Localhost.Backend = server.GetBackend("localhost", server.Localhost)
	cmd.Execute()
}
