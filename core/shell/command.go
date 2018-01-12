package shell

import "gopkg.in/abiosoft/ishell.v2"

func RunShell() {
	shell := ishell.New()

	shell.AddCmd(&ishell.Cmd{
		Name: "cleanup-emails",
		Help: "Find email duplicates & allow to clean them up.",
		Func: CleanupDuplicatedEmails,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "test-events",
		Help: "Test events abstraction.",
		Func: TestEventHandler,
	})

	// start shell
	shell.Start()
}
