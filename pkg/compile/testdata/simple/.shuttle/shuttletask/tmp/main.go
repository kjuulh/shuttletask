package main

import (
  "github.com/kjuulh/shuttletask/pkg/cmder"
)

func main() {
  rootcmd := cmder.NewRoot()
  buildcmd := cmder.NewCmd("build", Build)
  //buildcmd := cmder.WithArgs(buildcmd, "something")

  rootcmd.AddCmds(
    buildcmd,
  )

  rootcmd.Execute()
}