package shell

import (
	"github.com/abiosoft/ishell"
)

func RunShell() {
	shell := ishell.New()
	shell.Println("Blacker Interactive Shell 0.1")

	shell.AddCmd(&ishell.Cmd{
		Name: "cleanup-emails",
		Help: "Find email duplicates & allow to clean them up.",
		Func: CleanupDuplicatedEmails,
	})

	// start shell
	shell.Start()
}
