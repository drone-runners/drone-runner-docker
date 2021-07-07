// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/earthly/earthly/earthfile2llb"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	//command.Command()
	targets, _ := earthfile2llb.GetTargets("Earthfile")
	fmt.Print(targets)
}
